package entities

type ResizeResultStatus string

const (
	ResizeResultStatusSuccess    ResizeResultStatus = "success"
	ResizeResultStatusFailure    ResizeResultStatus = "failure"
	ResizeResultStatusProcessing ResizeResultStatus = "processing"
)

type ResizeRequest struct {
	URLs   []string `json:"urls"`
	Width  uint     `json:"width"`
	Height uint     `json:"height"`
}

type ResizeResult struct {
	Result ResizeResultStatus `json:"result"`
	URL    string             `json:"url,omitempty"`
	Cached bool               `json:"cached"`
}
