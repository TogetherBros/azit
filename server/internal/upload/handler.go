package upload

import (
	"errors"

	"github.com/TogetherBros/azit/proto"
	"github.com/TogetherBros/azit/server/internal/auth"
	"github.com/TogetherBros/azit/server/internal/httperr"
	"github.com/TogetherBros/azit/server/internal/storage"
	"github.com/gofiber/fiber/v3"
)

type Handler struct {
	r2 *storage.R2
}

func NewHandler(r2 *storage.R2) *Handler {
	return &Handler{r2: r2}
}

// Routes 는 /v1/uploads 아래 경로를 붙인다. r 에는 인증 미들웨어가 걸려 있어야 한다.
func (h *Handler) Routes(r fiber.Router) {
	r.Post("/", h.Create)
}

func (h *Handler) Create(c fiber.Ctx) error {
	var req proto.UploadRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.ErrBadRequest
	}

	accountID := auth.AccountID(c)
	key, err := storage.NewKey(storage.Kind(req.Kind), accountID, req.ContentType, req.Size)
	if errors.Is(err, storage.ErrUnknownKind) || errors.Is(err, storage.ErrBadType) || errors.Is(err, storage.ErrTooLarge) {
		return httperr.New(fiber.StatusBadRequest, proto.ErrCodeBadRequest, err.Error())
	}

	if err != nil {
		return err
	}

	url, err := h.r2.PresignPut(c.Context(), key, req.ContentType, req.Size)
	if err != nil {
		return err
	}

	return c.JSON(proto.UploadResponse{
		Key: key,
		URL: url,
	})

}
