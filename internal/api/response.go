package api

type StatusResponse struct {
	Online bool `json:"online"`
}

type WakeResponse struct {
	Sent bool `json:"sent"` // TODO: get rid of it?
}

type ErrorResponse struct {
	Error *Error `json:"error"`
}

type Error struct {
	Code    ErrorCode `json:"code"`
	Message string    `json:"message"`
}
