package models

import (
	"time"
)

type Client struct {
	ID         string            `json:"id"`
	RegisterAt time.Time         `json:"registerAt"`
	Events     chan *QueueEvents `json:"-"`
}
