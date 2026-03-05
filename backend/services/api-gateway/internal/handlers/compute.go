package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/example/appfoundrylab/backend/services/api-gateway/internal/worker"
	"github.com/example/appfoundrylab/backend/services/api-gateway/pkg/httpx"
)

type fibonacciRequest struct {
	N uint32 `json:"n"`
}

type hashRequest struct {
	Input string `json:"input"`
}

type ComputeHandler struct {
	workerClient *worker.Client
}

func NewComputeHandler(workerClient *worker.Client) *ComputeHandler {
	return &ComputeHandler{workerClient: workerClient}
}

func (h *ComputeHandler) Fibonacci(w http.ResponseWriter, r *http.Request) {
	if h.workerClient == nil {
		httpx.WriteError(w, r, http.StatusServiceUnavailable, "worker_unavailable", "worker service is unavailable", nil)
		return
	}

	var payload fibonacciRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		httpx.WriteError(w, r, http.StatusBadRequest, "invalid_json", "invalid request body", nil)
		return
	}

	result, err := h.workerClient.ComputeFibonacci(r.Context(), payload.N)
	if err != nil {
		httpx.WriteError(w, r, http.StatusBadGateway, "worker_call_failed", "worker call failed", nil)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, map[string]any{
		"n":     result.N,
		"value": result.Value,
	})
}

func (h *ComputeHandler) Hash(w http.ResponseWriter, r *http.Request) {
	if h.workerClient == nil {
		httpx.WriteError(w, r, http.StatusServiceUnavailable, "worker_unavailable", "worker service is unavailable", nil)
		return
	}

	var payload hashRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		httpx.WriteError(w, r, http.StatusBadRequest, "invalid_json", "invalid request body", nil)
		return
	}

	result, err := h.workerClient.ComputeHash(r.Context(), payload.Input)
	if err != nil {
		httpx.WriteError(w, r, http.StatusBadGateway, "worker_call_failed", "worker call failed", nil)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, map[string]any{
		"algorithm": result.Algorithm,
		"hash":      result.Hash,
	})
}
