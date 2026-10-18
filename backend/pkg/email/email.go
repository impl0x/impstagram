package email

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// email manager, which sends email using resend api and uses [Renderer] to render email data
type Manager struct {
	apiKey     string
	emailID    string
	httpClient *http.Client
	Renderer   *Renderer
}

func NewManager(apiKey, emailID string, httpClient *http.Client, renderer *Renderer) Manager {
	return Manager{
		apiKey:     apiKey,
		emailID:    emailID,
		httpClient: httpClient,
		Renderer:   renderer,
	}
}

type sendRequest struct {
	From    string   `json:"from"`
	To      []string `json:"to"`
	Subject string   `json:"subject"`
	HTML    string   `json:"html"`
}

func (m Manager) send(ctx context.Context, sub string, to []string, html HTML) error {
	req := sendRequest{
		From:    m.emailID,
		To:      to,
		Subject: sub,
		HTML:    string(html),
	}
	body, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("email: marshal email: %w", err) // should not happen
	}

	httpReq, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		"https://api.resend.com/emails",
		bytes.NewReader(body),
	)
	if err != nil {
		return fmt.Errorf("email: create request: %w", err) // should not happen
	}

	httpReq.Header.Set("Authorization", "Bearer "+m.apiKey)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := m.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("email: send email: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("email: resend returned unexpected status %s", resp.Status)
	}
	return nil
}

func (m Manager) SendSingle(ctx context.Context, sub string, to string, html HTML) error {
	return m.send(ctx, sub, []string{to}, html)
}

func (m Manager) SendMultiple(ctx context.Context, sub string, to []string, html HTML) error {
	return m.send(ctx, sub, to, html)
}
