package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

type Instance struct {
	InstanceID       string    `json:"instance_id"`
	Type             string    `json:"type"`
	AvailabilityZone string    `json:"availability_zone"`
	PrivateIPAddress string    `json:"private_ip_address"`
	LastSeenAt       time.Time `json:"last_seen_at"`
}

type Job struct {
	InstanceID string `json:"instance_id"`
	Repository string `json:"repository"`
	Workflow   string `json:"workflow"`
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

func findWorkerFiles(rootDir string) ([]string, error) {
	var workerFiles []string
	err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if strings.HasPrefix(info.Name(), "Worker_") {
			workerFiles = append(workerFiles, path)
		}
		return nil
	})
	return workerFiles, err
}

func extractRepository(filePath string) (string, error) {
	cmd := exec.Command("sh", "-c", fmt.Sprintf(`grep '"v": "https://api.github.com/repos/' %s | awk -F'/' '{print $5 "/" $6}' | head -1`, filePath))
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

func extractWorkflow(filePath string) (string, error) {
	cmd := exec.Command("sh", "-c", fmt.Sprintf(`grep 'workflows/' %s | grep '@refs/heads/' | awk -F'"' '{print $4}'`, filePath))
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
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

func sendJob(job *Job) error {
	data, err := json.Marshal(job)
	if err != nil {
		return err
	}

	resp, err := http.Post(
		"https://runner-controller.cfappsecurity.com/api/jobs",
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
		instance, err := collectInstanceInfo(token)
		if err != nil {
			fmt.Printf("Error collecting instance info: %v\n", err)
			continue
		}

		// Send instance info first
		err = sendInstanceInfo(instance)
		if err != nil {
			fmt.Printf("Error sending instance info: %v\n", err)
		} else {
			fmt.Printf("Successfully sent instance info: %+v\n", instance)
		}

		workerFiles, err := findWorkerFiles("/app/worker_files")
		if err != nil {
			fmt.Printf("Error finding worker files: %v\n", err)
			continue
		}

		for _, file := range workerFiles {
			repo, err := extractRepository(file)
			if err != nil || repo == "" {
				fmt.Printf("Error extracting repo from %s: %v\n", file, err)
				continue
			}

			workflow, err := extractWorkflow(file)
			if err != nil || workflow == "" {
				fmt.Printf("Error extracting workflow from %s: %v\n", file, err)
				continue
			}

			job := &Job{
				InstanceID: instance.InstanceID,
				Repository: repo,
				Workflow:   workflow,
			}

			// Send job info
			err = sendJob(job)
			if err != nil {
				fmt.Printf("Error sending job: %v\n", err)
			} else {
				fmt.Printf("Successfully sent job: %+v\n", job)
			}
		}

		time.Sleep(30 * time.Second)
	}
}
