package models

type Image struct {
	FileType string `json:"file_type"`
	UploadDate int64 `json:"upload_date"`
	FileID string `json:"file_id"`
	FilePath string `json:"file_path"`
}

type ImageView struct {
	DirectURL string `json:"direct_url"`
	Filename string `json:"filename"`
	MainURL string `json:"main_url"`
}