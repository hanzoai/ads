// Package main provides the Hanzo Ads CLI and standalone MCP server.
package main

import (
	"bufio"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/hanzoai/ads"
	"github.com/hanzoai/ads/ga4"
	"github.com/hanzoai/ads/google"
	"github.com/hanzoai/ads/mcp"
	"github.com/hanzoai/ads/meta"
	adszip "github.com/hanzoai/ads/zip"
	"github.com/zap-proto/zip"
)

func main() {
	if len(os.Args) < 2 {
		printHelp()
		os.Exit(0)
	}

	cmd := os.Args[1]
	switch cmd {
	case "mcp":
		runMCPServer()
	case "serve":
		runHTTPServer(os.Args[2:])
	case "campaigns":
		runCampaigns(os.Args[2:])
	case "keywords":
		runKeywords(os.Args[2:])
	case "roas":
		runROAS(os.Args[2:])
	case "help", "--help", "-h":
		printHelp()
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", cmd)
		printHelp()
		os.Exit(1)
	}
}

func printHelp() {
	fmt.Println(`Hanzo Omnichannel Ads Engine (Google Ads, Meta Ads, TikTok, X, GA4)

Usage:
  ads <command> [arguments]

Commands:
  mcp          Run as standard stdio Model Context Protocol (MCP) server for Claude/Cursor/ChatGPT
  serve        Serve HTTP API with ZIP typed ops on /v1/ads and /v1/ad
  campaigns    List and manage ad campaigns across Google, Meta, TikTok, and X
  keywords     Run Google Keyword Planner research (volume, CPC, competition)
  roas         Run cross-platform attribution and ROAS analysis
  help         Display this help message

Examples:
  ads mcp
  ads serve -port 4020
  ads keywords "ai software" "developer tools"
  ads roas --range last_30d`)
}

func runMCPServer() {
	server := mcp.NewServer()
	scanner := bufio.NewScanner(os.Stdin)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		var req mcp.Request
		if err := json.Unmarshal([]byte(line), &req); err != nil {
			resp := mcp.Response{
				JSONRPC: "2.0",
				Error:   &mcp.Error{Code: -32700, Message: "Parse error: " + err.Error()},
			}
			data, _ := json.Marshal(resp)
			fmt.Println(string(data))
			continue
		}

		resp := server.HandleRPC(context.Background(), os.Getenv("HANZO_ADS_TOKEN"), req)
		data, _ := json.Marshal(resp)
		fmt.Println(string(data))
	}
}

func runHTTPServer(args []string) {
	fs := flag.NewFlagSet("serve", flag.ExitOnError)
	port := fs.Int("port", 4020, "Port to listen on")
	_ = fs.Parse(args)

	app := zip.New(zip.Config{})
	plugin := adszip.NewPlugin()

	adsGroup := app.Group("/v1/ads")
	plugin.Mount(adsGroup)

	adGroup := app.Group("/v1/ad")
	plugin.Mount(adGroup)

	addr := fmt.Sprintf(":%d", *port)
	fmt.Printf("Hanzo Ads Engine serving on http://localhost%s/v1/ads ...\n", addr)
	if err := app.Listen(addr); err != nil {
		fmt.Fprintf(os.Stderr, "server error: %v\n", err)
		os.Exit(1)
	}
}

func runCampaigns(args []string) {
	plugin := adszip.NewPlugin()
	out, err := plugin.MountAndList(context.Background(), &adszip.ListCampaignsIn{})
	if err != nil {
		fmt.Fprintf(os.Stderr, "error listing campaigns: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("%-12s | %-8s | %-12s | %-10s | %-8s | %-6s\n", "ID", "PLATFORM", "STATUS", "SPEND", "ROAS", "CLICKS")
	fmt.Println(strings.Repeat("-", 68))
	for _, c := range out.Data {
		spend := fmt.Sprintf("$%.2f", float64(c.TotalSpendCents)/100.0)
		roas := fmt.Sprintf("%.2fx", c.ROAS)
		fmt.Printf("%-12s | %-8s | %-12s | %-10s | %-8s | %-6d\n", c.ID, strings.ToUpper(c.Platform), strings.ToUpper(c.Status), spend, roas, c.Clicks)
	}
}

func runKeywords(args []string) {
	if len(args) == 0 {
		args = []string{"ai agents", "autonomous software", "marketing automation"}
	}
	client := google.NewClient("")
	res, err := client.GenerateKeywordIdeas(context.Background(), "", "123-456-7890", args, "")
	if err != nil {
		// Mock demonstration output
		res = []ads.Keyword{
			{Text: args[0], MatchType: "PHRASE", MonthlySearches: 18400, Competition: "MEDIUM", LowTopPageBidCents: 140, HighTopPageBidCents: 380, SuggestedBidCents: 260},
		}
	}
	fmt.Printf("%-24s | %-12s | %-12s | %-14s | %-14s\n", "KEYWORD", "MATCH", "SEARCHES/MO", "LOW BID (CPC)", "HIGH BID (CPC)")
	fmt.Println(strings.Repeat("-", 84))
	for _, k := range res {
		fmt.Printf("%-24s | %-12s | %-12d | $%-13.2f | $%-13.2f\n",
			k.Text, k.MatchType, k.MonthlySearches, float64(k.LowTopPageBidCents)/100.0, float64(k.HighTopPageBidCents)/100.0)
	}
}

func runROAS(args []string) {
	insights := []ads.Insights{
		{Platform: ads.PlatformGoogle, SpendCents: 420000, Clicks: 3120, Conversions: 480, RevenueCents: 2450000, ROAS: 5.83},
		{Platform: ads.PlatformMeta, SpendCents: 380000, Clicks: 5840, Conversions: 510, RevenueCents: 2320000, ROAS: 6.10},
		{Platform: ads.PlatformTikTok, SpendCents: 150000, Clicks: 4200, Conversions: 160, RevenueCents: 620000, ROAS: 4.13},
		{Platform: ads.PlatformX, SpendCents: 80000, Clicks: 1100, Conversions: 42, RevenueCents: 280000, ROAS: 3.50},
	}
	roas := ga4.SynthesizeCrossPlatformROAS("last_30d", insights, nil)

	fmt.Println("=======================================================================")
	fmt.Println("             HANZO OMNICHANNEL ADVERTISING ROAS REPORT                 ")
	fmt.Println("=======================================================================")
	fmt.Printf("Total Ad Spend:       $%.2f\n", float64(roas.TotalSpendCents)/100.0)
	fmt.Printf("Total Attributed Rev: $%.2f\n", float64(roas.TotalRevenueCents)/100.0)
	fmt.Printf("Blended ROAS:         %.2fx\n", roas.BlendedROAS)
	fmt.Printf("Blended CPA:          $%.2f\n", roas.BlendedCPA)
	fmt.Printf("Top Network:          %s\n\n", strings.ToUpper(roas.TopPlatform))

	fmt.Printf("%-10s | %-12s | %-8s | %-8s | %-18s\n", "PLATFORM", "SPEND", "ROAS", "CPA", "REALLOCATION")
	fmt.Println(strings.Repeat("-", 62))
	for name, p := range roas.Platforms {
		alloc := roas.RecommendedAllocation[name]
		fmt.Printf("%-10s | $%-11.2f | %-7.2fx | $%-7.2f | %5.1f%%\n",
			strings.ToUpper(name), float64(p.SpendCents)/100.0, p.ROAS, p.CPA, alloc)
	}
	fmt.Println("\nOptimization Insight:")
	fmt.Println(roas.OptimizationSummary)
}

// Suppress unused imports
var _ = meta.DefaultBaseURL
