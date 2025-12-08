package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/timestamppb"

	"noytech-ga-optimizer/api/proto"
	"noytech-ga-optimizer/internal/services/optimizer"
	"noytech-ga-optimizer/internal/validation"
	"noytech-ga-optimizer/pkg/errors"
)

type OptimizeHandler struct {
	optimizer *optimizer.Service
	logger    *slog.Logger
}

func NewOptimizeHandler(opt *optimizer.Service, l *slog.Logger) *OptimizeHandler {
	return &OptimizeHandler{
		optimizer: opt,
		logger:    l,
	}
}

func (h *OptimizeHandler) HandleOptimize(w http.ResponseWriter, r *http.Request) {
	logger := h.logger.With(slog.String("method", "HandleOptimize"))

	if r.Method != http.MethodPost {
		appErr := errors.NewErrInvalidArgument(nil, "method not allowed")
		h.sendError(w, appErr, appErr.Status, logger, r)
		return
	}

	var req proto.OptimizeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Error("Failed to decode request body", "error", err)
		appErr := errors.NewErrInvalidArgument(err, "invalid JSON in request body")
		h.sendError(w, appErr, appErr.Status, logger, r)
		return
	}

	req.Direction = strings.TrimSpace(req.Direction)
	for i, day := range req.DeliveryDays {
		req.DeliveryDays[i] = strings.TrimSpace(day)
	}

	if err := validation.ValidateOptimizeRequest(&req); err != nil {
		logger.Error("Validation failed", "error", err)
		if customErr, ok := err.(*errors.ErrorResponse); ok {
			h.sendError(w, customErr, customErr.Status, logger, r)
			return
		}
		appErr := errors.NewInternalServerError("validation error")
		h.sendError(w, appErr, appErr.Status, logger, r)
		return
	}

	ctx := r.Context()
	result, err := h.optimizer.Optimize(ctx, &req)
	if err != nil {
		logger.Error("Optimizer service failed", "error", err)
		if customErr, ok := err.(*errors.ErrorResponse); ok {
			h.sendError(w, customErr, customErr.Status, logger, r)
			return
		}
		appErr := errors.NewInternalServerError("optimization failed")
		h.sendError(w, appErr, appErr.Status, logger, r)
		return
	}

	resp := &proto.OptimizeResponse{
		Success:    true,
		Message:    "Optimization completed successfully",
		Results:    []*proto.OptimizationResult{result},
		SolutionId: uuid.NewString(),
		CreatedAt:  timestamppb.Now(),
	}

	logger.Info("Optimization completed successfully")
	h.sendJSON(w, resp, http.StatusOK)
}

func (h *OptimizeHandler) sendJSON(w http.ResponseWriter, data interface{}, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(data)
}

func (h *OptimizeHandler) sendError(w http.ResponseWriter, appErr *errors.ErrorResponse, statusCode int, logger *slog.Logger, r *http.Request) {
	requestID := ""
	if reqID := r.Context().Value("requestID"); reqID != nil {
		if id, ok := reqID.(string); ok {
			requestID = id
		}
	}

	if requestID != "" && appErr.RequestID == "" {
		appErr = errors.NewErrorResponseWithRequestID(appErr.Status, appErr.Message, appErr.Details, requestID)
	}

	if appErr.Status >= 500 {
		logger.Error("Internal error", "status", appErr.Status, "error", appErr.Error(), "request_id", appErr.RequestID)
	} else {
		logger.Warn("Client error", "status", appErr.Status, "error", appErr.Error(), "request_id", appErr.RequestID)
	}

	h.sendJSON(w, appErr, appErr.Status)
}
