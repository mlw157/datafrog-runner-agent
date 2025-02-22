package agent

import (
	"fmt"
	"runner-agent/internal/api/aws"
	"runner-agent/internal/api/controller"
	"runner-agent/internal/models"
	"time"
)

type Agent struct {
	ControllerClient controller.DatafrogControllerClient
	AWSClient        aws.AgentAWSClient
	Instance         *models.Instance
	Jobs             []models.Job
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
		return fmt.Errorf("error retrieving instance: %w", err)
	}

	err = a.ControllerClient.CreateInstance(*a.Instance)
	if err != nil {
		return err
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
