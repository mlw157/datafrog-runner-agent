package controller

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"runner-agent/internal/models"
)

type DatafrogControllerClient struct {
	HTTPClient *http.Client
	BaseURL    string
	Token      string
}

func NewDatafrogControllerClient(baseURL, token string) *DatafrogControllerClient {
	return &DatafrogControllerClient{
		HTTPClient: &http.Client{},
		BaseURL:    baseURL,
		Token:      token,
	}
}

func (d *DatafrogControllerClient) CreateInstance(instance models.Instance) error {
	requestURL := d.BaseURL + "/instances"

	bodyData, err := json.Marshal(instance)
	if err != nil {
		return fmt.Errorf("failed to encode request body: %w", err)
	}

	req, err := http.NewRequest("POST", requestURL, bytes.NewBuffer(bodyData))
	if err != nil {
		return err
	}

	//req.Header.Set("Authorization", "Token "+d.Token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := d.HTTPClient.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("failed to read error response: %w", err)
		}

		return fmt.Errorf("controller returned status code %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

func (d *DatafrogControllerClient) CreateJob(job models.Job) error {
	requestURL := d.BaseURL + "/jobs"

	bodyData, err := json.Marshal(job)
	if err != nil {
		return fmt.Errorf("failed to encode request body: %w", err)
	}

	req, err := http.NewRequest("POST", requestURL, bytes.NewBuffer(bodyData))
	if err != nil {
		return err
	}

	//req.Header.Set("Authorization", "Token "+d.Token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := d.HTTPClient.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("failed to read error response: %w", err)
		}

		return fmt.Errorf("controller returned status code %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

func (d *DatafrogControllerClient) CreateMemoryLog(log models.MemoryLog) error {
	requestURL := d.BaseURL + "/memorylogs"

	bodyData, err := json.Marshal(log)
	if err != nil {
		return fmt.Errorf("failed to encode request body: %w", err)
	}

	req, err := http.NewRequest("POST", requestURL, bytes.NewBuffer(bodyData))
	if err != nil {
		return err
	}

	//req.Header.Set("Authorization", "Token "+d.Token)

	resp, err := d.HTTPClient.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("failed to read error response: %w", err)
		}

		return fmt.Errorf("controller returned status code %d: %s", resp.StatusCode, string(body))
	}

	return nil
}
