package member

import (
	"errors"

	"github.com/TogetherBros/azit/proto"
	"github.com/TogetherBros/azit/server/internal/auth"
	"github.com/TogetherBros/azit/server/internal/httperr"
	"github.com/gofiber/fiber/v3"
)

// localsKey 는 c.Locals 에 멤버를 넣을 때 쓰는 키. 패키지 밖에서는 Current 로만 꺼낸다.
type localsKey struct{}

// RequireMember 는 경로의 :id 아지트에 지금 들어와 있는 멤버만 통과시킨다.
// requireAuth 뒤에 와야 한다. 멤버가 아니면 404 (아지트가 있다는 것도 숨김).
func RequireMember(repo *Repository) fiber.Handler {
	return func(c fiber.Ctx) error {
		azitID := fiber.Params[int64](c, "id") // 숫자가 아니면 0
		if azitID <= 0 {
			return fiber.ErrBadRequest
		}

		m, err := repo.FindActive(c.Context(), azitID, auth.AccountID(c))
		if errors.Is(err, ErrNotMember) {
			return httperr.New(fiber.StatusNotFound, proto.ErrCodeNotFound, "아지트가 없어요")
		}
		if err != nil {
			return err
		}

		c.Locals(localsKey{}, m)
		return c.Next()
	}
}

// Current 는 RequireMember 가 넣어둔 멤버를 꺼낸다.
func Current(c fiber.Ctx) Member {
	return fiber.Locals[Member](c, localsKey{})
}
