package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

type Instance struct {
	InstanceID       string    `json:"instance_id"`
	Type             string    `json:"type"`
	AvailabilityZone string    `json:"availability_zone"`
	PrivateIPAddress string    `json:"private_ip_address"`
	LastSeenAt       time.Time `json:"last_seen_at"`
}

func getMetadataToken() (string, error) {
	req, err := http.NewRequest("PUT", "http://169.254.169.254/latest/api/token", nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("X-aws-ec2-metadata-token-ttl-seconds", "21600")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	token, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(token), nil
}

func getMetadata(token, path string) (string, error) {
	req, err := http.NewRequest("GET", "http://169.254.169.254/latest/meta-data/"+path, nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("X-aws-ec2-metadata-token", token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

func collectInstanceInfo(token string) (*Instance, error) {
	instanceID, err := getMetadata(token, "instance-id")
	if err != nil {
		return nil, err
	}

	instanceType, err := getMetadata(token, "instance-type")
	if err != nil {
		return nil, err
	}

	az, err := getMetadata(token, "placement/availability-zone")
	if err != nil {
		return nil, err
	}

	privateIP, err := getMetadata(token, "local-ipv4")
	if err != nil {
		return nil, err
	}

	return &Instance{
		InstanceID:       instanceID,
		Type:             instanceType,
		AvailabilityZone: az,
		PrivateIPAddress: privateIP,
		LastSeenAt:       time.Now().UTC(),
	}, nil
}

func sendInstanceInfo(info *Instance) error {
	data, err := json.Marshal(info)
	if err != nil {
		return err
	}

	resp, err := http.Post(
		"https://runner-controller.cfappsecurity.com/api/instances",
		"application/json",
		bytes.NewBuffer(data),
	)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}

func main() {
	token, err := getMetadataToken()
	if err != nil {
		fmt.Printf("Error getting token: %v\n", err)
		os.Exit(1)
	}

	for {
		info, err := collectInstanceInfo(token)
		if err != nil {
			fmt.Printf("Error collecting instance info: %v\n", err)
		} else {
			err = sendInstanceInfo(info)
			if err != nil {
				fmt.Printf("Error sending instance info: %v\n", err)
			} else {
				fmt.Printf("Successfully sent instance info: %+v\n", info)
			}
		}

		time.Sleep(30 * time.Second)
	}
}
