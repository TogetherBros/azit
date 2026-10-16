package message

import (
	"errors"

	"github.com/TogetherBros/azit/proto"
	"github.com/TogetherBros/azit/server/internal/httperr"
	"github.com/TogetherBros/azit/server/internal/member"
	"github.com/TogetherBros/azit/server/internal/meme"
	"github.com/gofiber/fiber/v3"
)

// AddReaction 은 메시지에 맞장구를 단다. 멤버면 누구나. PUT /v1/azits/{id}/messages/{messageId}/reactions
// 글자는 아지트 맞장구 목록에 있는 것만, 짤은 이 아지트 짤 모음에 있는 것만 (목록에 없어도 됨). 이미 달았으면 그대로 (두 번 눌러도 같은 결과).
// 응답은 그 메시지의 맞장구 묶음 전체 (앱이 그대로 다시 그린다).
func (h *Handler) AddReaction(c fiber.Ctx) error {
	ctx := c.Context()
	me := member.Current(c)

	m, text, memeID, err := h.reactionTarget(c, me)
	if err != nil {
		return err
	}

	// 글자 맞장구면 아지트 맞장구 목록에 있는 글자인지
	if text != "" {
		ok, err := h.memeRepo.HasReactText(ctx, me.AzitID, text)
		if err != nil {
			return err
		}
		if !ok {
			return badRequest(errors.New("아지트 맞장구 목록에 있는 글자만 달 수 있어요"))
		}
	}

	// 짤 맞장구면 우리 아지트의 지우지 않은 짤인지
	if memeID > 0 {
		if _, err := h.memeRepo.FindActive(ctx, me.AzitID, memeID); err != nil {
			if errors.Is(err, meme.ErrNotFound) {
				return badRequest(err)
			}
			return err
		}
	}

	err = h.repo.AddReaction(ctx, Reaction{MessageID: m.MessageID, AccountID: me.AccountID, Text: text, MemeID: memeID})
	if errors.Is(err, ErrReactionLimit) {
		return httperr.New(fiber.StatusConflict, proto.ErrCodeReactLimit, err.Error())
	}
	if err != nil {
		return err
	}

	return h.respondReactions(c, me, m.MessageID)
}

// RemoveReaction 은 내가 단 맞장구를 뗀다. DELETE /v1/azits/{id}/messages/{messageId}/reactions (본문은 AddReaction 과 같다)
// 안 달려 있었어도 성공. 응답은 그 메시지의 맞장구 묶음 전체.
func (h *Handler) RemoveReaction(c fiber.Ctx) error {
	me := member.Current(c)

	m, text, memeID, err := h.reactionTarget(c, me)
	if err != nil {
		return err
	}

	if err := h.repo.RemoveReaction(c.Context(), m.MessageID, me.AccountID, text, memeID); err != nil {
		return err
	}
	return h.respondReactions(c, me, m.MessageID)
}

// reactionTarget 은 달기/떼기 공통: 메시지를 찾고 요청 본문을 읽어 검사한다.
func (h *Handler) reactionTarget(c fiber.Ctx, me member.Member) (Message, string, int64, error) {
	m, err := h.messageOf(c, me)
	if err != nil {
		return Message{}, "", 0, err
	}
	if m.DeletedAt != nil {
		return Message{}, "", 0, badRequest(errors.New("지운 메시지예요"))
	}

	var req proto.ReactionRequest
	if err := c.Bind().JSON(&req); err != nil {
		return Message{}, "", 0, fiber.ErrBadRequest
	}
	text := meme.NormalizeReactText(req.Text)
	if err := meme.ValidateReact(text, req.MemeID); err != nil {
		return Message{}, "", 0, badRequest(err)
	}
	return m, text, req.MemeID, nil
}

// respondReactions 는 메시지 하나의 맞장구 묶음을 다시 읽어 응답한다.
func (h *Handler) respondReactions(c fiber.Ctx, me member.Member, messageID int64) error {
	rows, err := h.repo.ReactionsByMessages(c.Context(), []int64{messageID})
	if err != nil {
		return err
	}
	memeURLs, err := h.memeURLs(c.Context(), rows)
	if err != nil {
		return err
	}
	return c.JSON(groupReactions(rows[messageID], me.AccountID, memeURLs))
}
