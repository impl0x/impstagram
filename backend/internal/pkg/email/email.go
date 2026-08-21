package email

import (
	"backend/internal/config"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type Sender interface { // sender interface to mock http email client for testing
	Send(req SendRequest) error
}

type MockClient struct{}

func (mc MockClient) Send(req SendRequest) error {
	println("Mocking email send to " + req.To[0] + " from " + req.From)
	return nil
}

type Client struct {
	httpClient *http.Client
}

func NewClient(httpClient *http.Client) Client {
	return Client{
		httpClient: httpClient,
	}
}

type SendRequest struct {
	From    string   `json:"from"`
	To      []string `json:"to"`
	Subject string   `json:"subject"`
	HTML    string   `json:"html"`
}

func NewSendRequest(to, subject, html string) SendRequest {
	return SendRequest{config.EmailID, []string{to}, subject, html}
}

type SendResponse struct {
	ID string `json:"id"`
}

func (c Client) Send(req SendRequest) error {
	body, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("marshal email: %w", err)
	}

	httpReq, err := http.NewRequest(
		http.MethodPost,
		"https://api.resend.com/emails",
		bytes.NewReader(body),
	)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	httpReq.Header.Set("Authorization", "Bearer "+config.ResendApiKey)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("send email: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("resend returned status %s", resp.Status)
	}
	return nil
}
