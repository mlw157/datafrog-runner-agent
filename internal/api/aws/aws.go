package aws

import (
	"io"
	"net/http"
	"net/url"
)

type AgentAWSClient struct {
	HTTPClient *http.Client
}

func NewAgentAWSClient() *AgentAWSClient {
	return &AgentAWSClient{
		HTTPClient: &http.Client{},
	}
}

func (a *AgentAWSClient) GetMetadataToken() (string, error) {
	requestURL := "http://169.254.169.254/latest/api/token"

	req, err := http.NewRequest("PUT", requestURL, nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("X-aws-ec2-metadata-token-ttl-seconds", "21600")

	resp, err := a.HTTPClient.Do(req)
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

func (a *AgentAWSClient) GetMetadata(token, path string) (string, error) {
	requestURL := "http://169.254.169.254/latest/meta-data/" + url.QueryEscape(path)

	req, err := http.NewRequest("GET", requestURL, nil)
	if err != nil {
		return "", nil
	}

	req.Header.Set("X-aws-ec2-metadata-token", token)

	resp, err := a.HTTPClient.Do(req)
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
