package models

import "time"

type MemoryLog struct {
	InstanceID string    `json:"instance_id"`
	Total      int       `json:"total"`
	Used       int       `json:"used"`
	Free       int       `json:"free"`
	Timestamp  time.Time `json:"timestamp"`
}
