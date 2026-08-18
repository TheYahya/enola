package tiktok

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

// TikTokProfile holds scraped profile data.
type TikTokProfile struct {
	Username    string
	Bio         string
	Links       []string
	Nickname    string
	Region      string // if TikTok exposes it
	VideoTitles []string
	Hashtags    []string
	RawText     string // all harvested text concatenated
}

// Scraper handles all HTTP fetching.
type Scraper struct {
	client *http.Client
}

// NewScraper returns a Scraper with browser-like defaults.
func NewScraper(client *http.Client) *Scraper {
	return &Scraper{client: client}
}

// DefaultScraper builds a Scraper with sensible timeout/redirect settings.
func DefaultScraper() *Scraper {
	client := &http.Client{
		Timeout: 20 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 8 {
				return fmt.Errorf("too many redirects")
			}
			return nil
		},
	}
	return &Scraper{client: client}
}

var (
	reEmails   = regexp.MustCompile(`[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}`)
	reGmails   = regexp.MustCompile(`[a-zA-Z0-9._%+\-]+@gmail\.com`)
	reLinks    = regexp.MustCompile(`https?://[^\s"'<>\]\)]+`)
	reHashtags = regexp.MustCompile(`#[A-Za-z0-9_]+`)

	// JSON keys TikTok embeds in the page HTML
	reBioJSON    = regexp.MustCompile(`"desc"\s*:\s*"([^"]*)"`)
	reNickJSON   = regexp.MustCompile(`"nickname"\s*:\s*"([^"]*)"`)
	reRegionJSON = regexp.MustCompile(`"region"\s*:\s*"([^"]*)"`)
	reBioAlt     = regexp.MustCompile(`"signature"\s*:\s*"([^"]*)"`)
)

// tiktokOEmbed is the TikTok oEmbed API response.
type tiktokOEmbed struct {
	Title        string `json:"title"`
	AuthorName   string `json:"author_name"`
	AuthorURL    string `json:"author_url"`
	HTML         string `json:"html"`
	ThumbnailURL string `json:"thumbnail_url"`
}

// FetchTikTokProfile fetches and parses a TikTok user profile.
func (s *Scraper) FetchTikTokProfile(ctx context.Context, username string) (*TikTokProfile, error) {
	profile := &TikTokProfile{Username: username}

	// Try oEmbed first (lightweight, no bot detection)
	oembed, _ := s.fetchOEmbed(ctx, username)
	if oembed != nil {
		profile.Nickname = oembed.AuthorName
	}

	// Fetch the profile page
	profileURL := fmt.Sprintf("https://www.tiktok.com/@%s", username)
	body, err := s.fetchWithBrowserHeaders(ctx, profileURL)
	if err == nil {
		profile.RawText = body
		extractProfileFromHTML(body, profile)
	}

	// If bio still empty, try the mobile URL
	if profile.Bio == "" {
		mobileURL := fmt.Sprintf("https://m.tiktok.com/h5/share/usr/@%s", username)
		mBody, err := s.fetchWithBrowserHeaders(ctx, mobileURL)
		if err == nil {
			profile.RawText += " " + mBody
			extractProfileFromHTML(mBody, profile)
		}
	}

	return profile, nil
}

// fetchOEmbed calls TikTok's oEmbed endpoint.
func (s *Scraper) fetchOEmbed(ctx context.Context, username string) (*tiktokOEmbed, error) {
	profileURL := fmt.Sprintf("https://www.tiktok.com/@%s", username)
	apiURL := fmt.Sprintf("https://www.tiktok.com/oembed?url=%s", url.QueryEscape(profileURL))

	body, err := s.fetchWithBrowserHeaders(ctx, apiURL)
	if err != nil {
		return nil, err
	}
	var oe tiktokOEmbed
	if err := json.Unmarshal([]byte(body), &oe); err != nil {
		return nil, err
	}
	return &oe, nil
}

// extractProfileFromHTML parses TikTok's embedded JSON and HTML for profile data.
func extractProfileFromHTML(body string, profile *TikTokProfile) {
	// Extract bio
	if m := reBioJSON.FindStringSubmatch(body); len(m) > 1 {
		profile.Bio = unescapeJSON(m[1])
	}
	if profile.Bio == "" {
		if m := reBioAlt.FindStringSubmatch(body); len(m) > 1 {
			profile.Bio = unescapeJSON(m[1])
		}
	}

	// Extract nickname
	if profile.Nickname == "" {
		if m := reNickJSON.FindStringSubmatch(body); len(m) > 1 {
			profile.Nickname = unescapeJSON(m[1])
		}
	}

	// Extract region hint
	if m := reRegionJSON.FindStringSubmatch(body); len(m) > 1 {
		profile.Region = m[1]
	}

	// Extract links from bio text
	for _, link := range reLinks.FindAllString(body, -1) {
		link = strings.TrimRight(link, ".,;!?")
		if isUsefulLink(link) {
			profile.Links = append(profile.Links, link)
		}
	}
	profile.Links = dedupe(profile.Links)

	// Extract hashtags from the raw body
	for _, tag := range reHashtags.FindAllString(body, -1) {
		profile.Hashtags = append(profile.Hashtags, strings.ToLower(tag))
	}
	profile.Hashtags = dedupe(profile.Hashtags)
}

// isUsefulLink filters out TikTok's own CDN/tracking links.
func isUsefulLink(link string) bool {
	skip := []string{
		"tiktok.com", "tiktokcdn.com", "tiktokv.com", "muscdn.com",
		"javascript:", "data:", "cdn.", "static.", "analytics.",
		"firebase", "sentry", "newrelic",
	}
	lower := strings.ToLower(link)
	for _, s := range skip {
		if strings.Contains(lower, s) {
			return false
		}
	}
	return strings.HasPrefix(link, "http")
}

// FetchPage fetches a generic web page and returns the body text.
func (s *Scraper) FetchPage(ctx context.Context, pageURL string) (string, error) {
	return s.fetchWithBrowserHeaders(ctx, pageURL)
}

// SearchDDG performs a DuckDuckGo HTML search and returns body text.
func (s *Scraper) SearchDDG(ctx context.Context, query string) (string, error) {
	searchURL := fmt.Sprintf("https://html.duckduckgo.com/html/?q=%s", url.QueryEscape(query))
	return s.fetchWithBrowserHeaders(ctx, searchURL)
}

// fetchWithBrowserHeaders does an HTTP GET with realistic browser headers.
func (s *Scraper) fetchWithBrowserHeaders(ctx context.Context, targetURL string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, targetURL, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/125.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req.Header.Set("Accept-Encoding", "gzip, deflate, br")
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("DNT", "1")

	resp, err := s.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(io.LimitReader(resp.Body, 2<<20)) // 2 MB cap
	if err != nil {
		return "", err
	}
	return string(raw), nil
}

// ExtractEmails extracts all email addresses from text.
func ExtractEmails(text string) []string {
	return dedupe(reEmails.FindAllString(text, -1))
}

// ExtractGmails extracts only Gmail addresses from text.
func ExtractGmails(text string) []string {
	return dedupe(reGmails.FindAllString(text, -1))
}

// ExtractHashtags returns all #hashtags found in text.
func ExtractHashtags(text string) []string {
	tags := reHashtags.FindAllString(text, -1)
	for i, t := range tags {
		tags[i] = strings.ToLower(t)
	}
	return dedupe(tags)
}

// ParseUsername extracts the raw username from a TikTok URL or handle.
func ParseUsername(input string) string {
	input = strings.TrimSpace(input)
	// Handle full URLs: https://www.tiktok.com/@username or https://vm.tiktok.com/...
	if strings.Contains(input, "tiktok.com") {
		if idx := strings.Index(input, "@"); idx != -1 {
			handle := input[idx+1:]
			handle = strings.Split(handle, "?")[0]
			handle = strings.Split(handle, "/")[0]
			return handle
		}
	}
	return strings.TrimPrefix(input, "@")
}

// unescapeJSON converts \\n → \n, \\u00XX → rune, etc. for JSON string values.
func unescapeJSON(s string) string {
	// Quick unescape by wrapping in a JSON string and unmarshalling.
	var out string
	if err := json.Unmarshal([]byte(`"`+s+`"`), &out); err == nil {
		return out
	}
	return s
}

// dedupe removes duplicate strings preserving order.
func dedupe(ss []string) []string {
	seen := make(map[string]struct{}, len(ss))
	out := ss[:0]
	for _, s := range ss {
		if _, ok := seen[s]; !ok {
			seen[s] = struct{}{}
			out = append(out, s)
		}
	}
	return out
}
