package agent

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
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
	MonitorInterval  int
}

func NewAgent(controllerClient controller.DatafrogControllerClient, awsClient aws.AgentAWSClient, monitorInterval int) *Agent {
	return &Agent{
		ControllerClient: controllerClient,
		AWSClient:        awsClient,
		InitTime:         time.Now(),
		MonitorInterval:  monitorInterval,
	}
}

func (a *Agent) Deploy() error {
	a.Instance = &models.Instance{}
	err := a.getInstance()
	if err != nil {
		log.Printf("agent deploy: error retrieving instance: %w", err)
	}

	err = a.ControllerClient.CreateInstance(*a.Instance)
	if err != nil {
		log.Printf("agent deploy: error creating instance: %w", err)
	}

	a.monitor(a.MonitorInterval)

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

func (a *Agent) monitor(monitorInterval int) {
	interval := time.Duration(monitorInterval) * time.Second
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			newJobFiles, err := a.getNewJobFiles(".")
			if err != nil {
				log.Fatalf("agent monitor: error finding job files: %v", err)
			}

			for _, file := range newJobFiles {
				job, err := a.processJobFile(file)
				if err != nil {
					log.Printf("agent monitor: error processing job file: %v", err)
					continue
				}
				if job != nil {
					err = a.ControllerClient.CreateJob(*job)
					if err != nil {
						log.Printf("agent monitor: error sending job to controller api: %v", err)
					}
				}

			}
		}

	}
}

func (a *Agent) getNewJobFiles(root string) ([]string, error) {
	log.Println("finding worker files")
	var newJobFiles []string
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		// job logs are located inside _diag folders
		//if d.IsDir() && d.Name() != "_diag" {
		//	return filepath.SkipDir
		//}

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

	log.Printf("found files %v\n", newJobFiles)

	return newJobFiles, nil
}

func (a *Agent) processJobFile(path string) (*models.Job, error) {
	jobFile, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer jobFile.Close()

	var org, repo, workflow string
	var startTime, endTime time.Time

	repoPattern := regexp.MustCompile(`"v": "https://api.github.com/repos/([^/]+)/([^/]+)`)
	workflowPattern := regexp.MustCompile(`workflows/([^"]+)`)
	timePattern := regexp.MustCompile(`\[(\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}Z)`)

	scanner := bufio.NewScanner(jobFile)
	for scanner.Scan() {
		line := scanner.Text()

		repoMatch := repoPattern.FindStringSubmatch(line)
		if repoMatch != nil {
			org = repoMatch[1]
			repo = repoMatch[2]
		}

		workflowMatch := workflowPattern.FindStringSubmatch(line)
		if workflowMatch != nil {
			workflow = workflowMatch[1]
		}

		timeMatch := timePattern.FindStringSubmatch(line)
		if timeMatch != nil {
			timestamp, err := time.Parse("2006-01-02 15:04:05Z", timeMatch[1])
			if err != nil {
				return nil, err
			}
			// startTime will be the first timestamp found
			if startTime.IsZero() {
				startTime = timestamp
			}
			// endTime will always update so it will be the last timestamp of the log file
			endTime = timestamp
		}
	}
	err = scanner.Err()
	if err != nil {
		return nil, fmt.Errorf("error reading log file: %w", err)
	}

	return &models.Job{
		InstanceID:   a.Instance.InstanceID,
		Organization: org,
		Repository:   repo,
		Workflow:     workflow,
		StartTime:    startTime,
		EndTime:      endTime,
	}, nil
}
