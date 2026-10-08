package email

import (
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
	serviceName string
	apiKey      string
	emailID     string
	httpClient  *http.Client
}

func NewClient(apiKey, serviceName, emailID string, httpClient *http.Client) Client {
	return Client{
		serviceName: serviceName,
		apiKey:      apiKey,
		emailID:     emailID,
		httpClient:  httpClient,
	}
}

type SendRequest struct {
	From    string   `json:"from"`
	To      []string `json:"to"`
	Subject string   `json:"subject"`
	HTML    string   `json:"html"`
}

func NewSendRequest(to, subject, html string) SendRequest {
	return SendRequest{
		To:      []string{to},
		Subject: subject,
		HTML:    html,
	}
}

type SendResponse struct {
	ID string `json:"id"`
}

func (c Client) Send(req SendRequest) error {
	req.From = c.emailID
	body, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("email: marshal email: %w", err)
	}

	httpReq, err := http.NewRequest(
		http.MethodPost,
		"https://api.resend.com/emails",
		bytes.NewReader(body),
	)
	if err != nil {
		return fmt.Errorf("email: create request: %w", err)
	}

	httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("email: send email: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("email: resend returned unexpected status %s", resp.Status)
	}
	return nil
}
