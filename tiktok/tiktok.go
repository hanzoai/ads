// Package tiktok implements the TikTok Business Marketing API v1.3 client,
// inspired by hung12ct/go-tiktok-business-sdk.
package tiktok

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
	DefaultBaseURL = "https://business-api.tiktok.com/open_api/v1.3"
)

// Client executes operations on the TikTok Marketing API.
type Client struct {
	BaseURL    string
	HTTPClient *http.Client
}

// NewClient initializes a TikTok Business API client.
func NewClient() *Client {
	return &Client{
		BaseURL: DefaultBaseURL,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// CreateCampaignParams defines options for TikTok campaign creation.
type CreateCampaignParams struct {
	AdvertiserID     string
	CampaignName     string
	ObjectiveType    string // TRAFFIC, CONVERSIONS, LEAD_GENERATION, REACH
	BudgetMode       string // BUDGET_MODE_DAY, BUDGET_MODE_TOTAL, BUDGET_MODE_INFINITE
	BudgetCents      int64
	OperationStatus  string // ENABLE, DISABLE
}

// CreateCampaign creates a campaign on TikTok Business.
func (c *Client) CreateCampaign(ctx context.Context, token string, p CreateCampaignParams) (string, error) {
	endpoint := fmt.Sprintf("%s/campaign/create/", c.BaseURL)

	budgetMode := p.BudgetMode
	if budgetMode == "" {
		budgetMode = "BUDGET_MODE_DAY"
	}
	obj := p.ObjectiveType
	if obj == "" {
		obj = "TRAFFIC"
	}

	payload := map[string]any{
		"advertiser_id":    p.AdvertiserID,
		"campaign_name":    p.CampaignName,
		"objective_type":   obj,
		"budget_mode":      budgetMode,
		"operation_status": "ENABLE",
	}
	if p.BudgetCents > 0 {
		payload["budget"] = float64(p.BudgetCents) / 100.0
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(data))
	if err != nil {
		return "", err
	}
	req.Header.Set("Access-Token", token)
	req.Header.Set("Content-Type", "application/json")

	var res struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Data    struct {
			CampaignID string `json:"campaign_id"`
		} `json:"data"`
	}

	if err := c.do(req, &res); err != nil {
		return "", err
	}
	if res.Code != 0 {
		return "", fmt.Errorf("%w: tiktok api error %d: %s", ads.ErrUpstream, res.Code, res.Message)
	}
	return res.Data.CampaignID, nil
}

// GetSpend retrieves spend data for a campaign.
func (c *Client) GetSpend(ctx context.Context, token, advertiserID, campaignID string) (int64, error) {
	endpoint := fmt.Sprintf("%s/report/integrated/get/?advertiser_id=%s&report_type=BASIC&data_level=AUCTION_CAMPAIGN&dimensions=[\"campaign_id\"]&metrics=[\"spend\"]&filters=[{\"field_name\":\"campaign_id\",\"filter_type\":\"IN\",\"filter_value\":\"[\\\"%s\\\"]\"}]",
		c.BaseURL, advertiserID, campaignID)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return 0, err
	}
	req.Header.Set("Access-Token", token)

	var res struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Data    struct {
			List []struct {
				Metrics struct {
					Spend string `json:"spend"`
				} `json:"metrics"`
			} `json:"list"`
		} `json:"data"`
	}

	if err := c.do(req, &res); err != nil {
		return 0, err
	}
	if res.Code != 0 || len(res.Data.List) == 0 {
		return 0, nil
	}

	var spend float64
	fmt.Sscanf(res.Data.List[0].Metrics.Spend, "%f", &spend)
	return int64(spend*100 + 0.5), nil
}

// PauseCampaign updates campaign status to DISABLE.
func (c *Client) PauseCampaign(ctx context.Context, token, advertiserID, campaignID string) error {
	endpoint := fmt.Sprintf("%s/campaign/status/update/", c.BaseURL)

	payload := map[string]any{
		"advertiser_id":   advertiserID,
		"campaign_ids":    []string{campaignID},
		"operation_status": "DISABLE",
	}
	data, _ := json.Marshal(payload)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(data))
	if err != nil {
		return err
	}
	req.Header.Set("Access-Token", token)
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
		return fmt.Errorf("%w: tiktok status %d: %s", ads.ErrUpstream, resp.StatusCode, string(b))
	}
	if out != nil {
		return json.NewDecoder(resp.Body).Decode(out)
	}
	return nil
}
