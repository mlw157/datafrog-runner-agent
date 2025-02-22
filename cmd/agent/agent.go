package main

import (
	"log"
	"os"
	"runner-agent/internal/agent"
	"runner-agent/internal/api/aws"
	"runner-agent/internal/api/controller"
)

func main() {

	controllerURL := os.Getenv("CONTROLLER_API_URL")
	controllerToken := os.Getenv("CONTROLLER_API_TOKEN")

	controllerClient := controller.NewDatafrogControllerClient(controllerURL, controllerToken)
	awsClient := aws.NewAgentAWSClient()

	frogAgent := agent.NewAgent(*controllerClient, *awsClient)

	err := frogAgent.Deploy()
	if err != nil {
		log.Fatalf("error deploying agent: %v", err)
	}
}
