package models

type UploadResponse struct {
	Success   bool         `json:"success"`
	Message   string       `json:"message"`
	Processed []FileResult `json:"processed_files,omitempty"`
	Errors    []FileError  `json:"errors,omitempty"`
}

type FileResult struct {
	Name        string `json:"name"`
	SizeBytes   int64  `json:"size_bytes"`
	ProcessedAt string `json:"processed_at"`
}

type FileError struct {
	Name  string `json:"name"`
	Error string `json:"error"`
}
