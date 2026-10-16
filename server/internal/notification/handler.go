package notification

import (
	"github.com/TogetherBros/azit/proto"
	"github.com/TogetherBros/azit/server/internal/auth"
	"github.com/gofiber/fiber/v3"
)

type Handler struct {
	repo *Repository
}

func NewHandler(repo *Repository) *Handler {
	return &Handler{repo: repo}
}

// Routes 는 /v1/notifications 아래 경로를 붙인다. r 에는 인증 미들웨어가 걸려 있어야 한다.
func (h *Handler) Routes(r fiber.Router) {
	r.Get("/", h.List)
	r.Post("/read-all", h.ReadAll)
}

// List 는 내 알림 최신 50개와 안 읽은 수. GET /v1/notifications
// 앱이 몇 초마다 불러서 종 숫자를 갱신하고, 새 흔들기 알림이 오면 창을 흔든다.
func (h *Handler) List(c fiber.Ctx) error {
	ctx := c.Context()
	me := auth.AccountID(c)

	rows, err := h.repo.List(ctx, me, ListLimit)
	if err != nil {
		return err
	}
	unread, err := h.repo.CountUnread(ctx, me)
	if err != nil {
		return err
	}

	res := proto.NotificationListResponse{Notifications: make([]proto.NotificationResponse, 0, len(rows)), Unread: unread}
	for _, n := range rows {
		res.Notifications = append(res.Notifications, proto.NotificationResponse{
			NotificationID: n.NotificationID,
			Kind:           string(n.Kind),
			AzitID:         n.AzitID,
			AzitName:       n.AzitName,
			RoomID:         n.RoomID,
			RoomName:       n.RoomName,
			MessageID:      n.MessageID,
			ActorID:        n.ActorID,
			ActorNickname:  n.ActorNick,
			Preview:        n.Preview,
			Read:           n.IsRead,
			CreatedAt:      n.CreatedAt,
		})
	}
	return c.JSON(res)
}

// ReadAll 은 내 알림을 전부 읽음으로. POST /v1/notifications/read-all
func (h *Handler) ReadAll(c fiber.Ctx) error {
	if err := h.repo.ReadAll(c.Context(), auth.AccountID(c)); err != nil {
		return err
	}
	return c.SendStatus(fiber.StatusNoContent)
}
