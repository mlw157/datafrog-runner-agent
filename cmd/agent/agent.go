package main

import (
	"log"
	"os"
	"runner-agent/internal/agent"
	"runner-agent/internal/api/aws"
	"runner-agent/internal/api/controller"
	"strconv"
)

func main() {

	controllerURL := os.Getenv("CONTROLLER_API_URL")
	controllerToken := os.Getenv("CONTROLLER_API_TOKEN")
	monitorIntervalString := os.Getenv("MONITOR_INTERVAL")

	if controllerURL == "" || controllerToken == "" {
		log.Fatal("missing required controller environment variables: CONTROLLER_API_URL or CONTROLLER_API_TOKEN")
	}

	monitorInterval, err := strconv.Atoi(monitorIntervalString)
	if err != nil {
		log.Fatalf("invalid MONITOR_INTERVAL value: %v", err)
	}

	controllerClient := controller.NewDatafrogControllerClient(controllerURL, controllerToken)
	awsClient := aws.NewAgentAWSClient()

	frogAgent := agent.NewAgent(*controllerClient, *awsClient, monitorInterval)

	err = frogAgent.Deploy()
	if err != nil {
		log.Fatal(err)
	}
}
