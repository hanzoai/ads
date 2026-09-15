// Package ga4 implements the Google Analytics 4 (GA4) Data API client
// and omnichannel cross-platform attribution engine.
package ga4

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/hanzoai/ads"
)

const (
	DefaultBaseURL = "https://analyticsdata.googleapis.com/v1beta"
)

// Client executes reports on the GA4 Data API.
type Client struct {
	BaseURL    string
	HTTPClient *http.Client
}

// NewClient initializes a Google Analytics 4 Data API client.
func NewClient() *Client {
	return &Client{
		BaseURL: DefaultBaseURL,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// RunReportParams defines date ranges, metrics, and dimensions for standard reporting.
type RunReportParams struct {
	PropertyID string
	StartDate  string
	EndDate    string
}

// RunReport fetches sessions, users, pageviews, and conversion revenue from GA4.
func (c *Client) RunReport(ctx context.Context, token string, p RunReportParams) (*ads.GA4Report, error) {
	propertyID := cleanPropertyID(p.PropertyID)
	endpoint := fmt.Sprintf("%s/properties/%s:runReport", c.BaseURL, propertyID)

	start := p.StartDate
	if start == "" {
		start = "30daysAgo"
	}
	end := p.EndDate
	if end == "" {
		end = "yesterday"
	}

	payload := map[string]any{
		"dateRanges": []map[string]string{
			{"startDate": start, "endDate": end},
		},
		"dimensions": []map[string]string{
			{"name": "sessionDefaultChannelGrouping"},
		},
		"metrics": []map[string]string{
			{"name": "activeUsers"},
			{"name": "sessions"},
			{"name": "screenPageViews"},
			{"name": "conversions"},
			{"name": "purchaseRevenue"},
			{"name": "bounceRate"},
		},
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	var res struct {
		Rows []struct {
			DimensionValues []struct {
				Value string `json:"value"`
			} `json:"dimensionValues"`
			MetricValues []struct {
				Value string `json:"value"`
			} `json:"metricValues"`
		} `json:"rows"`
		Totals []struct {
			MetricValues []struct {
				Value string `json:"value"`
			} `json:"metricValues"`
		} `json:"totals"`
	}

	if err := c.do(req, &res); err != nil {
		return nil, err
	}

	report := &ads.GA4Report{
		PropertyID: propertyID,
		DateRange:  fmt.Sprintf("%s..%s", start, end),
	}

	if len(res.Totals) > 0 && len(res.Totals[0].MetricValues) >= 6 {
		vals := res.Totals[0].MetricValues
		report.ActiveUsers, _ = strconv.ParseInt(vals[0].Value, 10, 64)
		report.Sessions, _ = strconv.ParseInt(vals[1].Value, 10, 64)
		report.ScreenPageViews, _ = strconv.ParseInt(vals[2].Value, 10, 64)
		report.Conversions, _ = strconv.ParseInt(vals[3].Value, 10, 64)
		report.PurchaseRevenue, _ = strconv.ParseFloat(vals[4].Value, 64)
		report.BounceRate, _ = strconv.ParseFloat(vals[5].Value, 64)
	}

	for _, row := range res.Rows {
		if len(row.DimensionValues) > 0 && len(row.MetricValues) >= 5 {
			channel := row.DimensionValues[0].Value
			sessions, _ := strconv.ParseInt(row.MetricValues[1].Value, 10, 64)
			conversions, _ := strconv.ParseInt(row.MetricValues[3].Value, 10, 64)
			revenue, _ := strconv.ParseFloat(row.MetricValues[4].Value, 64)

			report.TopChannels = append(report.TopChannels, ads.ChannelShare{
				ChannelName: channel,
				Sessions:    sessions,
				Conversions: conversions,
				Revenue:     revenue,
			})
		}
	}

	return report, nil
}

// RunRealtimeReport fetches active users within the last 30 minutes.
func (c *Client) RunRealtimeReport(ctx context.Context, token, propertyID string) (int64, error) {
	pid := cleanPropertyID(propertyID)
	endpoint := fmt.Sprintf("%s/properties/%s:runRealtimeReport", c.BaseURL, pid)

	payload := map[string]any{
		"metrics": []map[string]string{
			{"name": "activeUsers"},
		},
	}
	data, _ := json.Marshal(payload)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(data))
	if err != nil {
		return 0, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	var res struct {
		Rows []struct {
			MetricValues []struct {
				Value string `json:"value"`
			} `json:"metricValues"`
		} `json:"rows"`
	}

	if err := c.do(req, &res); err != nil {
		return 0, err
	}

	if len(res.Rows) > 0 && len(res.Rows[0].MetricValues) > 0 {
		active, _ := strconv.ParseInt(res.Rows[0].MetricValues[0].Value, 10, 64)
		return active, nil
	}
	return 0, nil
}

// SynthesizeCrossPlatformROAS correlates ad network spend with GA4 conversion performance.
func SynthesizeCrossPlatformROAS(dateRange string, networkInsights []ads.Insights, ga4 *ads.GA4Report) *ads.CrossPlatformROAS {
	result := &ads.CrossPlatformROAS{
		DateRange:             dateRange,
		Platforms:             make(map[string]ads.PlatformROAS),
		RecommendedAllocation: make(map[string]float64),
	}

	var totalSpend int64
	var totalRevenue int64
	var totalConversions int64
	bestROAS := -1.0
	bestPlatform := ""

	for _, ins := range networkInsights {
		totalSpend += ins.SpendCents
		totalConversions += ins.Conversions

		rev := ins.RevenueCents
		if rev == 0 && ins.ROAS > 0 && ins.SpendCents > 0 {
			rev = int64(float64(ins.SpendCents) * ins.ROAS)
		}
		totalRevenue += rev

		var roas float64
		if ins.SpendCents > 0 {
			roas = float64(rev) / float64(ins.SpendCents)
		}
		var cpa float64
		if ins.Conversions > 0 {
			cpa = float64(ins.SpendCents) / float64(ins.Conversions*100)
		}

		result.Platforms[ins.Platform] = ads.PlatformROAS{
			Platform:    ins.Platform,
			SpendCents:  ins.SpendCents,
			Clicks:      ins.Clicks,
			Conversions: ins.Conversions,
			ROAS:        roas,
			CPA:         cpa,
		}

		if roas > bestROAS {
			bestROAS = roas
			bestPlatform = ins.Platform
		}
	}

	if ga4 != nil && ga4.PurchaseRevenue > 0 && totalRevenue == 0 {
		totalRevenue = int64(ga4.PurchaseRevenue * 100)
	}

	result.TotalSpendCents = totalSpend
	result.TotalRevenueCents = totalRevenue
	result.TopPlatform = bestPlatform

	if totalSpend > 0 {
		result.BlendedROAS = float64(totalRevenue) / float64(totalSpend)
	}
	if totalConversions > 0 {
		result.BlendedCPA = float64(totalSpend) / float64(totalConversions*100)
	}

	// Calculate algorithmic allocation based on normalized ROAS weight
	var roasSum float64
	for _, p := range result.Platforms {
		if p.ROAS > 0 {
			roasSum += p.ROAS
		}
	}
	if roasSum > 0 {
		for name, p := range result.Platforms {
			if p.ROAS > 0 {
				result.RecommendedAllocation[name] = float64(int((p.ROAS/roasSum)*10000)) / 100.0
			} else {
				result.RecommendedAllocation[name] = 5.0
			}
		}
	}

	result.OptimizationSummary = fmt.Sprintf(
		"Blended ROAS is %.2fx across %d networks. %s leads performance with %.2fx ROAS. Reallocate +15%% budget to %s for highest marginal return.",
		result.BlendedROAS, len(result.Platforms), strings.ToUpper(bestPlatform), bestROAS, strings.ToUpper(bestPlatform),
	)

	return result
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
		return fmt.Errorf("%w: ga4 status %d: %s", ads.ErrUpstream, resp.StatusCode, string(b))
	}
	if out != nil {
		return json.NewDecoder(resp.Body).Decode(out)
	}
	return nil
}

func cleanPropertyID(id string) string {
	id = strings.TrimSpace(id)
	return strings.TrimPrefix(id, "properties/")
}
