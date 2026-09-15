// Package ads provides a unified omnichannel advertising and analytics engine
// for Google Ads, Meta Ads (Facebook/Instagram), TikTok Business, X (Twitter) Ads,
// and Google Analytics 4 (GA4).
//
// It serves as the native Go foundation for api.hanzo.ai (/v1/ads and /v1/ad),
// delivering:
//   1. Cross-network campaign lifecycle management (create, update, pause, delete, launch)
//   2. Ad set, targeting, and creative composition (Search, Display, Social, Video)
//   3. Keyword planning and search query intelligence (GAQL)
//   4. Audience synchronization (Custom, Lookalike, Customer Match)
//   5. Performance reporting, unified ROAS, and GA4 conversion attribution
//   6. Native Model Context Protocol (MCP) server for AI assistants (Claude, Cursor, ChatGPT)
//   7. CLI tooling and ZIP framework typed ops for zero-overhead projection.
package ads

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

// Supported advertising platforms.
const (
	PlatformGoogle = "google"
	PlatformMeta   = "meta"
	PlatformTikTok = "tiktok"
	PlatformX      = "x"
)

// Supported campaign statuses.
const (
	StatusDraft     = "draft"
	StatusActive    = "active"
	StatusPaused    = "paused"
	StatusCompleted = "completed"
)

// Supported campaign objectives.
const (
	ObjectiveTraffic     = "TRAFFIC"
	ObjectiveLeads       = "LEAD_GENERATION"
	ObjectiveSales       = "CONVERSIONS"
	ObjectiveAwareness   = "BRAND_AWARENESS"
	ObjectiveAppInstalls = "APP_INSTALLS"
)

var (
	// ErrNotFound is returned when an entity is missing.
	ErrNotFound = errors.New("ads: entity not found")
	// ErrNotConnected is returned when an ad network account is not connected.
	ErrNotConnected = errors.New("ads: provider account not connected")
	// ErrUnsupportedPlatform is returned when an unsupported ad network is specified.
	ErrUnsupportedPlatform = errors.New("ads: unsupported platform")
	// ErrInvalidInput is returned on invalid request parameters.
	ErrInvalidInput = errors.New("ads: invalid input")
	// ErrUpstream is returned when an ad network API call fails.
	ErrUpstream = errors.New("ads: upstream ad network error")
)

// Campaign represents a unified paid advertising campaign across any ad network.
type Campaign struct {
	ID               string    `json:"id"`
	Org              string    `json:"org"`
	Platform         string    `json:"platform"`
	AccountID        string    `json:"accountId"`
	ExternalID       string    `json:"externalId,omitempty"`
	Name             string    `json:"name"`
	Objective        string    `json:"objective"`
	Status           string    `json:"status"`
	DailyBudgetCents int64     `json:"dailyBudgetCents"`
	TotalSpendCents  int64     `json:"totalSpendCents"`
	Impressions      int64     `json:"impressions"`
	Clicks           int64     `json:"clicks"`
	Conversions      int64     `json:"conversions"`
	CTR              float64   `json:"ctr"`
	CPC              float64   `json:"cpc"`
	ROAS             float64   `json:"roas"`
	CreatedAt        time.Time `json:"createdAt"`
	UpdatedAt        time.Time `json:"updatedAt"`
}

// AdSet represents an ad group, line item, or ad set with targeting parameters.
type AdSet struct {
	ID               string    `json:"id"`
	CampaignID       string    `json:"campaignId"`
	Platform         string    `json:"platform"`
	Name             string    `json:"name"`
	DailyBudgetCents int64     `json:"dailyBudgetCents"`
	BidAmountCents   int64     `json:"bidAmountCents"`
	Status           string    `json:"status"`
	Targeting        Targeting `json:"targeting"`
}

// Targeting defines audience, geographic, and device targeting filters.
type Targeting struct {
	GeoLocations []string `json:"geoLocations,omitempty"`
	AgeMin       int      `json:"ageMin,omitempty"`
	AgeMax       int      `json:"ageMax,omitempty"`
	Genders      []string `json:"genders,omitempty"`
	Interests    []string `json:"interests,omitempty"`
	Devices      []string `json:"devices,omitempty"`
	Placements   []string `json:"placements,omitempty"`
}

// Creative represents ad copy, visuals, and destination URLs.
type Creative struct {
	ID           string   `json:"id"`
	Platform     string   `json:"platform"`
	Name         string   `json:"name"`
	Headlines    []string `json:"headlines"`
	Descriptions []string `json:"descriptions"`
	ImageURLs    []string `json:"imageUrls,omitempty"`
	VideoURLs    []string `json:"videoUrls,omitempty"`
	CallToAction string   `json:"callToAction"`
	FinalURL     string   `json:"finalUrl"`
}

// Keyword represents keyword research data from Google Keyword Planner or GAQL.
type Keyword struct {
	Text                string  `json:"text"`
	MatchType           string  `json:"matchType"` // EXACT, PHRASE, BROAD
	MonthlySearches     int64   `json:"monthlySearches"`
	Competition         string  `json:"competition"` // LOW, MEDIUM, HIGH
	CompetitionIndex    int     `json:"competitionIndex"`
	LowTopPageBidCents  int64   `json:"lowTopPageBidCents"`
	HighTopPageBidCents int64   `json:"highTopPageBidCents"`
	SuggestedBidCents   int64   `json:"suggestedBidCents"`
}

// Audience represents a custom, lookalike, or remarketing audience segment.
type Audience struct {
	ID          string `json:"id"`
	Platform    string `json:"platform"`
	Name        string `json:"name"`
	Type        string `json:"type"` // CUSTOM, LOOKALIKE, CUSTOMER_MATCH, REMARKETING
	Description string `json:"description"`
	ApproxSize  int64  `json:"approxSize"`
}

// Insights captures multi-channel ad performance metrics over a date window.
type Insights struct {
	Platform     string  `json:"platform"`
	AccountID    string  `json:"accountId"`
	CampaignID   string  `json:"campaignId,omitempty"`
	DateStart    string  `json:"dateStart"`
	DateEnd      string  `json:"dateEnd"`
	Impressions  int64   `json:"impressions"`
	Clicks       int64   `json:"clicks"`
	SpendCents   int64   `json:"spendCents"`
	Conversions  int64   `json:"conversions"`
	RevenueCents int64   `json:"revenueCents"`
	CTR          float64 `json:"ctr"`
	CPC          float64 `json:"cpc"`
	CPM          float64 `json:"cpm"`
	CPA          float64 `json:"cpa"`
	ROAS         float64 `json:"roas"`
}

// GA4Report captures Google Analytics 4 sessions, conversions, and traffic data.
type GA4Report struct {
	PropertyID          string            `json:"propertyId"`
	DateRange           string            `json:"dateRange"`
	ActiveUsers         int64             `json:"activeUsers"`
	Sessions            int64             `json:"sessions"`
	ScreenPageViews     int64             `json:"screenPageViews"`
	Conversions         int64             `json:"conversions"`
	PurchaseRevenue     float64           `json:"purchaseRevenue"`
	BounceRate          float64           `json:"bounceRate"`
	TopChannels         []ChannelShare    `json:"topChannels"`
	RealtimeActiveUsers int64             `json:"realtimeActiveUsers,omitempty"`
}

// ChannelShare represents attribution traffic and conversion breakdown.
type ChannelShare struct {
	ChannelName string  `json:"channelName"`
	Sessions    int64   `json:"sessions"`
	Conversions int64   `json:"conversions"`
	Revenue     float64 `json:"revenue"`
}

// CrossPlatformROAS evaluates omnichannel advertising return across all networks.
type CrossPlatformROAS struct {
	DateRange             string                    `json:"dateRange"`
	TotalSpendCents       int64                     `json:"totalSpendCents"`
	TotalRevenueCents     int64                     `json:"totalRevenueCents"`
	BlendedROAS           float64                   `json:"blendedRoas"`
	BlendedCPA            float64                   `json:"blendedCpa"`
	TopPlatform           string                    `json:"topPlatform"`
	Platforms             map[string]PlatformROAS   `json:"platforms"`
	RecommendedAllocation map[string]float64        `json:"recommendedAllocation"`
	OptimizationSummary   string                    `json:"optimizationSummary"`
}

// PlatformROAS details per-network metrics.
type PlatformROAS struct {
	Platform    string  `json:"platform"`
	SpendCents  int64   `json:"spendCents"`
	Clicks      int64   `json:"clicks"`
	Conversions int64   `json:"conversions"`
	ROAS        float64 `json:"roas"`
	CPA         float64 `json:"cpa"`
}

// NormalizePlatform verifies and defaults the platform identifier.
func NormalizePlatform(p string) (string, error) {
	norm := strings.ToLower(strings.TrimSpace(p))
	if norm == "" {
		return PlatformMeta, nil
	}
	switch norm {
	case PlatformGoogle, PlatformMeta, PlatformTikTok, PlatformX:
		return norm, nil
	default:
		return "", fmt.Errorf("%w: %s", ErrUnsupportedPlatform, p)
	}
}
