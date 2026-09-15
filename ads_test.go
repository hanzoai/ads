package ads_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/hanzoai/ads"
	"github.com/hanzoai/ads/ga4"
	"github.com/hanzoai/ads/google"
	"github.com/hanzoai/ads/mcp"
	"github.com/hanzoai/ads/meta"
	"github.com/hanzoai/ads/tiktok"
	"github.com/hanzoai/ads/x"
	adszip "github.com/hanzoai/ads/zip"
	"github.com/zap-proto/zip"
)

func TestNormalizePlatform(t *testing.T) {
	cases := []struct {
		input string
		want  string
		err   bool
	}{
		{"google", ads.PlatformGoogle, false},
		{"GOOGLE", ads.PlatformGoogle, false},
		{"meta", ads.PlatformMeta, false},
		{"Meta", ads.PlatformMeta, false},
		{"tiktok", ads.PlatformTikTok, false},
		{"x", ads.PlatformX, false},
		{"", ads.PlatformMeta, false}, // defaults to meta
		{"unsupported_ad_network", "", true},
	}

	for _, tc := range cases {
		got, err := ads.NormalizePlatform(tc.input)
		if tc.err && err == nil {
			t.Errorf("NormalizePlatform(%q) want error, got nil", tc.input)
		}
		if !tc.err && err != nil {
			t.Errorf("NormalizePlatform(%q) unexpected error: %v", tc.input, err)
		}
		if got != tc.want {
			t.Errorf("NormalizePlatform(%q) = %q, want %q", tc.input, got, tc.want)
		}
	}
}

func TestCrossPlatformROAS_Synthesis(t *testing.T) {
	insights := []ads.Insights{
		{Platform: ads.PlatformGoogle, SpendCents: 100000, RevenueCents: 500000, Conversions: 50, ROAS: 5.0},
		{Platform: ads.PlatformMeta, SpendCents: 80000, RevenueCents: 480000, Conversions: 40, ROAS: 6.0},
		{Platform: ads.PlatformTikTok, SpendCents: 50000, RevenueCents: 200000, Conversions: 20, ROAS: 4.0},
		{Platform: ads.PlatformX, SpendCents: 20000, RevenueCents: 60000, Conversions: 5, ROAS: 3.0},
	}

	report := ga4.SynthesizeCrossPlatformROAS("2026-Q1", insights, nil)

	if report.TotalSpendCents != 250000 {
		t.Errorf("TotalSpendCents want 250000, got %d", report.TotalSpendCents)
	}
	if report.TotalRevenueCents != 1240000 {
		t.Errorf("TotalRevenueCents want 1240000, got %d", report.TotalRevenueCents)
	}
	expectedBlendedROAS := 1240000.0 / 250000.0
	if report.BlendedROAS != expectedBlendedROAS {
		t.Errorf("BlendedROAS want %.2f, got %.2f", expectedBlendedROAS, report.BlendedROAS)
	}
	if report.TopPlatform != ads.PlatformMeta {
		t.Errorf("TopPlatform want meta, got %s", report.TopPlatform)
	}
	if len(report.RecommendedAllocation) != 4 {
		t.Errorf("RecommendedAllocation want 4 platforms, got %d", len(report.RecommendedAllocation))
	}
	// Meta has highest ROAS (6.0), so it should have highest recommended budget allocation
	metaAlloc := report.RecommendedAllocation[ads.PlatformMeta]
	googleAlloc := report.RecommendedAllocation[ads.PlatformGoogle]
	if metaAlloc <= googleAlloc {
		t.Errorf("Meta allocation (%.1f%%) must exceed Google allocation (%.1f%%)", metaAlloc, googleAlloc)
	}
}

func TestMCP_ToolsList(t *testing.T) {
	server := mcp.NewServer()
	tools := server.ListTools()

	if len(tools) < 12 {
		t.Fatalf("ListTools want at least 12 tools, got %d", len(tools))
	}

	toolMap := make(map[string]bool)
	for _, tool := range tools {
		toolMap[tool.Name] = true
	}

	expectedTools := []string{
		"google_ads_list_campaigns",
		"google_ads_create_campaign",
		"google_ads_keyword_research",
		"google_ads_search_gaql",
		"meta_ads_create_campaign",
		"meta_ads_create_adset",
		"meta_ads_insights",
		"meta_ads_pause_campaign",
		"tiktok_ads_create_campaign",
		"tiktok_ads_insights",
		"x_ads_create_campaign",
		"x_ads_insights",
		"ga4_run_report",
		"ga4_realtime_report",
		"cross_platform_roas_analysis",
	}

	for _, expected := range expectedTools {
		if !toolMap[expected] {
			t.Errorf("MCP missing expected tool: %s", expected)
		}
	}
}

func TestMCP_HandleRPC(t *testing.T) {
	server := mcp.NewServer()
	ctx := context.Background()

	// 1. Initialize
	initReq := mcp.Request{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "initialize",
	}
	initResp := server.HandleRPC(ctx, "fake-token", initReq)
	if initResp.Error != nil {
		t.Fatalf("initialize error: %v", initResp.Error)
	}

	// 2. Tools List
	listReq := mcp.Request{
		JSONRPC: "2.0",
		ID:      2,
		Method:  "tools/list",
	}
	listResp := server.HandleRPC(ctx, "fake-token", listReq)
	if listResp.Error != nil {
		t.Fatalf("tools/list error: %v", listResp.Error)
	}

	// 3. Tools Call: cross_platform_roas_analysis
	callParams, _ := json.Marshal(map[string]any{
		"name": "cross_platform_roas_analysis",
		"arguments": map[string]any{
			"dateRange": "last_30d",
		},
	})
	callReq := mcp.Request{
		JSONRPC: "2.0",
		ID:      3,
		Method:  "tools/call",
		Params:  callParams,
	}
	callResp := server.HandleRPC(ctx, "fake-token", callReq)
	if callResp.Error != nil {
		t.Fatalf("tools/call error: %v", callResp.Error)
	}
}

func TestGoogleClient_GenerateKeywordIdeas(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("developer-token") != "test-dev-token" {
			t.Errorf("developer-token header missing or invalid")
		}
		resp := map[string]any{
			"results": []map[string]any{
				{
					"text": "best crm for local business",
					"keywordIdeaMetrics": map[string]any{
						"avgMonthlySearches":     "14800",
						"competition":            "MEDIUM",
						"competitionIndex":       52,
						"lowTopOfPageBidMicros":  "1800000",
						"highTopOfPageBidMicros": "4200000",
					},
				},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	client := google.NewClient("test-dev-token")
	client.BaseURL = srv.URL

	keywords, err := client.GenerateKeywordIdeas(context.Background(), "auth-token", "123-456-7890", []string{"crm"}, "")
	if err != nil {
		t.Fatalf("GenerateKeywordIdeas: %v", err)
	}
	if len(keywords) != 1 {
		t.Fatalf("keywords want 1, got %d", len(keywords))
	}
	kw := keywords[0]
	if kw.Text != "best crm for local business" || kw.MonthlySearches != 14800 || kw.LowTopPageBidCents != 180 {
		t.Errorf("keyword parsed incorrectly: %+v", kw)
	}
}

func TestMetaClient_CreateCampaignAndInsights(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/campaigns") {
			_ = json.NewEncoder(w).Encode(map[string]any{"id": "meta_camp_12345"})
			return
		}
		if strings.HasSuffix(r.URL.Path, "/insights") {
			_ = json.NewEncoder(w).Encode(map[string]any{
				"data": []map[string]any{
					{
						"spend":       "542.50",
						"impressions": "88400",
						"clicks":      "4210",
						"ctr":         "4.76",
						"cpc":         "0.13",
						"cpm":         "6.14",
						"purchase_roas": []map[string]string{
							{"value": "6.45"},
						},
					},
				},
			})
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	client := meta.NewClient()
	client.BaseURL = srv.URL

	campID, err := client.CreateCampaign(context.Background(), "meta-tok", meta.CreateCampaignParams{
		AccountID:        "act_998877",
		Name:             "Spring Promo",
		Objective:        "OUTCOME_SALES",
		DailyBudgetCents: 5000,
	})
	if err != nil {
		t.Fatalf("CreateCampaign: %v", err)
	}
	if campID != "meta_camp_12345" {
		t.Errorf("campID want meta_camp_12345, got %s", campID)
	}

	insights, err := client.GetInsights(context.Background(), "meta-tok", campID, "last_30d")
	if err != nil {
		t.Fatalf("GetInsights: %v", err)
	}
	if insights.SpendCents != 54250 || insights.Impressions != 88400 || insights.ROAS != 6.45 {
		t.Errorf("insights parsed incorrectly: %+v", insights)
	}
}

func TestZIP_PluginEndpoints(t *testing.T) {
	app := zip.New(zip.Config{})
	plugin := adszip.NewPlugin()
	plugin.Mount(app.Group("/v1/ads"))

	ctx := context.Background()

	// Test ListCampaigns via programmatic helper
	list, err := plugin.MountAndList(ctx, &adszip.ListCampaignsIn{})
	if err != nil {
		t.Fatalf("MountAndList error: %v", err)
	}
	if len(list.Data) < 2 {
		t.Fatalf("seed campaigns want at least 2, got %d", len(list.Data))
	}

	// Test Summary
	summReq, _ := http.NewRequestWithContext(ctx, http.MethodGet, "/v1/ads/summary", nil)
	summResp, err := app.Test(summReq)
	if err != nil {
		t.Fatalf("Test summary: %v", err)
	}
	if summResp.StatusCode != http.StatusOK {
		t.Errorf("summary status want 200, got %d", summResp.StatusCode)
	}

	// Test Keyword Research
	kwReq, _ := http.NewRequestWithContext(ctx, http.MethodPost, "/v1/ads/keywords", strings.NewReader(`{"seedKeywords":["ai ads"]}`))
	kwReq.Header.Set("Content-Type", "application/json")
	kwResp, err := app.Test(kwReq)
	if err != nil {
		t.Fatalf("Test keywords: %v", err)
	}
	if kwResp.StatusCode != http.StatusOK {
		t.Errorf("keywords status want 200, got %d", kwResp.StatusCode)
	}

	// Test CrossPlatform ROAS
	roasReq, _ := http.NewRequestWithContext(ctx, http.MethodGet, "/v1/ads/roas", nil)
	roasResp, err := app.Test(roasReq)
	if err != nil {
		t.Fatalf("Test roas: %v", err)
	}
	if roasResp.StatusCode != http.StatusOK {
		t.Errorf("roas status want 200, got %d", roasResp.StatusCode)
	}

	// Test GA4
	ga4Req, _ := http.NewRequestWithContext(ctx, http.MethodGet, "/v1/ads/ga4", nil)
	ga4Resp, err := app.Test(ga4Req)
	if err != nil {
		t.Fatalf("Test ga4: %v", err)
	}
	if ga4Resp.StatusCode != http.StatusOK {
		t.Errorf("ga4 status want 200, got %d", ga4Resp.StatusCode)
	}
}

// Suppress unused imports
var _ = x.DefaultBaseURL
var _ = tiktok.DefaultBaseURL
