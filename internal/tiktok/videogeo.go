package tiktok

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/url"
	"strings"
)

// CDNRegionMap maps TikTok CDN path tokens to human-readable regions.
// These tokens appear in TikTok video CDN URLs and encode the upload region,
// i.e. the data centre the video was uploaded to — a strong geolocation signal.
var CDNRegionMap = map[string]string{
	"useast2a":    "US East (Virginia)",
	"useast5":     "US East (Ohio)",
	"useast2b":    "US East (N. Virginia)",
	"us-east-2":   "US East",
	"us-east-1":   "US East",
	"uswest2a":    "US West (Oregon)",
	"us-west-2":   "US West",
	"us-west-1":   "US West (N. California)",
	"eu-central-1": "EU (Frankfurt)",
	"eu-west-1":   "EU (Ireland)",
	"eu-west-2":   "EU (London)",
	"ap-southeast-1": "Asia Pacific (Singapore)",
	"ap-southeast-2": "Asia Pacific (Sydney)",
	"ap-northeast-1": "Asia Pacific (Tokyo)",
	"maliva":      "Malaysia / SE Asia",
	"alisg":       "Singapore (Alibaba)",
	"alias":       "Singapore (Alibaba global)",
	"sg":          "Singapore",
	"in":          "India",
	"br":          "Brazil",
	"ap":          "Asia Pacific",
	"na":          "North America",
	"aka":         "Global (Akamai CDN)",
}

// cdnCountryMap maps CDN region names to likely country (used for scoring).
var cdnCountryMap = map[string]string{
	"useast2a": "United States", "useast5": "United States",
	"useast2b": "United States", "us-east-2": "United States",
	"us-east-1": "United States", "uswest2a": "United States",
	"us-west-2": "United States", "us-west-1": "United States",
	"eu-central-1": "Germany", "eu-west-1": "Ireland",
	"eu-west-2": "United Kingdom",
	"ap-southeast-1": "Singapore", "maliva": "Malaysia",
	"alisg": "Singapore", "alias": "Singapore",
	"br": "Brazil", "in": "India",
}

// IPGeoInfo holds GeoIP lookup results.
type IPGeoInfo struct {
	IP          string
	Country     string
	CountryCode string
	Region      string
	City        string
	Lat         float64
	Lon         float64
	ISP         string
}

// VideoGeoResult holds all location intelligence extracted from a single video.
type VideoGeoResult struct {
	VideoID    string
	Desc       string
	POI        *POIInfo   // tagged location from metadata
	CDNToken   string     // raw CDN region token from URL path
	CDNRegion  string     // human-readable CDN region
	CDNIP      string     // IP from DNS lookup of CDN hostname
	IPGeo      *IPGeoInfo // GeoIP result for CDNIP
	GeoScore   float64    // contribution to country scoring
}

// ipAPIResponse matches ip-api.com/json/{ip} response.
type ipAPIResponse struct {
	Status      string  `json:"status"`
	Country     string  `json:"country"`
	CountryCode string  `json:"countryCode"`
	RegionName  string  `json:"regionName"`
	City        string  `json:"city"`
	Lat         float64 `json:"lat"`
	Lon         float64 `json:"lon"`
	ISP         string  `json:"isp"`
	Query       string  `json:"query"`
}

// VideoGeoAnalyzer resolves video CDN IPs and extracts location metadata.
type VideoGeoAnalyzer struct {
	scraper *Scraper
}

// NewVideoGeoAnalyzer returns a VideoGeoAnalyzer.
func NewVideoGeoAnalyzer(scraper *Scraper) *VideoGeoAnalyzer {
	return &VideoGeoAnalyzer{scraper: scraper}
}

// Analyze extracts geolocation data from a VideoItem.
func (v *VideoGeoAnalyzer) Analyze(ctx context.Context, video VideoItem) *VideoGeoResult {
	vgr := &VideoGeoResult{
		VideoID: video.VideoID,
		Desc:    video.Desc,
		POI:     video.POI,
	}

	// Try to resolve CDN IP from the first available play URL
	for _, playURL := range video.PlayURLs {
		token, region := parseCDNToken(playURL)
		if token != "" {
			vgr.CDNToken = token
			vgr.CDNRegion = region
		}

		ip, err := resolveCDNIP(playURL)
		if err == nil && ip != "" {
			vgr.CDNIP = ip
			geo, err := v.geolocateIP(ctx, ip)
			if err == nil {
				vgr.IPGeo = geo
			}
			break
		}
	}

	return vgr
}

// parseCDNToken finds the region token embedded in TikTok CDN URL paths.
// Example: https://v16m-webapp.tiktok.com/video/tos/useast2a/tos-useast2a-ve-...
func parseCDNToken(cdnURL string) (token, humanRegion string) {
	// Check URL path segments
	u, err := url.Parse(cdnURL)
	if err != nil {
		return "", ""
	}
	parts := strings.Split(u.Path, "/")
	for _, p := range parts {
		p = strings.ToLower(p)
		if human, ok := CDNRegionMap[p]; ok {
			return p, human
		}
		// Also check inside hyphenated segments like "tos-useast2a-ve"
		for _, sub := range strings.Split(p, "-") {
			if human, ok := CDNRegionMap[sub]; ok {
				return sub, human
			}
		}
	}
	// Check hostname too, e.g. v16m-webapp.tiktok.com might have region info
	host := strings.ToLower(u.Hostname())
	for token, human := range CDNRegionMap {
		if strings.Contains(host, token) {
			return token, human
		}
	}
	return "", ""
}

// resolveCDNIP does a DNS lookup on the CDN hostname.
func resolveCDNIP(cdnURL string) (string, error) {
	u, err := url.Parse(cdnURL)
	if err != nil {
		return "", err
	}
	hostname := u.Hostname()
	if hostname == "" {
		return "", fmt.Errorf("no hostname in URL")
	}
	ips, err := net.LookupHost(hostname)
	if err != nil || len(ips) == 0 {
		return "", fmt.Errorf("DNS lookup failed: %w", err)
	}
	return ips[0], nil
}

// geolocateIP calls ip-api.com to turn an IP into country/city/ISP.
func (v *VideoGeoAnalyzer) geolocateIP(ctx context.Context, ip string) (*IPGeoInfo, error) {
	apiURL := fmt.Sprintf(
		"http://ip-api.com/json/%s?fields=status,country,countryCode,regionName,city,lat,lon,isp,query",
		ip,
	)
	body, err := v.scraper.fetchWithBrowserHeaders(ctx, apiURL)
	if err != nil {
		return nil, err
	}
	var resp ipAPIResponse
	if err := json.Unmarshal([]byte(body), &resp); err != nil {
		return nil, err
	}
	if resp.Status != "success" {
		return nil, fmt.Errorf("ip-api returned status %q", resp.Status)
	}
	return &IPGeoInfo{
		IP:          resp.Query,
		Country:     resp.Country,
		CountryCode: resp.CountryCode,
		Region:      resp.RegionName,
		City:        resp.City,
		Lat:         resp.Lat,
		Lon:         resp.Lon,
		ISP:         resp.ISP,
	}, nil
}

// CountryFromCDNToken maps a CDN token directly to the most likely country.
func CountryFromCDNToken(token string) string {
	return cdnCountryMap[strings.ToLower(token)]
}
