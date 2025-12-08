package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"noytech-ga-optimizer/internal/models"
	"noytech-ga-optimizer/internal/services/importer"
	"noytech-ga-optimizer/pkg/errors"
)

const MaxMemorySize int64 = 32 << 20

type UploadHandler struct {
	importer *importer.Service
	logger   *slog.Logger
}

func NewUploadHandler(imp *importer.Service, l *slog.Logger) *UploadHandler {
	return &UploadHandler{
		importer: imp,
		logger:   l,
	}
}

func (h *UploadHandler) HandleUpload(w http.ResponseWriter, r *http.Request) {
	logger := h.logger.With(slog.String("method", "HandleUpload"))

	if r.Method != http.MethodPost {
		appErr := errors.NewBadRequestError("method not allowed")
		h.sendError(w, appErr, logger, r)
		return
	}

	r.ParseMultipartForm(MaxMemorySize)

	formFileHeaders := r.MultipartForm.File["file"]
	if len(formFileHeaders) == 0 {
		logger.Warn("No files with key 'file' received in form")
		resp := models.UploadResponse{
			Success: false,
			Message: "No files received or all files were invalid (expected key 'file')",
		}
		h.sendJSON(w, resp, http.StatusBadRequest)
		return
	}

	var files []importer.FileData
	var errorsList []models.FileError
	var results []models.FileResult

	for _, fileHeader := range formFileHeaders {
		filename := fileHeader.Filename

		file, err := fileHeader.Open()
		if err != nil {
			logger.Error("Error opening uploaded file", "filename", filename, "error", err)
			errorsList = append(errorsList, models.FileError{
				Name:  filename,
				Error: fmt.Sprintf("failed to open file: %v", err),
			})
			continue
		}
		defer file.Close()

		content, err := io.ReadAll(file)
		if err != nil {
			logger.Error("Error reading file content", "filename", filename, "error", err)
			errorsList = append(errorsList, models.FileError{
				Name:  filename,
				Error: fmt.Sprintf("failed to read file: %v", err),
			})
			continue
		}

		files = append(files, importer.FileData{
			Name:    filename,
			Content: content,
		})

		results = append(results, models.FileResult{
			Name:        filename,
			SizeBytes:   int64(len(content)),
			ProcessedAt: time.Now().Format(time.RFC3339),
		})
	}

	if len(files) == 0 {
		logger.Warn("No files received in request")
		resp := models.UploadResponse{
			Success: false,
			Message: "No files received or all files were invalid",
			Errors:  errorsList,
		}
		h.sendJSON(w, resp, http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	err := h.importer.ImportFromXLSX(ctx, files)
	if err != nil {
		logger.Error("ImportFromXLSX failed", "error", err)
		if appErr, ok := err.(*errors.ErrorResponse); ok {
			h.sendError(w, appErr, logger, r)
			return
		}

		logger.Error("Unexpected error during import (not a custom ErrorResponse)", "error", err)
		h.sendError(w, errors.NewInternalServerError("an unexpected error occurred during import"), logger, r)
		return
	}

	logger.Info("Files imported successfully", "file_count", len(files))
	resp := models.UploadResponse{
		Success:   true,
		Message:   fmt.Sprintf("Successfully imported %d file(s)", len(files)),
		Processed: results,
		Errors:    errorsList,
	}
	h.sendJSON(w, resp, http.StatusOK)
}

func (h *UploadHandler) sendJSON(w http.ResponseWriter, data interface{}, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(data)
}

func (h *UploadHandler) sendError(w http.ResponseWriter, appErr *errors.ErrorResponse, logger *slog.Logger, r *http.Request) {
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
