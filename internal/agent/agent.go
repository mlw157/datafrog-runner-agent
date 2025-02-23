package agent

import (
	"fmt"
	"os"
	"path/filepath"
	"runner-agent/internal/api/aws"
	"runner-agent/internal/api/controller"
	"runner-agent/internal/models"
	"slices"
	"strings"
	"time"
)

type Agent struct {
	ControllerClient controller.DatafrogControllerClient
	AWSClient        aws.AgentAWSClient
	Instance         *models.Instance
	JobFiles         []string
	InitTime         time.Time
}

func NewAgent(controllerClient controller.DatafrogControllerClient, awsClient aws.AgentAWSClient) *Agent {
	return &Agent{
		ControllerClient: controllerClient,
		AWSClient:        awsClient,
		InitTime:         time.Now(),
	}
}

func (a *Agent) Deploy() error {
	a.Instance = &models.Instance{}
	err := a.getInstance()
	if err != nil {
		return fmt.Errorf("agent deploy: error retrieving instance: %w", err)
	}

	err = a.ControllerClient.CreateInstance(*a.Instance)
	if err != nil {
		return fmt.Errorf("agent deploy: error creating instance: %w", err)
	}

	return nil
}

func (a *Agent) getInstance() error {
	metadataToken, err := a.AWSClient.GetMetadataToken()
	if err != nil {
		return err
	}

	instanceID, err := a.AWSClient.GetMetadata(metadataToken, "instance-id")
	if err != nil {
		return err
	}

	instanceType, err := a.AWSClient.GetMetadata(metadataToken, "instance-type")
	if err != nil {
		return err
	}

	instanceAZ, err := a.AWSClient.GetMetadata(metadataToken, "placement/availability-zone")
	if err != nil {
		return err
	}

	instancePrivateIP, err := a.AWSClient.GetMetadata(metadataToken, "local-ipv4")
	if err != nil {
		return err
	}

	a.Instance.InstanceID = instanceID
	a.Instance.Type = instanceType
	a.Instance.AvailabilityZone = instanceAZ
	a.Instance.PrivateIPAddress = instancePrivateIP
	a.Instance.LastSeenAt = time.Now()

	return nil
}

func (a *Agent) getJobs(root string) ([]string, error) {
	var newJobFiles []string
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		// job logs are located inside _diag folders
		if d.IsDir() && d.Name() != "_diag" {
			return filepath.SkipDir
		}

		if !d.IsDir() && strings.HasPrefix(d.Name(), "Worker_") {
			if !slices.Contains(a.JobFiles, path) {
				a.JobFiles = append(a.JobFiles, path)
				newJobFiles = append(newJobFiles, path)
			}
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return newJobFiles, nil
}
