// Package logapi 는 아지트 변경 기록을 보여 주고 되돌리는 HTTP 경로다.
// changelog 는 member 를 모르고(member 가 changelog 로 기록을 남기니까), 여기서 둘을 같이 쓴다.
package logapi

import (
	"errors"

	"github.com/TogetherBros/azit/proto"
	"github.com/TogetherBros/azit/server/internal/changelog"
	"github.com/TogetherBros/azit/server/internal/httperr"
	"github.com/TogetherBros/azit/server/internal/member"
	"github.com/gofiber/fiber/v3"
)

// ListLimit 은 한 번에 보여 주는 최근 기록 수.
const ListLimit = 100

type Handler struct {
	logs *changelog.Repository
}

func NewHandler(logs *changelog.Repository) *Handler {
	return &Handler{logs: logs}
}

// Routes 는 /v1/azits 그룹 아래에 변경 기록 경로를 단다.
func (h *Handler) Routes(r fiber.Router, requireMember fiber.Handler) {
	r.Get("/:id/log", requireMember, h.List)
	r.Post("/:id/log/:entryId/undo", requireMember, h.Undo)
}

// List 는 최근 변경 기록. 멤버면 누구나. GET /v1/azits/{id}/log
func (h *Handler) List(c fiber.Ctx) error {
	me := member.Current(c)
	entries, err := h.logs.List(c.Context(), me.AzitID, ListLimit)
	if err != nil {
		return err
	}
	res := make([]proto.ChangeLogResponse, 0, len(entries))
	for _, e := range entries {
		res = append(res, h.toResponse(e))
	}
	return c.JSON(res)
}

// Undo 는 기록 하나를 되돌린다. 관리자 이상. POST /v1/azits/{id}/log/{entryId}/undo
// 되돌린 기록 줄을 돌려준다.
func (h *Handler) Undo(c fiber.Ctx) error {
	me := member.Current(c)
	if !me.Role.AtLeast(member.RoleAdmin) {
		return httperr.New(fiber.StatusForbidden, proto.ErrCodeForbidden, "관리자만 되돌릴 수 있어요")
	}
	e, err := h.logs.Undo(c.Context(), me.AzitID, fiber.Params[int64](c, "entryId"), me.AccountID, me.NickName)
	switch {
	case errors.Is(err, changelog.ErrNotFound):
		return httperr.New(fiber.StatusNotFound, proto.ErrCodeNotFound, err.Error())
	case errors.Is(err, changelog.ErrAlreadyUndone), errors.Is(err, changelog.ErrNotUndoable):
		return httperr.New(fiber.StatusBadRequest, proto.ErrCodeBadRequest, err.Error())
	case err != nil:
		var he *httperr.Error
		if errors.As(err, &he) {
			return err
		}
		return httperr.New(fiber.StatusBadRequest, proto.ErrCodeBadRequest, "되돌리지 못했어요: "+err.Error())
	}
	return c.JSON(h.toResponse(e))
}

func (h *Handler) toResponse(e changelog.Entry) proto.ChangeLogResponse {
	return proto.ChangeLogResponse{
		EntryID:       e.EntryID,
		Kind:          e.Kind,
		ActorNickname: e.ActorNick,
		Summary:       e.Summary,
		CreatedAt:     e.CreatedAt,
		Undoable:      e.UndoneAt == nil && h.logs.Undoable(e.Kind),
		UndoneAt:      e.UndoneAt,
		UndoneBy:      e.UndoneByNick,
	}
}
