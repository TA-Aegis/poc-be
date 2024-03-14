package models

type QueueEvents struct {
	UserID            string  `json:"userId"`
	QueueNumber       int     `json:"queueNumber"`
	EstimatedTime     float64 `json:"estimatedTime"`
	PercetageProgress float64 `json:"percentageProgress"`
	IsFinished        bool    `json:"isFinished"`
	Message           string  `json:"message,omitempty"`
}
