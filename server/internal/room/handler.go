package room

import (
	"context"
	"errors"
	"fmt"

	"github.com/TogetherBros/azit/proto"
	"github.com/TogetherBros/azit/server/internal/changelog"
	"github.com/TogetherBros/azit/server/internal/httperr"
	"github.com/TogetherBros/azit/server/internal/live"
	"github.com/TogetherBros/azit/server/internal/member"
	"github.com/gofiber/fiber/v3"
)

// UnreadCounter 는 방마다 안 읽은 메시지 수를 세는 쪽. message.Repository 가 이걸 만족한다.
//
// room 이 message 를 직접 import 하지 않고 인터페이스로 받는 이유: message 가 이미 room 을 import 한다
// (메시지 보낼 때 방 확인). 서로 import 하면 Go 는 컴파일을 안 해 준다 (순환 import).
// 그래서 room 은 "이런 메서드가 있는 무언가" 만 알고, 실제 값은 wire.go 에서 넣어 준다.
// 스프링으로 치면 인터페이스에 의존하고 구현체는 DI 로 받는 것.
type UnreadCounter interface {
	CountUnread(ctx context.Context, azitID, accountID int64) (map[int64]int64, error)
}

type Handler struct {
	roomRepo *Repository
	unread   UnreadCounter
	live     *live.Live            // 방이 바뀌면 멤버들에게 알림 (nil 이면 안 보냄)
	logs     *changelog.Repository // 변경 기록 (nil 이면 안 남김)
}

func NewHandler(roomRepo *Repository, unread UnreadCounter, lv *live.Live, logs *changelog.Repository) *Handler {
	return &Handler{roomRepo: roomRepo, unread: unread, live: lv, logs: logs}
}

// Routes 는 /v1/azits 그룹 아래에 방 경로를 단다.
// requireMember 는 member.RequireMember 로 만든 미들웨어. 모든 경로 앞에 끼운다.
func (h *Handler) Routes(r fiber.Router, requireMember fiber.Handler) {
	r.Get("/:id/rooms", requireMember, h.List)
	r.Post("/:id/rooms", requireMember, h.Create)
	r.Put("/:id/rooms/order", requireMember, h.Reorder)
	r.Patch("/:id/rooms/:roomId", requireMember, h.Rename)
	r.Delete("/:id/rooms/:roomId", requireMember, h.Delete)
}

// Delete 는 방을 지운다. 관리자 이상만. DELETE /v1/azits/{id}/rooms/{roomId}
// 메시지는 남기고 방만 숨긴다. 변경 기록에서 되돌리면 다시 보인다. 마지막 방은 못 지운다.
func (h *Handler) Delete(c fiber.Ctx) error {
	me := member.Current(c)
	if err := requireAdmin(me); err != nil {
		return err
	}
	roomID := fiber.Params[int64](c, "roomId")
	room, err := h.roomRepo.FindByID(c.Context(), roomID)
	if errors.Is(err, ErrNotFound) || (err == nil && room.AzitID != me.AzitID) {
		return httperr.New(fiber.StatusNotFound, proto.ErrCodeNotFound, ErrNotFound.Error())
	}
	if err != nil {
		return err
	}

	err = h.roomRepo.Delete(c.Context(), me.AzitID, roomID)
	if errors.Is(err, ErrLastRoom) || errors.Is(err, ErrDefaultRoom) {
		return httperr.New(fiber.StatusBadRequest, proto.ErrCodeBadRequest, err.Error())
	}
	if errors.Is(err, ErrNotFound) {
		return httperr.New(fiber.StatusNotFound, proto.ErrCodeNotFound, err.Error())
	}
	if err != nil {
		return err
	}

	h.logs.Record(c.Context(), changelog.Entry{
		AzitID: me.AzitID, ActorID: me.AccountID, ActorNick: me.NickName,
		Kind: changelog.KindRoomDelete, TargetID: roomID, Summary: fmt.Sprintf("'%s' 방을 지움", room.Name),
	}, nil)
	h.live.AzitChanged(c.Context(), me.AzitID, proto.ChangedRooms)
	return c.SendStatus(fiber.StatusNoContent)
}

// List 는 아지트의 방 목록. 멤버면 누구나. GET /v1/azits/{id}/rooms
func (h *Handler) List(c fiber.Ctx) error {
	me := member.Current(c)
	return h.writeList(c, me)
}

// Create 는 방을 맨 아래에 만든다. 관리자 이상만. POST /v1/azits/{id}/rooms
func (h *Handler) Create(c fiber.Ctx) error {
	me := member.Current(c)
	if err := requireAdmin(me); err != nil {
		return err
	}

	var req proto.CreateRoomRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.ErrBadRequest
	}

	name := NormalizeName(req.Name)
	if err := ValidateName(name); err != nil {
		return httperr.New(fiber.StatusBadRequest, proto.ErrCodeBadRequest, err.Error())
	}

	room := Room{
		AzitID:    me.AzitID,
		Name:      name,
		CreatedBy: me.AccountID,
	}
	if err := h.roomRepo.Create(c.Context(), &room); err != nil {
		return err
	}
	h.logs.Record(c.Context(), changelog.Entry{
		AzitID: me.AzitID, ActorID: me.AccountID, ActorNick: me.NickName,
		Kind: changelog.KindRoomCreate, TargetID: room.RoomID, Summary: fmt.Sprintf("'%s' 방을 만듦", room.Name),
	}, nil)
	h.live.AzitChanged(c.Context(), me.AzitID, proto.ChangedRooms)

	return c.Status(fiber.StatusCreated).JSON(toResponse(room))
}

// Rename 은 방 이름을 바꾼다. 관리자 이상만. PATCH /v1/azits/{id}/rooms/{roomId}
func (h *Handler) Rename(c fiber.Ctx) error {
	me := member.Current(c)
	if err := requireAdmin(me); err != nil {
		return err
	}

	roomID := fiber.Params[int64](c, "roomId")
	if roomID <= 0 {
		return fiber.ErrBadRequest
	}

	room, err := h.roomRepo.FindByID(c.Context(), roomID)
	// 다른 아지트 방이면 없는 방처럼 다룬다. 남의 아지트 방을 고치지 못하게.
	if errors.Is(err, ErrNotFound) || (err == nil && room.AzitID != me.AzitID) {
		return httperr.New(fiber.StatusNotFound, proto.ErrCodeNotFound, ErrNotFound.Error())
	}
	if err != nil {
		return err
	}

	var req proto.UpdateRoomRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.ErrBadRequest
	}

	name := NormalizeName(req.Name)
	if err := ValidateName(name); err != nil {
		return httperr.New(fiber.StatusBadRequest, proto.ErrCodeBadRequest, err.Error())
	}

	if err := h.roomRepo.Rename(c.Context(), roomID, name); err != nil {
		return err
	}
	if name != room.Name {
		h.logs.Record(c.Context(), changelog.Entry{
			AzitID: me.AzitID, ActorID: me.AccountID, ActorNick: me.NickName,
			Kind: changelog.KindRoomRename, TargetID: roomID, Summary: fmt.Sprintf("'%s' 방 이름을 '%s'(으)로", room.Name, name),
		}, map[string]string{"name": room.Name})
		h.live.AzitChanged(c.Context(), me.AzitID, proto.ChangedRooms)
	}

	room.Name = name
	return c.JSON(toResponse(room))
}

// Reorder 는 방 순서를 통째로 바꾼다. 관리자 이상만. PUT /v1/azits/{id}/rooms/order
// 바뀐 전체 목록을 돌려준다.
func (h *Handler) Reorder(c fiber.Ctx) error {
	me := member.Current(c)
	if err := requireAdmin(me); err != nil {
		return err
	}

	var req proto.ReorderRoomsRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.ErrBadRequest
	}

	before, err := h.roomRepo.ListByAzit(c.Context(), me.AzitID) // 되돌리기용 옛 순서
	if err != nil {
		return err
	}
	err = h.roomRepo.Reorder(c.Context(), me.AzitID, req.RoomIDs)
	if errors.Is(err, ErrBadOrder) {
		return httperr.New(fiber.StatusBadRequest, proto.ErrCodeBadRequest, err.Error())
	}
	if err != nil {
		return err
	}
	oldIDs := make([]int64, 0, len(before))
	for _, r := range before {
		oldIDs = append(oldIDs, r.RoomID)
	}
	h.logs.Record(c.Context(), changelog.Entry{
		AzitID: me.AzitID, ActorID: me.AccountID, ActorNick: me.NickName,
		Kind: changelog.KindRoomOrder, Summary: "방 순서를 바꿈",
	}, map[string][]int64{"room_ids": oldIDs})
	h.live.AzitChanged(c.Context(), me.AzitID, proto.ChangedRooms)

	return h.writeList(c, me)
}

// writeList 는 아지트 방 목록에 내 안 읽은 수를 붙여 응답한다. 목록과 순서 바꾸기가 같이 쓴다.
func (h *Handler) writeList(c fiber.Ctx, me member.Member) error {
	rooms, err := h.roomRepo.ListByAzit(c.Context(), me.AzitID)
	if err != nil {
		return err
	}
	unread, err := h.unread.CountUnread(c.Context(), me.AzitID, me.AccountID)
	if err != nil {
		return err
	}

	def, err := h.roomRepo.DefaultID(c.Context(), me.AzitID)
	if err != nil {
		return err
	}

	res := toResponses(rooms)
	for i := range res {
		res[i].Unread = unread[res[i].RoomID] // 없는 방은 0
		res[i].IsDefault = res[i].RoomID == def
	}
	return c.JSON(res)
}

// requireAdmin 은 관리자 이상이 아니면 403 을 돌려준다.
func requireAdmin(me member.Member) error {
	if !me.Role.AtLeast(member.RoleAdmin) {
		return httperr.New(fiber.StatusForbidden, proto.ErrCodeForbidden, "관리자만 할 수 있어요")
	}
	return nil
}

func toResponse(r Room) proto.RoomResponse {
	return proto.RoomResponse{
		RoomID:   r.RoomID,
		AzitID:   r.AzitID,
		Name:     r.Name,
		Position: r.Position,
	}
}

// toResponses 는 목록을 바꾼다. 비어 있어도 null 이 아니라 [] 로 나가게.
func toResponses(rooms []Room) []proto.RoomResponse {
	res := make([]proto.RoomResponse, 0, len(rooms))
	for _, r := range rooms {
		res = append(res, toResponse(r))
	}
	return res
}
