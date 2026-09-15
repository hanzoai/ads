// Package google implements the Google Ads API v18 client and GAQL executor,
// inspired by kritzware/google-ads-go.
package google

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/hanzoai/ads"
)

const (
	DefaultBaseURL = "https://googleads.googleapis.com/v18"
)

// Client executes operations against the Google Ads REST API.
type Client struct {
	BaseURL         string
	HTTPClient      *http.Client
	DeveloperToken  string
	DefaultCustomer string
}

// NewClient initializes a Google Ads client with sane timeouts.
func NewClient(devToken string) *Client {
	return &Client{
		BaseURL:        DefaultBaseURL,
		DeveloperToken: devToken,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// CampaignMutateRequest represents a Google Ads campaign creation/mutation payload.
type CampaignMutateRequest struct {
	Operations []CampaignOperation `json:"operations"`
}

// CampaignOperation specifies an individual create, update, or remove operation.
type CampaignOperation struct {
	Create *GoogleCampaign `json:"create,omitempty"`
	Update *GoogleCampaign `json:"update,omitempty"`
	Remove string          `json:"remove,omitempty"`
}

// GoogleCampaign defines the Google Ads campaign resource.
type GoogleCampaign struct {
	ResourceName       string          `json:"resourceName,omitempty"`
	ID                 string          `json:"id,omitempty"`
	Name               string          `json:"name"`
	Status             string          `json:"status"` // ENABLED, PAUSED, REMOVED
	AdvertisingChannel string          `json:"advertisingChannelType"`
	CampaignBudget     string          `json:"campaignBudget,omitempty"`
	ManualCPC          *ManualCPCBid   `json:"manualCpc,omitempty"`
	TargetCPA          *TargetCPABid   `json:"targetCpa,omitempty"`
	TargetROAS         *TargetROASBid  `json:"targetRoas,omitempty"`
	NetworkSettings    *NetworkSetting `json:"networkSettings,omitempty"`
}

type ManualCPCBid struct {
	EnhancedCPCEnabled bool `json:"enhancedCpcEnabled"`
}

type TargetCPABid struct {
	TargetCPA int64 `json:"targetCpaMicros"`
}

type TargetROASBid struct {
	TargetROAS float64 `json:"targetRoas"`
}

type NetworkSetting struct {
	TargetGoogleSearch         bool `json:"targetGoogleSearch"`
	TargetSearchNetwork        bool `json:"targetSearchNetwork"`
	TargetContentNetwork       bool `json:"targetContentNetwork"`
	TargetPartnerSearchNetwork bool `json:"targetPartnerSearchNetwork"`
}

// MutateCampaigns creates, updates, or deletes campaigns for a customer.
func (c *Client) MutateCampaigns(ctx context.Context, token, customerID string, ops []CampaignOperation) ([]string, error) {
	cid := cleanCustomerID(customerID)
	endpoint := fmt.Sprintf("%s/customers/%s/campaigns:mutate", c.BaseURL, cid)

	reqBody := CampaignMutateRequest{Operations: ops}
	data, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal mutate: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	c.setHeaders(req, token, cid)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ads.ErrUpstream, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		return nil, ads.ErrNotConnected
	}
	if resp.StatusCode >= 300 {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return nil, fmt.Errorf("%w: google ads api error %d: %s", ads.ErrUpstream, resp.StatusCode, string(b))
	}

	var result struct {
		Results []struct {
			ResourceName string `json:"resourceName"`
		} `json:"results"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	var resourceNames []string
	for _, r := range result.Results {
		resourceNames = append(resourceNames, r.ResourceName)
	}
	return resourceNames, nil
}

// GenerateKeywordIdeas calls the Keyword Planner service to generate keyword volume and CPC estimates.
func (c *Client) GenerateKeywordIdeas(ctx context.Context, token, customerID string, seedKeywords []string, pageURL string) ([]ads.Keyword, error) {
	cid := cleanCustomerID(customerID)
	endpoint := fmt.Sprintf("%s/customers/%s:generateKeywordIdeas", c.BaseURL, cid)

	payload := map[string]any{
		"geoTargetConstants": []string{"geoTargetConstants/2840"}, // USA default
		"keywordPlanNetwork": "GOOGLE_SEARCH",
	}
	if len(seedKeywords) > 0 {
		payload["keywordSeed"] = map[string]any{"keywords": seedKeywords}
	}
	if pageURL != "" {
		payload["urlSeed"] = map[string]any{"url": pageURL}
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	c.setHeaders(req, token, cid)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ads.ErrUpstream, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		return nil, ads.ErrNotConnected
	}
	if resp.StatusCode >= 300 {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return nil, fmt.Errorf("%w: %d: %s", ads.ErrUpstream, resp.StatusCode, string(b))
	}

	var out struct {
		Results []struct {
			Text             string `json:"text"`
			KeywordIdeaMetrics struct {
				AvgMonthlySearches string `json:"avgMonthlySearches"`
				Competition        string `json:"competition"`
				CompetitionIndex   int    `json:"competitionIndex"`
				LowTopOfPageBid    string `json:"lowTopOfPageBidMicros"`
				HighTopOfPageBid   string `json:"highTopOfPageBidMicros"`
			} `json:"keywordIdeaMetrics"`
		} `json:"results"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}

	var results []ads.Keyword
	for _, item := range out.Results {
		var searches int64
		var lowBidMicros, highBidMicros int64
		fmt.Sscanf(item.KeywordIdeaMetrics.AvgMonthlySearches, "%d", &searches)
		fmt.Sscanf(item.KeywordIdeaMetrics.LowTopOfPageBid, "%d", &lowBidMicros)
		fmt.Sscanf(item.KeywordIdeaMetrics.HighTopOfPageBid, "%d", &highBidMicros)

		results = append(results, ads.Keyword{
			Text:                item.Text,
			MatchType:           "PHRASE",
			MonthlySearches:     searches,
			Competition:         item.KeywordIdeaMetrics.Competition,
			CompetitionIndex:    item.KeywordIdeaMetrics.CompetitionIndex,
			LowTopPageBidCents:  lowBidMicros / 10000,
			HighTopPageBidCents: highBidMicros / 10000,
			SuggestedBidCents:   (lowBidMicros + highBidMicros) / 20000,
		})
	}
	return results, nil
}

// SearchGAQL executes a Google Ads Query Language query (GAQL) against Google Ads searchStream.
func (c *Client) SearchGAQL(ctx context.Context, token, customerID, query string) ([]map[string]any, error) {
	cid := cleanCustomerID(customerID)
	endpoint := fmt.Sprintf("%s/customers/%s/googleAds:searchStream", c.BaseURL, cid)

	body, err := json.Marshal(map[string]string{"query": query})
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	c.setHeaders(req, token, cid)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ads.ErrUpstream, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		return nil, ads.ErrNotConnected
	}
	if resp.StatusCode >= 300 {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return nil, fmt.Errorf("%w: %d: %s", ads.ErrUpstream, resp.StatusCode, string(b))
	}

	var streamRows []struct {
		Results []map[string]any `json:"results"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&streamRows); err != nil {
		return nil, err
	}

	var allRows []map[string]any
	for _, batch := range streamRows {
		allRows = append(allRows, batch.Results...)
	}
	return allRows, nil
}

func (c *Client) setHeaders(req *http.Request, token, customerID string) {
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	if c.DeveloperToken != "" {
		req.Header.Set("developer-token", c.DeveloperToken)
	}
	if customerID != "" {
		req.Header.Set("login-customer-id", customerID)
	}
}

func cleanCustomerID(id string) string {
	return strings.ReplaceAll(strings.TrimSpace(id), "-", "")
}
