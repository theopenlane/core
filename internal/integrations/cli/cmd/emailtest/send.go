package emailtest

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/samber/lo"
	"github.com/spf13/cobra"

	"github.com/theopenlane/core/internal/integrations/cli/cmd"
)

// emailTestSendRequest mirrors handlers.EmailTestSendRequest
type emailTestSendRequest struct {
	To   string `json:"to"`
	Name string `json:"name,omitempty"`
}

// emailTestSendResult mirrors handlers.EmailTestSendResult
type emailTestSendResult struct {
	Name   string `json:"name"`
	Status string `json:"status"`
	Error  string `json:"error,omitempty"`
}

// emailTestSendResponse mirrors handlers.EmailTestSendResponse
type emailTestSendResponse struct {
	Results []emailTestSendResult `json:"results"`
}

var sendCmd = &cobra.Command{
	Use:   "send",
	Short: "send a single test email by dispatcher name",
	Long: `Sends a test email through the running server's email-test endpoint using
scaffolded fixture data. Requires --name and --to.

The server must be running in dev mode with integrations enabled.`,
	RunE: func(c *cobra.Command, _ []string) error {
		name := cmd.Config.String("name")
		if name == "" {
			return ErrNameRequired
		}

		return sendTestEmail(c.Context(), cmd.Config.String("to"), name)
	},
}

var sendAllCmd = &cobra.Command{
	Use:   "send-all",
	Short: "send test emails for all registered dispatchers",
	Long: `Sends a test email for every registered dispatcher through the running
server's email-test endpoint using scaffolded fixture data. Requires --to.

The server must be running in dev mode with integrations enabled.`,
	RunE: func(c *cobra.Command, _ []string) error {
		return sendTestEmail(c.Context(), cmd.Config.String("to"), "")
	},
}

func init() {
	command.AddCommand(sendCmd)
	command.AddCommand(sendAllCmd)

	sendCmd.Flags().String("to", "", "recipient email address (required)")
	sendCmd.Flags().String("name", "", "dispatcher name to send (required)")

	sendAllCmd.Flags().String("to", "", "recipient email address (required)")
}

// sendTestEmail calls the server's email-test/send endpoint.
// When name is empty, all dispatchers are sent
func sendTestEmail(ctx context.Context, toEmail, name string) error {
	if toEmail == "" {
		return ErrRecipientRequired
	}

	host := cmd.Config.String("openlane.host")
	if host == "" {
		return ErrHostRequired
	}

	body, err := json.Marshal(emailTestSendRequest{
		To:   toEmail,
		Name: name,
	})
	if err != nil {
		return fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, host+"/email-test/send", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}

	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned %d: %s", resp.StatusCode, string(respBody))
	}

	var result emailTestSendResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return fmt.Errorf("parse response: %w", err)
	}

	headers := []string{"Dispatcher", "Status", "Error"}
	rows := lo.Map(result.Results, func(r emailTestSendResult, _ int) []string {
		return []string{r.Name, r.Status, r.Error}
	})

	return cmd.RenderTable(result, headers, rows)
}
