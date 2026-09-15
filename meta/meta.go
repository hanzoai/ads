// Package meta implements the Meta Marketing API client (v21.0),
// inspired by yury-egorenkov/fb-ads.
package meta

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/hanzoai/ads"
)

const (
	DefaultBaseURL = "https://graph.facebook.com/v21.0"
)

// Client executes operations against the Meta Marketing API.
type Client struct {
	BaseURL    string
	HTTPClient *http.Client
}

// NewClient initializes a Meta Ads API client.
func NewClient() *Client {
	return &Client{
		BaseURL: DefaultBaseURL,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// MetaError represents a structured Graph API error envelope.
type MetaError struct {
	Message   string `json:"message"`
	Type      string `json:"type"`
	Code      int    `json:"code"`
	ErrorSubcode int `json:"error_subcode"`
}

// CreateCampaignParams holds parameters for creating a Meta campaign.
type CreateCampaignParams struct {
	AccountID           string
	Name                string
	Objective           string // OUTCOME_TRAFFIC, OUTCOME_SALES, OUTCOME_LEADS, etc.
	Status              string // ACTIVE, PAUSED
	SpecialAdCategories []string
	DailyBudgetCents    int64
	LifetimeBudgetCents int64
}

// CreateCampaign creates an ad campaign under act_<id>.
func (c *Client) CreateCampaign(ctx context.Context, token string, p CreateCampaignParams) (string, error) {
	act := NormalizeAccountID(p.AccountID)
	endpoint := fmt.Sprintf("%s/%s/campaigns", c.BaseURL, act)

	obj := p.Objective
	if obj == "" {
		obj = "OUTCOME_TRAFFIC"
	}
	status := p.Status
	if status == "" {
		status = "ACTIVE"
	}

	form := url.Values{
		"name":                  {p.Name},
		"objective":             {obj},
		"status":                {status},
		"special_ad_categories": {"[]"},
	}
	if p.DailyBudgetCents > 0 {
		form.Set("daily_budget", strconv.FormatInt(p.DailyBudgetCents, 10))
	}
	if p.LifetimeBudgetCents > 0 {
		form.Set("lifetime_budget", strconv.FormatInt(p.LifetimeBudgetCents, 10))
	}

	var res struct {
		ID    string     `json:"id"`
		Error *MetaError `json:"error"`
	}
	if err := c.postForm(ctx, endpoint, token, form, &res); err != nil {
		return "", err
	}
	if res.Error != nil {
		return "", fmt.Errorf("%w: meta api %d: %s", ads.ErrUpstream, res.Error.Code, res.Error.Message)
	}
	return res.ID, nil
}

// CreateAdSetParams defines ad set configuration and targeting.
type CreateAdSetParams struct {
	AccountID        string
	CampaignID       string
	Name             string
	OptimizationGoal string // LINK_CLICKS, OFFSITE_CONVERSIONS, IMPRESSIONS
	BillingEvent     string // IMPRESSIONS, LINK_CLICKS
	DailyBudgetCents int64
	BidAmountCents   int64
	Status           string
	Targeting        ads.Targeting
}

// CreateAdSet creates an ad set within a campaign.
func (c *Client) CreateAdSet(ctx context.Context, token string, p CreateAdSetParams) (string, error) {
	act := NormalizeAccountID(p.AccountID)
	endpoint := fmt.Sprintf("%s/%s/adsets", c.BaseURL, act)

	targetJSON, err := json.Marshal(map[string]any{
		"geo_locations": map[string]any{
			"countries": []string{"US"},
		},
		"age_min": p.Targeting.AgeMin,
		"age_max": p.Targeting.AgeMax,
	})
	if err != nil {
		return "", err
	}

	goal := p.OptimizationGoal
	if goal == "" {
		goal = "LINK_CLICKS"
	}
	billing := p.BillingEvent
	if billing == "" {
		billing = "IMPRESSIONS"
	}

	form := url.Values{
		"name":              {p.Name},
		"campaign_id":       {p.CampaignID},
		"optimization_goal": {goal},
		"billing_event":     {billing},
		"status":            {"ACTIVE"},
		"targeting":         {string(targetJSON)},
	}
	if p.DailyBudgetCents > 0 {
		form.Set("daily_budget", strconv.FormatInt(p.DailyBudgetCents, 10))
	}
	if p.BidAmountCents > 0 {
		form.Set("bid_amount", strconv.FormatInt(p.BidAmountCents, 10))
	}

	var res struct {
		ID    string     `json:"id"`
		Error *MetaError `json:"error"`
	}
	if err := c.postForm(ctx, endpoint, token, form, &res); err != nil {
		return "", err
	}
	if res.Error != nil {
		return "", fmt.Errorf("%w: meta api %d: %s", ads.ErrUpstream, res.Error.Code, res.Error.Message)
	}
	return res.ID, nil
}

// GetInsights fetches real-time campaign performance metrics from Meta.
func (c *Client) GetInsights(ctx context.Context, token, objectID, datePreset string) (*ads.Insights, error) {
	if datePreset == "" {
		datePreset = "last_30d"
	}
	fields := "spend,impressions,clicks,cpc,cpm,ctr,actions,action_values,purchase_roas"
	endpoint := fmt.Sprintf("%s/%s/insights?fields=%s&date_preset=%s", c.BaseURL, objectID, fields, datePreset)

	var res struct {
		Data []struct {
			Spend        string `json:"spend"`
			Impressions  string `json:"impressions"`
			Clicks       string `json:"clicks"`
			CTR          string `json:"ctr"`
			CPC          string `json:"cpc"`
			CPM          string `json:"cpm"`
			PurchaseROAS []struct {
				Value string `json:"value"`
			} `json:"purchase_roas"`
			Actions []struct {
				ActionType string `json:"action_type"`
				Value      string `json:"value"`
			} `json:"actions"`
		} `json:"data"`
		Error *MetaError `json:"error"`
	}

	if err := c.getJSON(ctx, endpoint, token, &res); err != nil {
		return nil, err
	}
	if res.Error != nil {
		return nil, fmt.Errorf("%w: meta error %d: %s", ads.ErrUpstream, res.Error.Code, res.Error.Message)
	}
	if len(res.Data) == 0 {
		return &ads.Insights{Platform: ads.PlatformMeta, CampaignID: objectID}, nil
	}

	row := res.Data[0]
	spendDollars, _ := strconv.ParseFloat(row.Spend, 64)
	impressions, _ := strconv.ParseInt(row.Impressions, 10, 64)
	clicks, _ := strconv.ParseInt(row.Clicks, 10, 64)
	ctr, _ := strconv.ParseFloat(row.CTR, 64)
	cpc, _ := strconv.ParseFloat(row.CPC, 64)
	cpm, _ := strconv.ParseFloat(row.CPM, 64)

	var roas float64
	if len(row.PurchaseROAS) > 0 {
		roas, _ = strconv.ParseFloat(row.PurchaseROAS[0].Value, 64)
	}

	var conversions int64
	for _, a := range row.Actions {
		if a.ActionType == "offsite_conversion.fb_pixel_purchase" || a.ActionType == "lead" {
			val, _ := strconv.ParseInt(a.Value, 10, 64)
			conversions += val
		}
	}

	return &ads.Insights{
		Platform:     ads.PlatformMeta,
		CampaignID:   objectID,
		DateStart:    datePreset,
		SpendCents:   int64(spendDollars*100 + 0.5),
		Impressions:  impressions,
		Clicks:       clicks,
		CTR:          ctr,
		CPC:          cpc,
		CPM:          cpm,
		Conversions:  conversions,
		ROAS:         roas,
	}, nil
}

// PauseCampaign sets status=PAUSED on a Meta campaign.
func (c *Client) PauseCampaign(ctx context.Context, token, campaignID string) error {
	endpoint := fmt.Sprintf("%s/%s", c.BaseURL, campaignID)
	form := url.Values{"status": {"PAUSED"}}

	var res struct {
		Success bool       `json:"success"`
		Error   *MetaError `json:"error"`
	}
	if err := c.postForm(ctx, endpoint, token, form, &res); err != nil {
		return err
	}
	if res.Error != nil {
		return fmt.Errorf("%w: meta pause: %s", ads.ErrUpstream, res.Error.Message)
	}
	return nil
}

func (c *Client) postForm(ctx context.Context, endpoint, token string, form url.Values, out any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	return c.do(req, out)
}

func (c *Client) getJSON(ctx context.Context, endpoint, token string, out any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	return c.do(req, out)
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
		return fmt.Errorf("%w: meta status %d: %s", ads.ErrUpstream, resp.StatusCode, string(b))
	}
	if out != nil {
		return json.NewDecoder(resp.Body).Decode(out)
	}
	return nil
}

// NormalizeAccountID formats an ad account ID with the "act_" prefix.
func NormalizeAccountID(account string) string {
	a := strings.TrimSpace(account)
	if a == "" {
		return ""
	}
	if strings.HasPrefix(a, "act_") {
		return a
	}
	return "act_" + a
}
