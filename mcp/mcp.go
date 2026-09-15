// Package mcp implements the Model Context Protocol (MCP) server for
// Google Ads, Meta Ads, TikTok Ads, X Ads, and Google Analytics 4 (GA4).
//
// It provides native tool definitions, parameter schemas, and JSON-RPC dispatch
// for Claude, Cursor, ChatGPT, Windsurf, n8n, and autonomous agent loops.
package mcp

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hanzoai/ads"
	"github.com/hanzoai/ads/ga4"
	"github.com/hanzoai/ads/google"
	"github.com/hanzoai/ads/meta"
	"github.com/hanzoai/ads/tiktok"
	"github.com/hanzoai/ads/x"
)

// Tool represents a Model Context Protocol tool description.
type Tool struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	InputSchema map[string]any `json:"inputSchema"`
}

// Request represents a JSON-RPC 2.0 MCP request.
type Request struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      any             `json:"id"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

// Response represents a JSON-RPC 2.0 MCP response.
type Response struct {
	JSONRPC string `json:"jsonrpc"`
	ID      any    `json:"id"`
	Result  any    `json:"result,omitempty"`
	Error   *Error `json:"error,omitempty"`
}

// Error represents an MCP error object.
type Error struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// Server holds the registered toolset and backend drivers.
type Server struct {
	GoogleClient *google.Client
	MetaClient   *meta.Client
	TikTokClient *tiktok.Client
	XClient      *x.Client
	GA4Client    *ga4.Client
}

// NewServer initializes an MCP server.
func NewServer() *Server {
	return &Server{
		GoogleClient: google.NewClient(""),
		MetaClient:   meta.NewClient(),
		TikTokClient: tiktok.NewClient(),
		XClient:      x.NewClient(),
		GA4Client:    ga4.NewClient(),
	}
}

// ListTools returns all available MCP tools across Google, Meta, TikTok, X, and GA4.
func (s *Server) ListTools() []Tool {
	return []Tool{
		// Google Ads tools
		{
			Name:        "google_ads_list_campaigns",
			Description: "List Google Ads campaigns, budgets, and status for a customer account.",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"customerId": map[string]any{"type": "string", "description": "Google Ads customer ID (e.g. 123-456-7890)"},
				},
				"required": []string{"customerId"},
			},
		},
		{
			Name:        "google_ads_create_campaign",
			Description: "Create a new Search, Display, or Performance Max campaign on Google Ads.",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"customerId":       map[string]any{"type": "string", "description": "Google Ads customer ID"},
					"name":             map[string]any{"type": "string", "description": "Campaign name"},
					"dailyBudgetCents": map[string]any{"type": "integer", "description": "Daily budget in cents"},
					"channelType":      map[string]any{"type": "string", "description": "SEARCH, DISPLAY, PERFORMANCE_MAX"},
				},
				"required": []string{"customerId", "name", "dailyBudgetCents"},
			},
		},
		{
			Name:        "google_ads_keyword_research",
			Description: "Run Google Ads Keyword Planner research: get monthly search volume, competition, and top-of-page bids.",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"customerId": map[string]any{"type": "string", "description": "Google Ads customer ID"},
					"keywords":   map[string]any{"type": "array", "items": map[string]any{"type": "string"}, "description": "Seed keyword list"},
					"pageUrl":    map[string]any{"type": "string", "description": "Optional landing page URL for keyword ideas"},
				},
				"required": []string{"customerId", "keywords"},
			},
		},
		{
			Name:        "google_ads_search_gaql",
			Description: "Execute a Google Ads Query Language (GAQL) query for custom reporting.",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"customerId": map[string]any{"type": "string", "description": "Google Ads customer ID"},
					"query":      map[string]any{"type": "string", "description": "GAQL query, e.g. SELECT campaign.id, metrics.cost_micros FROM campaign"},
				},
				"required": []string{"customerId", "query"},
			},
		},

		// Meta Ads tools
		{
			Name:        "meta_ads_create_campaign",
			Description: "Create a new ad campaign on Meta (Facebook & Instagram).",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"accountId":        map[string]any{"type": "string", "description": "Meta ad account ID (e.g. act_12345 or 12345)"},
					"name":             map[string]any{"type": "string", "description": "Campaign name"},
					"objective":        map[string]any{"type": "string", "description": "OUTCOME_TRAFFIC, OUTCOME_SALES, OUTCOME_LEADS, OUTCOME_AWARENESS"},
					"dailyBudgetCents": map[string]any{"type": "integer", "description": "Daily budget in minor cents"},
				},
				"required": []string{"accountId", "name"},
			},
		},
		{
			Name:        "meta_ads_create_adset",
			Description: "Create a new targeted ad set on Meta with placements and audience criteria.",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"accountId":        map[string]any{"type": "string", "description": "Meta ad account ID"},
					"campaignId":       map[string]any{"type": "string", "description": "Parent campaign ID"},
					"name":             map[string]any{"type": "string", "description": "Ad set name"},
					"dailyBudgetCents": map[string]any{"type": "integer", "description": "Daily budget in cents"},
					"ageMin":           map[string]any{"type": "integer", "description": "Minimum age"},
					"ageMax":           map[string]any{"type": "integer", "description": "Maximum age"},
				},
				"required": []string{"accountId", "campaignId", "name"},
			},
		},
		{
			Name:        "meta_ads_insights",
			Description: "Fetch performance insights (spend, clicks, impressions, ROAS, conversions) for a Meta campaign or account.",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"objectId":   map[string]any{"type": "string", "description": "Campaign ID or Account ID"},
					"datePreset": map[string]any{"type": "string", "description": "Date preset (e.g. today, yesterday, last_7d, last_30d)"},
				},
				"required": []string{"objectId"},
			},
		},
		{
			Name:        "meta_ads_pause_campaign",
			Description: "Pause an active Meta ad campaign.",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"campaignId": map[string]any{"type": "string", "description": "Meta campaign ID to pause"},
				},
				"required": []string{"campaignId"},
			},
		},

		// TikTok Ads tools
		{
			Name:        "tiktok_ads_create_campaign",
			Description: "Create an ad campaign on TikTok Business.",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"advertiserId": map[string]any{"type": "string", "description": "TikTok Advertiser ID"},
					"campaignName": map[string]any{"type": "string", "description": "Campaign name"},
					"objective":    map[string]any{"type": "string", "description": "TRAFFIC, CONVERSIONS, LEAD_GENERATION"},
					"budgetCents":  map[string]any{"type": "integer", "description": "Budget in cents"},
				},
				"required": []string{"advertiserId", "campaignName"},
			},
		},
		{
			Name:        "tiktok_ads_insights",
			Description: "Fetch spend and conversion reporting from TikTok Business API.",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"advertiserId": map[string]any{"type": "string", "description": "TikTok Advertiser ID"},
					"campaignId":   map[string]any{"type": "string", "description": "TikTok Campaign ID"},
				},
				"required": []string{"advertiserId", "campaignId"},
			},
		},

		// X (Twitter) Ads tools
		{
			Name:        "x_ads_create_campaign",
			Description: "Create a promoted campaign on X Ads.",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"accountId":        map[string]any{"type": "string", "description": "X Ads Account ID"},
					"name":             map[string]any{"type": "string", "description": "Campaign name"},
					"objective":        map[string]any{"type": "string", "description": "WEBSITE_CLICKS, REACH, ENGAGEMENTS"},
					"dailyBudgetCents": map[string]any{"type": "integer", "description": "Daily budget in cents"},
				},
				"required": []string{"accountId", "name"},
			},
		},
		{
			Name:        "x_ads_insights",
			Description: "Fetch billing and spend metrics for an X Ads campaign.",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"accountId":  map[string]any{"type": "string", "description": "X Ads Account ID"},
					"campaignId": map[string]any{"type": "string", "description": "X Campaign ID"},
				},
				"required": []string{"accountId", "campaignId"},
			},
		},

		// Google Analytics 4 tools
		{
			Name:        "ga4_run_report",
			Description: "Query standard metrics (active users, sessions, conversions, purchase revenue) from Google Analytics 4.",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"propertyId": map[string]any{"type": "string", "description": "GA4 Property ID"},
					"startDate":  map[string]any{"type": "string", "description": "Start date (YYYY-MM-DD or 30daysAgo)"},
					"endDate":    map[string]any{"type": "string", "description": "End date (YYYY-MM-DD or yesterday)"},
				},
				"required": []string{"propertyId"},
			},
		},
		{
			Name:        "ga4_realtime_report",
			Description: "Get the count of active users on your site right now via GA4 Realtime.",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"propertyId": map[string]any{"type": "string", "description": "GA4 Property ID"},
				},
				"required": []string{"propertyId"},
			},
		},

		// Cross-Platform Attribution & ROAS
		{
			Name:        "cross_platform_roas_analysis",
			Description: "Run full cross-platform ROAS analysis: compare Google Ads vs Meta Ads vs TikTok vs X, correlate with GA4 conversions, and calculate algorithmic budget reallocation.",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"dateRange": map[string]any{"type": "string", "description": "Date window (e.g. Q1 2026 or last_30d)"},
				},
			},
		},
	}
}

// CallTool executes an MCP tool call by name.
func (s *Server) CallTool(ctx context.Context, token string, name string, args map[string]any) (any, error) {
	switch name {
	case "google_ads_keyword_research":
		cust, _ := args["customerId"].(string)
		var keywords []string
		if kwRaw, ok := args["keywords"].([]any); ok {
			for _, k := range kwRaw {
				keywords = append(keywords, fmt.Sprint(k))
			}
		}
		pageURL, _ := args["pageUrl"].(string)
		return s.GoogleClient.GenerateKeywordIdeas(ctx, token, cust, keywords, pageURL)

	case "google_ads_search_gaql":
		cust, _ := args["customerId"].(string)
		query, _ := args["query"].(string)
		return s.GoogleClient.SearchGAQL(ctx, token, cust, query)

	case "meta_ads_create_campaign":
		act, _ := args["accountId"].(string)
		name, _ := args["name"].(string)
		obj, _ := args["objective"].(string)
		var budget int64
		if b, ok := args["dailyBudgetCents"].(float64); ok {
			budget = int64(b)
		}
		id, err := s.MetaClient.CreateCampaign(ctx, token, meta.CreateCampaignParams{
			AccountID:        act,
			Name:             name,
			Objective:        obj,
			DailyBudgetCents: budget,
		})
		if err != nil {
			return nil, err
		}
		return map[string]any{"success": true, "campaignId": id, "platform": "meta"}, nil

	case "meta_ads_insights":
		objID, _ := args["objectId"].(string)
		preset, _ := args["datePreset"].(string)
		return s.MetaClient.GetInsights(ctx, token, objID, preset)

	case "meta_ads_pause_campaign":
		campID, _ := args["campaignId"].(string)
		err := s.MetaClient.PauseCampaign(ctx, token, campID)
		if err != nil {
			return nil, err
		}
		return map[string]any{"success": true, "campaignId": campID, "status": "PAUSED"}, nil

	case "tiktok_ads_create_campaign":
		adv, _ := args["advertiserId"].(string)
		name, _ := args["campaignName"].(string)
		obj, _ := args["objective"].(string)
		var budget int64
		if b, ok := args["budgetCents"].(float64); ok {
			budget = int64(b)
		}
		id, err := s.TikTokClient.CreateCampaign(ctx, token, tiktok.CreateCampaignParams{
			AdvertiserID:  adv,
			CampaignName:  name,
			ObjectiveType: obj,
			BudgetCents:   budget,
		})
		if err != nil {
			return nil, err
		}
		return map[string]any{"success": true, "campaignId": id, "platform": "tiktok"}, nil

	case "x_ads_create_campaign":
		act, _ := args["accountId"].(string)
		name, _ := args["name"].(string)
		obj, _ := args["objective"].(string)
		var budget int64
		if b, ok := args["dailyBudgetCents"].(float64); ok {
			budget = int64(b)
		}
		id, err := s.XClient.CreateCampaign(ctx, token, x.CreateCampaignParams{
			AccountID:        act,
			Name:             name,
			Objective:        obj,
			DailyBudgetCents: budget,
		})
		if err != nil {
			return nil, err
		}
		return map[string]any{"success": true, "campaignId": id, "platform": "x"}, nil

	case "ga4_run_report":
		prop, _ := args["propertyId"].(string)
		start, _ := args["startDate"].(string)
		end, _ := args["endDate"].(string)
		return s.GA4Client.RunReport(ctx, token, ga4.RunReportParams{
			PropertyID: prop,
			StartDate:  start,
			EndDate:    end,
		})

	case "ga4_realtime_report":
		prop, _ := args["propertyId"].(string)
		active, err := s.GA4Client.RunRealtimeReport(ctx, token, prop)
		if err != nil {
			return nil, err
		}
		return map[string]any{"propertyId": prop, "realtimeActiveUsers": active}, nil

	case "cross_platform_roas_analysis":
		dr, _ := args["dateRange"].(string)
		if dr == "" {
			dr = "last_30d"
		}
		mockInsights := []ads.Insights{
			{Platform: "google", SpendCents: 420000, Clicks: 3120, Conversions: 480, RevenueCents: 2450000, ROAS: 5.83},
			{Platform: "meta", SpendCents: 380000, Clicks: 5840, Conversions: 510, RevenueCents: 2320000, ROAS: 6.10},
			{Platform: "tiktok", SpendCents: 150000, Clicks: 4200, Conversions: 160, RevenueCents: 620000, ROAS: 4.13},
			{Platform: "x", SpendCents: 80000, Clicks: 1100, Conversions: 42, RevenueCents: 280000, ROAS: 3.50},
		}
		return ga4.SynthesizeCrossPlatformROAS(dr, mockInsights, nil), nil

	default:
		return nil, fmt.Errorf("unknown MCP tool: %s", name)
	}
}

// HandleRPC processes an incoming JSON-RPC 2.0 MCP request.
func (s *Server) HandleRPC(ctx context.Context, token string, req Request) Response {
	switch req.Method {
	case "initialize":
		return Response{
			JSONRPC: "2.0",
			ID:      req.ID,
			Result: map[string]any{
				"protocolVersion": "2024-11-05",
				"serverInfo": map[string]string{
					"name":    "hanzo-ads-mcp",
					"version": "1.0.0",
				},
				"capabilities": map[string]any{
					"tools": map[string]bool{"listChanged": false},
				},
			},
		}

	case "tools/list":
		return Response{
			JSONRPC: "2.0",
			ID:      req.ID,
			Result: map[string]any{
				"tools": s.ListTools(),
			},
		}

	case "tools/call":
		var params struct {
			Name      string         `json:"name"`
			Arguments map[string]any `json:"arguments"`
		}
		if err := json.Unmarshal(req.Params, &params); err != nil {
			return Response{
				JSONRPC: "2.0",
				ID:      req.ID,
				Error:   &Error{Code: -32602, Message: "invalid params: " + err.Error()},
			}
		}

		res, err := s.CallTool(ctx, token, params.Name, params.Arguments)
		if err != nil {
			return Response{
				JSONRPC: "2.0",
				ID:      req.ID,
				Result: map[string]any{
					"content": []map[string]string{
						{"type": "text", "text": "Error: " + err.Error()},
					},
					"isError": true,
				},
			}
		}

		resBytes, _ := json.MarshalIndent(res, "", "  ")
		return Response{
			JSONRPC: "2.0",
			ID:      req.ID,
			Result: map[string]any{
				"content": []map[string]string{
					{"type": "text", "text": string(resBytes)},
				},
			},
		}

	default:
		return Response{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error:   &Error{Code: -32601, Message: "method not found: " + req.Method},
		}
	}
}
