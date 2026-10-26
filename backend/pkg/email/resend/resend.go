// package implementing the resend api for email sending
// contains not much as of now just as much as i need in my project
package resend

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type Resend struct {
	httpClient *http.Client
	emailID    string
	apiKey     string
}

func New(h *http.Client, emailID, apiKey string) Resend {
	return Resend{
		h, emailID, apiKey,
	}
}

type sendRequest struct {
	From    string   `json:"from"`
	To      []string `json:"to"`
	Subject string   `json:"subject"`
	HTML    string   `json:"html"`
}

func (m Resend) send(ctx context.Context, sub string, to []string, html string) error {
	req := sendRequest{
		From:    m.emailID,
		To:      to,
		Subject: sub,
		HTML:    html,
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
		body, err := io.ReadAll(resp.Body) // only read body if api return error
		if err != nil {
			return fmt.Errorf("email: resend returned unexpected status - %s, body could not be read, %w", resp.Status, err)
		}
		return fmt.Errorf("email: resend returned unexpected status - %s, body - %s", resp.Status, string(body))
	}
	return nil
}

func (m Resend) SendSingle(ctx context.Context, sub string, to string, html string) error {
	return m.send(ctx, sub, []string{to}, html)
}

func (m Resend) SendMultiple(ctx context.Context, sub string, to []string, html string) error {
	return m.send(ctx, sub, to, html)
}
