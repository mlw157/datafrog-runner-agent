package main

import (
	"fmt"
	"log"
	"os"
	"runner-agent/internal/agent"
	"runner-agent/internal/api/aws"
	"runner-agent/internal/api/controller"
)

func main() {

	controllerURL := os.Getenv("CONTROLLER_API_URL")
	controllerToken := os.Getenv("CONTROLLER_API_TOKEN")

	if controllerURL == "" || controllerToken == "" {
		log.Fatal("missing required environment variables: CONTROLLER_API_URL or CONTROLLER_API_TOKEN")
	}

	fmt.Println(controllerURL)

	controllerClient := controller.NewDatafrogControllerClient(controllerURL, controllerToken)
	awsClient := aws.NewAgentAWSClient()

	frogAgent := agent.NewAgent(*controllerClient, *awsClient)

	err := frogAgent.Deploy()
	if err != nil {
		log.Fatal(err)
	}
}
