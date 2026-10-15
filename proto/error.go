package proto

type ErrorResponse struct {
	Error ErrorBody `json:"error"`
}

type ErrorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

const (
	ErrCodeUnauthorized = "unauthorized"
	ErrCodeBadRequest   = "bad_request"
	ErrCodeForbidden    = "forbidden"
	ErrCodeNotFound     = "not_found"
	ErrCodeInternal     = "internal"

	ErrCodeGoogleExchangeFailed = "google_exchange_failed"
	ErrCodeAccountDisabled      = "account_disabled"
	ErrCodeNameTaken            = "name_taken"
	ErrCodeInviteUnusable       = "invite_unusable"
	ErrCodeAlreadyMember        = "already_member"
)
