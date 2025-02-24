package models

import "time"

type Job struct {
	InstanceID   string    `json:"instance_id"`
	Organization string    `json:"organization"`
	Repository   string    `json:"repository"`
	Workflow     string    `json:"workflow"`
	StartTime    time.Time `json:"start_time"`
	EndTime      time.Time `json:"end_time"`
}
