package room

import (
	"errors"

	"github.com/TogetherBros/azit/proto"
	"github.com/TogetherBros/azit/server/internal/httperr"
	"github.com/TogetherBros/azit/server/internal/member"
	"github.com/gofiber/fiber/v3"
)

type Handler struct {
	roomRepo *Repository
}

func NewHandler(roomRepo *Repository) *Handler {
	return &Handler{roomRepo: roomRepo}
}

// Routes 는 /v1/azits 그룹 아래에 방 경로를 단다.
// requireMember 는 member.RequireMember 로 만든 미들웨어. 모든 경로 앞에 끼운다.
func (h *Handler) Routes(r fiber.Router, requireMember fiber.Handler) {
	r.Get("/:id/rooms", requireMember, h.List)
	r.Post("/:id/rooms", requireMember, h.Create)
	r.Put("/:id/rooms/order", requireMember, h.Reorder)
	r.Patch("/:id/rooms/:roomId", requireMember, h.Rename)
}

// List 는 아지트의 방 목록. 멤버면 누구나. GET /v1/azits/{id}/rooms
func (h *Handler) List(c fiber.Ctx) error {
	azitID := fiber.Params[int64](c, "id")

	rooms, err := h.roomRepo.ListByAzit(c.Context(), azitID)
	if err != nil {
		return err
	}
	return c.JSON(toResponses(rooms))
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

	err := h.roomRepo.Reorder(c.Context(), me.AzitID, req.RoomIDs)
	if errors.Is(err, ErrBadOrder) {
		return httperr.New(fiber.StatusBadRequest, proto.ErrCodeBadRequest, err.Error())
	}
	if err != nil {
		return err
	}

	rooms, err := h.roomRepo.ListByAzit(c.Context(), me.AzitID)
	if err != nil {
		return err
	}
	return c.JSON(toResponses(rooms))
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
