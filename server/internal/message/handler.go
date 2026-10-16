package message

import (
	"errors"
	"time"

	"github.com/TogetherBros/azit/proto"
	"github.com/TogetherBros/azit/server/internal/account"
	"github.com/TogetherBros/azit/server/internal/httperr"
	"github.com/TogetherBros/azit/server/internal/member"
	"github.com/TogetherBros/azit/server/internal/meme"
	"github.com/TogetherBros/azit/server/internal/room"
	"github.com/TogetherBros/azit/server/internal/storage"
	"github.com/gofiber/fiber/v3"
)

// 목록 한 번에 가져오는 개수. limit 을 안 주면 DefaultLimit, 너무 크면 MaxLimit 으로 줄인다.
const (
	DefaultLimit = 50
	MaxLimit     = 100
)

type Handler struct {
	repo        *Repository
	roomRepo    *room.Repository    // 방이 이 아지트 것인지 확인
	memberRepo  *member.Repository  // @로 부른 사람 확인, 보낸 사람 닉네임
	accountRepo *account.Repository // 보낸 사람 색, 사진
	memeRepo    *meme.Repository    // 짤 맞장구 확인, 짤 보기 링크
	r2          *storage.R2         // 사진 키 → 보기 주소
}

func NewHandler(repo *Repository, roomRepo *room.Repository, memberRepo *member.Repository, accountRepo *account.Repository, memeRepo *meme.Repository, r2 *storage.R2) *Handler {
	return &Handler{repo: repo, roomRepo: roomRepo, memberRepo: memberRepo, accountRepo: accountRepo, memeRepo: memeRepo, r2: r2}
}

// Routes 는 /v1/azits 그룹 아래에 메시지 경로를 단다. room.Handler.Routes 와 같은 방식.
func (h *Handler) Routes(r fiber.Router, requireMember fiber.Handler) {
	// 방 안에서 하는 것: 보내기, 목록, 읽음
	r.Post("/:id/rooms/:roomId/messages", requireMember, h.SendMessage)
	r.Get("/:id/rooms/:roomId/messages", requireMember, h.ListMessages)
	r.Post("/:id/rooms/:roomId/read", requireMember, h.MarkRead)

	// 메시지 하나에 하는 것: 고치기, 지우기, 맞장구
	r.Patch("/:id/messages/:messageId", requireMember, h.UpdateMessage)
	r.Delete("/:id/messages/:messageId", requireMember, h.DeleteMessage)
	r.Put("/:id/messages/:messageId/reactions", requireMember, h.AddReaction)
	r.Delete("/:id/messages/:messageId/reactions", requireMember, h.RemoveReaction)
}

// SendMessage 는 메시지를 보낸다. 멤버면 누구나. POST /v1/azits/{id}/rooms/{roomId}/messages
//
// 같은 번호표(temp_id)로 다시 오면 (네트워크가 끊겨 앱이 재전송한 경우) 새로 만들지 않고
// 처음 저장한 메시지를 그대로 돌려준다. 번호표는 같은데 본문이 다르면 409.
func (h *Handler) SendMessage(c fiber.Ctx) error {
	ctx := c.Context()

	// 1. 누가 보냈는지. requireMember 가 이미 "이 아지트 멤버" 임을 확인하고 넣어 둔 값이다.
	//    그래서 아지트 id 도 경로에서 다시 꺼낼 필요 없이 me.AzitID 를 쓴다.
	me := member.Current(c)

	// 2. 방이 이 아지트 것인지. 남의 아지트 방 id 를 넣어서 글을 쓰는 걸 막는다.
	rm, err := h.roomOf(c, me)
	if err != nil {
		return err
	}

	// 3. 요청 읽기 + 검사
	var req proto.SendMessageRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.ErrBadRequest
	}
	if err := ValidateTempID(req.TempID); err != nil {
		return badRequest(err)
	}
	if err := ValidateBody(req.Body); err != nil {
		return badRequest(err)
	}

	// 4. 재전송인지 먼저 본다. 검사를 다 통과한 다음에 보는 이유:
	//    처음 보낼 때도 같은 검사를 통과했을 테니, 재전송이면 여기까지 똑같이 온다.
	old, err := h.repo.FindByTempID(ctx, me.AccountID, req.TempID)
	if err == nil {
		if old.Body != req.Body {
			return httperr.New(fiber.StatusConflict, proto.ErrCodeTempIDConflict, "같은 temp_id 로 다른 메시지가 이미 저장돼 있어요")
		}
		return h.respondOne(c, fiber.StatusCreated, me, old)
	}
	if !errors.Is(err, ErrNotFound) {
		return err // DB 에러
	}

	// 5. 답장이면 원래 메시지가 같은 방에 있는지. 다른 방 메시지에 답장을 다는 걸 막는다.
	if req.ReplyToID != nil {
		orig, err := h.repo.FindByID(ctx, *req.ReplyToID)
		if errors.Is(err, ErrNotFound) || (err == nil && orig.RoomID != rm.RoomID) {
			return badRequest(errors.New("답장할 메시지가 이 방에 없어요"))
		}
		if err != nil {
			return err
		}
	}

	// 6. @로 부른 사람들. 중복을 빼고, 전부 지금 이 아지트 멤버인지 확인한다.
	//    남의 아지트 사람 id 를 넣어서 흔드는 걸 막는다.
	shakeIDs := UniqueIDs(req.ShakeIDs)
	if len(shakeIDs) > 0 {
		n, err := h.memberRepo.CountActiveAmong(ctx, me.AzitID, shakeIDs)
		if err != nil {
			return err
		}
		if n != int64(len(shakeIDs)) {
			return badRequest(errors.New("이 아지트에 없는 사람을 불렀어요"))
		}
	}

	// 7. 저장. 메시지 + Shake 들이 한 트랜잭션으로 들어간다 (repository.Create).
	m := Message{
		AzitID:    me.AzitID,
		RoomID:    rm.RoomID,
		SenderID:  me.AccountID,
		TempID:    req.TempID,
		Kind:      KindText, // 사진/짤 메시지는 나중에. 지금은 글자만
		Body:      req.Body,
		ReplyToID: req.ReplyToID,
		ShakeAll:  req.ShakeAll,
	}
	if err := h.repo.Create(ctx, &m, shakeIDs); err != nil {
		return err
	}

	// 8. 내가 보낸 메시지까지는 읽은 걸로. 안 그러면 "여기까지 읽었음" 선이 내 메시지 위에 생긴다.
	if err := h.repo.MarkRead(ctx, me.AccountID, rm.RoomID, m.MessageID); err != nil {
		return err
	}

	return h.respondOne(c, fiber.StatusCreated, me, m)
}

// ListMessages 는 방 메시지 목록. 멤버면 누구나. GET /v1/azits/{id}/rooms/{roomId}/messages?before=&limit=
// 처음엔 before 없이 부르면 최신 limit 개. 위로 스크롤하면 지금 맨 위 메시지 id 를 before 로 넘긴다.
func (h *Handler) ListMessages(c fiber.Ctx) error {
	ctx := c.Context()
	me := member.Current(c)

	rm, err := h.roomOf(c, me)
	if err != nil {
		return err
	}

	before := fiber.Query[int64](c, "before") // 없으면 0 = 맨 최신부터
	limit := fiber.Query[int](c, "limit")
	if limit <= 0 {
		limit = DefaultLimit
	}
	if limit > MaxLimit {
		limit = MaxLimit
	}

	// limit 보다 하나 더 가져와 본다. 하나 더 나오면 위에 더 남아 있다는 뜻 (COUNT 쿼리 없이 has_more 를 안다).
	messages, err := h.repo.ListBefore(ctx, rm.RoomID, before, limit+1)
	if err != nil {
		return err
	}
	hasMore := len(messages) > limit
	if hasMore {
		messages = messages[:limit]
	}

	// DB 에서는 최신 → 오래된 순으로 가져왔다 (최신 limit 개를 자르려고). 앱은 위→아래로 그리니까 뒤집는다.
	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}

	res, err := h.toResponses(ctx, me, messages)
	if err != nil {
		return err
	}

	lastRead, err := h.repo.LastRead(ctx, me.AccountID, rm.RoomID)
	if err != nil {
		return err
	}

	return c.JSON(proto.MessageListResponse{Messages: res, HasMore: hasMore, LastReadID: lastRead})
}

// MarkRead 는 "이 방을 여기까지 읽었다" 를 저장한다. 멤버면 누구나. POST /v1/azits/{id}/rooms/{roomId}/read
// 앱은 방을 보고 있는 동안, 화면에 보인 가장 아래 메시지 id 로 부른다. 더 앞 id 가 오면 무시한다.
func (h *Handler) MarkRead(c fiber.Ctx) error {
	ctx := c.Context()
	me := member.Current(c)

	rm, err := h.roomOf(c, me)
	if err != nil {
		return err
	}

	var req proto.MarkReadRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.ErrBadRequest
	}

	// 이 방 메시지인지. 다른 방 메시지 id 나 아직 없는 큰 id 로 "다 읽음" 처리하는 걸 막는다.
	m, err := h.repo.FindByID(ctx, req.MessageID)
	if errors.Is(err, ErrNotFound) || (err == nil && m.RoomID != rm.RoomID) {
		return badRequest(errors.New("이 방 메시지가 아니에요"))
	}
	if err != nil {
		return err
	}

	if err := h.repo.MarkRead(ctx, me.AccountID, rm.RoomID, m.MessageID); err != nil {
		return err
	}
	return c.SendStatus(fiber.StatusNoContent)
}

// UpdateMessage 는 메시지 본문을 고친다. 보낸 본인만, 글자 메시지만. PATCH /v1/azits/{id}/messages/{messageId}
// @로 부른 사람과 답장 대상은 그대로 둔다 (본문만 고친다).
func (h *Handler) UpdateMessage(c fiber.Ctx) error {
	ctx := c.Context()
	me := member.Current(c)

	m, err := h.messageOf(c, me)
	if err != nil {
		return err
	}
	if m.DeletedAt != nil {
		return badRequest(errors.New("지운 메시지는 고칠 수 없어요"))
	}
	if m.SenderID != me.AccountID {
		return httperr.New(fiber.StatusForbidden, proto.ErrCodeForbidden, "내가 보낸 메시지만 고칠 수 있어요")
	}
	if m.Kind != KindText {
		return badRequest(errors.New("글자 메시지만 고칠 수 있어요"))
	}

	var req proto.UpdateMessageRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.ErrBadRequest
	}
	if err := ValidateBody(req.Body); err != nil {
		return badRequest(err)
	}

	// 바뀐 게 없으면 저장하지 않는다 ("수정됨" 표시가 괜히 붙지 않게).
	if req.Body != m.Body {
		now := time.Now()
		if err := h.repo.UpdateBody(ctx, m.MessageID, req.Body, now); err != nil {
			return err
		}
		m.Body = req.Body
		m.EditedAt = &now
	}

	return h.respondOne(c, fiber.StatusOK, me, m)
}

// DeleteMessage 는 메시지를 지운다. 보낸 본인이나 관리자 이상. DELETE /v1/azits/{id}/messages/{messageId}
// 줄은 남기고 "삭제된 메시지" 로 보인다. 이미 지운 거면 그냥 204 (두 번 눌러도 같은 결과).
func (h *Handler) DeleteMessage(c fiber.Ctx) error {
	me := member.Current(c)

	m, err := h.messageOf(c, me)
	if err != nil {
		return err
	}
	if m.SenderID != me.AccountID && !me.Role.AtLeast(member.RoleAdmin) {
		return httperr.New(fiber.StatusForbidden, proto.ErrCodeForbidden, "내가 보낸 메시지만 지울 수 있어요")
	}
	if m.DeletedAt != nil {
		return c.SendStatus(fiber.StatusNoContent)
	}

	if err := h.repo.SoftDelete(c.Context(), m.MessageID, time.Now()); err != nil {
		return err
	}
	return c.SendStatus(fiber.StatusNoContent)
}

// roomOf 는 경로의 roomId 방을 찾는다. 없는 방이든 남의 아지트 방이든 똑같이 404
// (남의 방이 있다는 사실도 안 알려 주려고).
func (h *Handler) roomOf(c fiber.Ctx, me member.Member) (room.Room, error) {
	rm, err := h.roomRepo.FindByID(c.Context(), fiber.Params[int64](c, "roomId"))
	if errors.Is(err, room.ErrNotFound) || (err == nil && rm.AzitID != me.AzitID) {
		return room.Room{}, httperr.New(fiber.StatusNotFound, proto.ErrCodeNotFound, "방이 없어요")
	}
	return rm, err
}

// messageOf 는 경로의 messageId 메시지를 찾는다. 없거나 남의 아지트 메시지면 404.
func (h *Handler) messageOf(c fiber.Ctx, me member.Member) (Message, error) {
	m, err := h.repo.FindByID(c.Context(), fiber.Params[int64](c, "messageId"))
	if errors.Is(err, ErrNotFound) || (err == nil && m.AzitID != me.AzitID) {
		return Message{}, httperr.New(fiber.StatusNotFound, proto.ErrCodeNotFound, "메시지가 없어요")
	}
	return m, err
}

func badRequest(err error) error {
	return httperr.New(fiber.StatusBadRequest, proto.ErrCodeBadRequest, err.Error())
}
