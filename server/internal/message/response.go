package message

import (
	"context"

	"github.com/TogetherBros/azit/proto"
	"github.com/TogetherBros/azit/server/internal/member"
	"github.com/gofiber/fiber/v3"
)

// respondOne 은 메시지 하나를 MessageResponse 로 응답한다.
func (h *Handler) respondOne(c fiber.Ctx, status int, me member.Member, m Message) error {
	res, err := h.toResponses(c.Context(), me, []Message{m})
	if err != nil {
		return err
	}
	return c.Status(status).JSON(res[0])
}

// toResponses 는 메시지 여러 개를 응답으로 바꾼다. 순서는 받은 그대로.
//
// 메시지마다 보낸 사람, @, 맞장구를 따로 조회하면 50개 목록에 쿼리가 수백 번 나간다 (N+1 문제).
// 그래서 id 를 먼저 다 모은 다음, 종류별로 IN (...) 쿼리 한 번씩만 날리고 map 으로 붙인다.
// JPA 로 치면 fetch join / @BatchSize 를 손으로 하는 것.
func (h *Handler) toResponses(ctx context.Context, me member.Member, messages []Message) ([]proto.MessageResponse, error) {
	// 1. id 모으기
	messageIDs := make([]int64, 0, len(messages))
	var senderIDs []int64
	for _, m := range messages {
		messageIDs = append(messageIDs, m.MessageID)
		if m.SenderID != 0 { // system 메시지는 보낸 사람이 없다
			senderIDs = append(senderIDs, m.SenderID)
		}
	}
	senderIDs = UniqueIDs(senderIDs)

	// 2. 종류별로 한 번씩 조회
	shakes, err := h.repo.ShakeIDsByMessages(ctx, messageIDs)
	if err != nil {
		return nil, err
	}
	reactions, err := h.repo.ReactionsByMessages(ctx, messageIDs)
	if err != nil {
		return nil, err
	}
	memeURLs, err := h.memeURLs(ctx, reactions)
	if err != nil {
		return nil, err
	}
	senders, err := h.senders(ctx, me.AzitID, senderIDs)
	if err != nil {
		return nil, err
	}

	// 3. 메시지마다 붙이기
	res := make([]proto.MessageResponse, 0, len(messages))
	for _, m := range messages {
		r := toResponse(m, shakes[m.MessageID], senders[m.SenderID],
			groupReactions(reactions[m.MessageID], me.AccountID, memeURLs))
		if r.MediaURL, err = h.mediaURL(ctx, m); err != nil {
			return nil, err
		}
		res = append(res, r)
	}
	return res, nil
}

// senders 는 보낸 사람 정보를 account id → SenderResponse 로 만든다.
// 닉네임은 아지트 안 닉네임(Member), 색과 사진은 계정(Account) 것.
func (h *Handler) senders(ctx context.Context, azitID int64, accountIDs []int64) (map[int64]*proto.SenderResponse, error) {
	members, err := h.memberRepo.FindByAccountIDs(ctx, azitID, accountIDs)
	if err != nil {
		return nil, err
	}
	accounts, err := h.accountRepo.FindByIDs(ctx, accountIDs)
	if err != nil {
		return nil, err
	}

	out := make(map[int64]*proto.SenderResponse, len(accountIDs))
	for _, id := range accountIDs {
		acc := accounts[id] // 없으면 빈 값 (탈퇴 등). 그래도 id 는 채운다
		avatarURL := ""
		if acc.AvatarKey != "" {
			avatarURL, err = h.r2.PresignGet(ctx, acc.AvatarKey)
			if err != nil {
				return nil, err
			}
		}
		out[id] = &proto.SenderResponse{
			AccountID: id,
			Nickname:  members[id].NickName,
			Color:     acc.Color,
			AvatarURL: avatarURL,
		}
	}
	return out, nil
}

// memeURLs 는 맞장구 줄들에 나온 짤의 보기 링크를 짤 id → 링크 로 만든다.
func (h *Handler) memeURLs(ctx context.Context, reactions map[int64][]Reaction) (map[int64]string, error) {
	var ids []int64
	for _, rows := range reactions {
		for _, rc := range rows {
			if rc.MemeID > 0 {
				ids = append(ids, rc.MemeID)
			}
		}
	}
	memes, err := h.memeRepo.FindByIDs(ctx, UniqueIDs(ids))
	if err != nil {
		return nil, err
	}

	out := make(map[int64]string, len(memes))
	for id, m := range memes {
		url, err := h.r2.PresignGet(ctx, m.MediaKey)
		if err != nil {
			return nil, err
		}
		out[id] = url
	}
	return out, nil
}

// mediaURL 은 사진·짤 메시지의 그림 보기 링크. 글자 메시지나 지운 메시지면 "".
// 링크 만들기는 서명만 하고 네트워크를 안 타서 메시지마다 불러도 된다.
func (h *Handler) mediaURL(ctx context.Context, m Message) (string, error) {
	if m.DeletedAt != nil || (m.Kind != KindPhoto && m.Kind != KindMeme) || m.Body == "" {
		return "", nil
	}
	return h.r2.PresignGet(ctx, m.Body)
}

// groupReactions 는 맞장구 줄들(한 사람 한 줄)을 종류별로 묶는다: "ㅋㅋ" 줄 3개 → {ㅋㅋ, count 3}.
// 순서는 그 종류가 처음 달린 순서. 내가 단 종류는 Mine = true.
func groupReactions(rows []Reaction, meID int64, memeURLs map[int64]string) []proto.ReactionResponse {
	type key struct {
		text   string
		memeID int64
	}
	index := map[key]int{} // 종류 → res 안의 위치
	res := []proto.ReactionResponse{}

	for _, rc := range rows {
		k := key{rc.Text, rc.MemeID}
		i, ok := index[k]
		if !ok {
			i = len(res)
			index[k] = i
			res = append(res, proto.ReactionResponse{Text: rc.Text, MemeID: rc.MemeID, MemeURL: memeURLs[rc.MemeID]})
		}
		res[i].Count++
		if rc.AccountID == meID {
			res[i].Mine = true
		}
	}
	return res
}

// toResponse 는 모델을 응답 DTO 로 옮긴다. 지운 메시지면 본문, @, 맞장구를 비운다.
func toResponse(m Message, shakeIDs []int64, sender *proto.SenderResponse, reactions []proto.ReactionResponse) proto.MessageResponse {
	deleted := m.DeletedAt != nil
	body := m.Body
	if m.Kind == KindPhoto || m.Kind == KindMeme {
		body = "" // 본문엔 R2 키가 들어 있다. 앱엔 media_url 로 준다
	}
	if deleted {
		body = ""
		shakeIDs = nil
		reactions = nil
	}
	if shakeIDs == nil {
		shakeIDs = []int64{} // JSON 에서 null 대신 [] 로 나가게
	}
	if reactions == nil {
		reactions = []proto.ReactionResponse{}
	}

	return proto.MessageResponse{
		MessageID: m.MessageID,
		RoomID:    m.RoomID,
		TempID:    m.TempID,
		Kind:      string(m.Kind),
		Body:      body,
		MemeID:    m.MemeID,
		ReplyToID: m.ReplyToID,
		ShakeIDs:  shakeIDs,
		ShakeAll:  m.ShakeAll && !deleted,
		CreatedAt: m.CreatedAt,
		EditedAt:  m.EditedAt,
		Deleted:   deleted,
		Sender:    sender,
		Reactions: reactions,
	}
}
