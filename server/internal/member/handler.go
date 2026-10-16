package member

import (
	"context"
	"errors"
	"slices"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/TogetherBros/azit/proto"
	"github.com/TogetherBros/azit/server/internal/account"
	"github.com/TogetherBros/azit/server/internal/httperr"
	"github.com/TogetherBros/azit/server/internal/storage"
	"github.com/gofiber/fiber/v3"
)

const maxNickName = 16 // 계정 닉네임과 같은 규칙

type Handler struct {
	memberRepo  *Repository
	accountRepo *account.Repository
	r2          *storage.R2
}

func NewHandler(memberRepo *Repository, accountRepo *account.Repository, r2 *storage.R2) *Handler {
	return &Handler{memberRepo: memberRepo, accountRepo: accountRepo, r2: r2}
}

// Routes 는 /v1/azits 그룹 아래에 멤버 경로를 단다.
// "/members/me" 를 "/members/:accountId" 보다 먼저 단다. Fiber 는 단 순서대로 맞춰 보기 때문에,
// 거꾸로 달면 "me" 가 :accountId 로 잡힌다.
func (h *Handler) Routes(r fiber.Router, requireMember fiber.Handler) {
	r.Get("/:id/members", requireMember, h.List)
	r.Patch("/:id/members/me", requireMember, h.UpdateMe)
	r.Delete("/:id/members/me", requireMember, h.Leave)
	r.Patch("/:id/members/:accountId/role", requireMember, h.UpdateRole)
	r.Delete("/:id/members/:accountId", requireMember, h.Kick)
}

// List 는 지금 들어와 있는 멤버 목록. 멤버면 누구나. GET /v1/azits/{id}/members
// 순서: 만든 사람 → 관리자 → 멤버, 같은 역할이면 들어온 순서.
func (h *Handler) List(c fiber.Ctx) error {
	me := Current(c)
	members, err := h.memberRepo.ListActive(c.Context(), me.AzitID)
	if err != nil {
		return err
	}
	res, err := h.toResponses(c.Context(), members)
	if err != nil {
		return err
	}
	// 역할 높은 순으로. 같은 역할 안에서는 ListActive 가 준 들어온 순서를 지킨다 (안정 정렬).
	sortByRole(res)
	return c.JSON(res)
}

// UpdateMe 는 이 아지트 안 내 닉네임을 바꾼다. PATCH /v1/azits/{id}/members/me
func (h *Handler) UpdateMe(c fiber.Ctx) error {
	me := Current(c)

	var req proto.UpdateMyMemberRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.ErrBadRequest
	}
	nick := strings.TrimSpace(req.Nickname)
	if n := utf8.RuneCountInString(nick); n == 0 {
		return badRequest("닉네임을 넣어 주세요")
	} else if n > maxNickName {
		return badRequest("닉네임은 16자까지예요")
	}

	if err := h.memberRepo.UpdateNickName(c.Context(), me.AzitID, me.AccountID, nick); err != nil {
		return err
	}
	me.NickName = nick
	return h.respondOne(c, me)
}

// Leave 는 아지트에서 스스로 나간다. DELETE /v1/azits/{id}/members/me
// 만든 사람(super_admin)은 나갈 수 없다. 넘기기 기능이 생기기 전까지는 아지트에 주인이 없어지면 안 되니까.
func (h *Handler) Leave(c fiber.Ctx) error {
	me := Current(c)
	if me.Role == RoleSuperAdmin {
		return badRequest("만든 사람은 아지트를 나갈 수 없어요")
	}
	if err := h.memberRepo.Leave(c.Context(), me.AzitID, me.AccountID, time.Now()); err != nil {
		return err
	}
	return c.SendStatus(fiber.StatusNoContent)
}

// UpdateRole 은 다른 멤버를 관리자로 올리거나 멤버로 내린다. 만든 사람만.
// PATCH /v1/azits/{id}/members/{accountId}/role
func (h *Handler) UpdateRole(c fiber.Ctx) error {
	me := Current(c)
	if me.Role != RoleSuperAdmin {
		return forbidden("만든 사람만 역할을 바꿀 수 있어요")
	}

	target, err := h.target(c, me)
	if err != nil {
		return err
	}
	if target.Role == RoleSuperAdmin {
		return forbidden("만든 사람의 역할은 바꿀 수 없어요")
	}

	var req proto.UpdateMemberRoleRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.ErrBadRequest
	}
	role := Role(req.Role)
	if role != RoleAdmin && role != RoleMember {
		return badRequest("관리자나 멤버만 고를 수 있어요")
	}

	if err := h.memberRepo.UpdateRole(c.Context(), me.AzitID, target.AccountID, role); err != nil {
		return err
	}
	target.Role = role
	return h.respondOne(c, target)
}

// Kick 은 멤버를 내보낸다. DELETE /v1/azits/{id}/members/{accountId}
// 관리자 이상만, 그리고 나보다 역할이 낮은 사람만 (관리자는 멤버만, 만든 사람은 관리자와 멤버).
// 내보내진 사람도 초대 코드가 있으면 다시 들어올 수 있다.
func (h *Handler) Kick(c fiber.Ctx) error {
	me := Current(c)
	if !me.Role.AtLeast(RoleAdmin) {
		return forbidden("관리자만 내보낼 수 있어요")
	}

	target, err := h.target(c, me)
	if err != nil {
		return err
	}
	if target.Role.AtLeast(me.Role) {
		return forbidden("나보다 역할이 낮은 사람만 내보낼 수 있어요")
	}

	if err := h.memberRepo.Leave(c.Context(), me.AzitID, target.AccountID, time.Now()); err != nil {
		return err
	}
	return c.SendStatus(fiber.StatusNoContent)
}

// target 은 경로의 :accountId 멤버를 찾는다. 나 자신이면 400, 멤버가 아니면 404.
func (h *Handler) target(c fiber.Ctx, me Member) (Member, error) {
	accountID := fiber.Params[int64](c, "accountId")
	if accountID <= 0 {
		return Member{}, fiber.ErrBadRequest
	}
	if accountID == me.AccountID {
		return Member{}, badRequest("나 자신에게는 할 수 없어요")
	}
	t, err := h.memberRepo.FindActive(c.Context(), me.AzitID, accountID)
	if errors.Is(err, ErrNotMember) {
		return Member{}, httperr.New(fiber.StatusNotFound, proto.ErrCodeNotFound, "이 아지트 멤버가 아니에요")
	}
	return t, err
}

func (h *Handler) respondOne(c fiber.Ctx, m Member) error {
	res, err := h.toResponses(c.Context(), []Member{m})
	if err != nil {
		return err
	}
	return c.JSON(res[0])
}

// toResponses 는 멤버 줄에 계정 정보(이름, 색, 사진, 소개)를 붙인다. 비어 있어도 [] 로 나가게.
func (h *Handler) toResponses(ctx context.Context, members []Member) ([]proto.MemberResponse, error) {
	ids := make([]int64, 0, len(members))
	for _, m := range members {
		ids = append(ids, m.AccountID)
	}
	accounts, err := h.accountRepo.FindByIDs(ctx, ids)
	if err != nil {
		return nil, err
	}

	res := make([]proto.MemberResponse, 0, len(members))
	for _, m := range members {
		acc := accounts[m.AccountID]
		avatarURL := ""
		if acc.AvatarKey != "" {
			if avatarURL, err = h.r2.PresignGet(ctx, acc.AvatarKey); err != nil {
				return nil, err
			}
		}
		name := ""
		if acc.Name != nil {
			name = *acc.Name
		}
		res = append(res, proto.MemberResponse{
			AccountID: m.AccountID,
			Name:      name,
			Nickname:  m.NickName,
			Color:     acc.Color,
			AvatarURL: avatarURL,
			Bio:       acc.Bio,
			Role:      string(m.Role),
			JoinedAt:  m.JoinedAt,
		})
	}
	return res, nil
}

// sortByRole 은 역할 높은 순으로 안정 정렬한다.
func sortByRole(res []proto.MemberResponse) {
	slices.SortStableFunc(res, func(a, b proto.MemberResponse) int {
		return rank[Role(b.Role)] - rank[Role(a.Role)]
	})
}

func badRequest(msg string) error {
	return httperr.New(fiber.StatusBadRequest, proto.ErrCodeBadRequest, msg)
}

func forbidden(msg string) error {
	return httperr.New(fiber.StatusForbidden, proto.ErrCodeForbidden, msg)
}
