package proto

type UploadRequest struct {
	Kind        string `json:"kind"`
	ContentType string `json:"content_type"`
	Size        int64  `json:"size"`
}

type UploadResponse struct {
	Key string `json:"key"`
	URL string `json:"url"`
}

type SetImageRequest struct {
	Key string `json:"key"`
}
