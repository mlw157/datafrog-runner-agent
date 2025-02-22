package models

import "time"

type Instance struct {
	InstanceID       string    `json:"instance_id"`
	Type             string    `json:"type"`
	AvailabilityZone string    `json:"availability_zone"`
	PrivateIPAddress string    `json:"private_ip_address"`
	LastSeenAt       time.Time `json:"last_seen_at"`
}
