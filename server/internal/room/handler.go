package room

type Handler struct {
	roomRepo *Repository
}

func NewHandler(roomRepo *Repository) *Handler {
	return &Handler{roomRepo: roomRepo}
}
