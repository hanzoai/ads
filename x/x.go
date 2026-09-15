// Package x implements the X (Twitter) Ads API v12 client,
// inspired by gxbvc/x-ads-cli.
package x

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/hanzoai/ads"
)

const (
	DefaultBaseURL = "https://ads-api.x.com/12"
)

// Client interacts with the X Ads REST API.
type Client struct {
	BaseURL    string
	HTTPClient *http.Client
}

// NewClient creates an X Ads API client.
func NewClient() *Client {
	return &Client{
		BaseURL: DefaultBaseURL,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// CreateCampaignParams holds fields to create a campaign on X Ads.
type CreateCampaignParams struct {
	AccountID           string
	Name                string
	Objective           string // WEBSITE_CLICKS, REACH, ENGAGEMENTS, APP_INSTALLS
	DailyBudgetCents    int64
	TotalBudgetCents    int64
	EntityStatus        string // ACTIVE, PAUSED
	StartTime           time.Time
}

// CreateCampaign creates a campaign under the specified X Ads account.
func (c *Client) CreateCampaign(ctx context.Context, token string, p CreateCampaignParams) (string, error) {
	endpoint := fmt.Sprintf("%s/accounts/%s/campaigns", c.BaseURL, p.AccountID)

	status := p.EntityStatus
	if status == "" {
		status = "ACTIVE"
	}
	obj := p.Objective
	if obj == "" {
		obj = "WEBSITE_CLICKS"
	}

	payload := map[string]any{
		"name":                p.Name,
		"funding_instrument_id": "",
		"daily_budget_amount_local_micro": p.DailyBudgetCents * 10000,
		"entity_status":       status,
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(data))
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	var res struct {
		Data struct {
			ID string `json:"id"`
		} `json:"data"`
		Errors []struct {
			Message string `json:"message"`
			Code    string `json:"code"`
		} `json:"errors"`
	}

	if err := c.do(req, &res); err != nil {
		return "", err
	}
	if len(res.Errors) > 0 {
		return "", fmt.Errorf("%w: x ads error: %s", ads.ErrUpstream, res.Errors[0].Message)
	}
	return res.Data.ID, nil
}

// GetSpend retrieves spend data for an account or campaign.
func (c *Client) GetSpend(ctx context.Context, token, accountID, campaignID string) (int64, error) {
	endpoint := fmt.Sprintf("%s/stats/accounts/%s?entity=CAMPAIGN&entity_ids=%s&metric_groups=BILLING", c.BaseURL, accountID, campaignID)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return 0, err
	}
	req.Header.Set("Authorization", "Bearer "+token)

	var res struct {
		Data []struct {
			IDString string `json:"id"`
			Metrics  struct {
				BilledChargeLocalMicro []int64 `json:"billed_charge_local_micro"`
			} `json:"metrics"`
		} `json:"data"`
	}

	if err := c.do(req, &res); err != nil {
		return 0, err
	}
	if len(res.Data) == 0 || len(res.Data[0].Metrics.BilledChargeLocalMicro) == 0 {
		return 0, nil
	}
	micros := res.Data[0].Metrics.BilledChargeLocalMicro[0]
	return micros / 10000, nil
}

// PauseCampaign sets entity_status=PAUSED on an X campaign.
func (c *Client) PauseCampaign(ctx context.Context, token, accountID, campaignID string) error {
	endpoint := fmt.Sprintf("%s/accounts/%s/campaigns/%s", c.BaseURL, accountID, campaignID)

	payload := map[string]string{"entity_status": "PAUSED"}
	data, _ := json.Marshal(payload)

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, endpoint, bytes.NewReader(data))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	return c.do(req, nil)
}

func (c *Client) do(req *http.Request, out any) error {
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("%w: %v", ads.ErrUpstream, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		return ads.ErrNotConnected
	}
	if resp.StatusCode >= 300 {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return fmt.Errorf("%w: x status %d: %s", ads.ErrUpstream, resp.StatusCode, string(b))
	}
	if out != nil {
		return json.NewDecoder(resp.Body).Decode(out)
	}
	return nil
}
