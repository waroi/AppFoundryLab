package httpx

import (
	"context"
	"net/http"
)

const TraceIDHeader = "X-Trace-Id"

type traceIDKey string

const traceIDContextKey traceIDKey = "trace_id"

type ErrorEnvelope struct {
	Error ErrorBody `json:"error"`
}

type ErrorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	TraceID string `json:"traceId,omitempty"`
	Details any    `json:"details,omitempty"`
}

func ContextWithTraceID(ctx context.Context, traceID string) context.Context {
	return context.WithValue(ctx, traceIDContextKey, traceID)
}

func TraceIDFromContext(ctx context.Context) string {
	traceID, _ := ctx.Value(traceIDContextKey).(string)
	return traceID
}

func WriteError(w http.ResponseWriter, r *http.Request, status int, code, message string, details any) {
	WriteJSON(w, status, ErrorEnvelope{
		Error: ErrorBody{
			Code:    code,
			Message: message,
			TraceID: TraceIDFromContext(r.Context()),
			Details: details,
		},
	})
}
