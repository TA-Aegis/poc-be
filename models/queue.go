package models

type QueueEvents struct {
	UserID        string  `json:"userId"`
	QueueNumber   int     `json:"queueNumber"`
	EstimatedTime float64 `json:"estimatedTime"`
	Message       string  `json:"message,omitempty"`
}
