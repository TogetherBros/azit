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
	ErrCodeTempIDConflict       = "temp_id_conflict" // 같은 번호표(temp_id)인데 내용이 다르게 왔다 (409)
	ErrCodeReactLimit           = "react_limit"      // 한 메시지에 맞장구 종류가 이미 10개다 (409)
)
