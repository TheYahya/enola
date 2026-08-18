package tiktok

import (
	"context"
	"fmt"
	"strings"
)

// GmailInfo holds what we discover about a Gmail address.
type GmailInfo struct {
	Email       string
	Sources     []string // URLs where the email was found
	CluesText   []string // raw text snippets containing geolocation clues
	GoogleMaps  string   // Google Maps profile URL if found
}

// GmailFinder aggressively hunts Gmail addresses and associated info.
type GmailFinder struct {
	scraper *Scraper
}

// NewGmailFinder returns a GmailFinder.
func NewGmailFinder(scraper *Scraper) *GmailFinder {
	return &GmailFinder{scraper: scraper}
}

// DiscoverFromPages fetches each URL and extracts Gmail addresses from the content.
func (g *GmailFinder) DiscoverFromPages(ctx context.Context, urls []string) map[string]*GmailInfo {
	results := make(map[string]*GmailInfo)
	for _, pageURL := range urls {
		body, err := g.scraper.FetchPage(ctx, pageURL)
		if err != nil {
			continue
		}
		for _, email := range ExtractGmails(body) {
			email = strings.ToLower(email)
			if _, exists := results[email]; !exists {
				results[email] = &GmailInfo{Email: email}
			}
			results[email].Sources = append(results[email].Sources, pageURL)
			// Grab surrounding text as a clue
			for _, snippet := range extractSnippets(body, email, 200) {
				results[email].CluesText = append(results[email].CluesText, snippet)
			}
		}
	}
	return results
}

// SearchForEmail searches DuckDuckGo aggressively for a Gmail and location clues.
func (g *GmailFinder) SearchForEmail(ctx context.Context, email string) []string {
	queries := []string{
		fmt.Sprintf(`"%s"`, email),
		fmt.Sprintf(`"%s" location`, email),
		fmt.Sprintf(`"%s" address country`, email),
		fmt.Sprintf(`"%s" jamaica OR trinidad OR barbados OR caribbean`, email),
		fmt.Sprintf(`site:instagram.com "%s"`, email),
		fmt.Sprintf(`site:facebook.com "%s"`, email),
	}

	var clues []string
	for _, q := range queries {
		body, err := g.scraper.SearchDDG(ctx, q)
		if err != nil {
			continue
		}
		clues = append(clues, extractSearchSnippets(body, 5)...)
	}
	return clues
}

// SearchGoogleMaps looks for a public Google Maps contributor profile.
// Google no longer exposes contributor pages by email directly, but we
// can search for review text linking the email to a location.
func (g *GmailFinder) SearchGoogleMaps(ctx context.Context, email string) string {
	q := fmt.Sprintf(`site:maps.google.com "%s"`, email)
	body, err := g.scraper.SearchDDG(ctx, q)
	if err != nil {
		return ""
	}
	// Look for Google Maps links in search results
	for _, link := range reLinks.FindAllString(body, -1) {
		if strings.Contains(link, "maps.google") || strings.Contains(link, "google.com/maps") {
			return link
		}
	}
	return ""
}

// SearchLinkedAccounts searches for social accounts linked to the Gmail.
func (g *GmailFinder) SearchLinkedAccounts(ctx context.Context, email string) []string {
	queries := []string{
		fmt.Sprintf(`"%s" site:linkedin.com`, email),
		fmt.Sprintf(`"%s" site:twitter.com OR site:x.com`, email),
		fmt.Sprintf(`"%s" site:youtube.com`, email),
	}
	var found []string
	for _, q := range queries {
		body, err := g.scraper.SearchDDG(ctx, q)
		if err != nil {
			continue
		}
		for _, link := range reLinks.FindAllString(body, -1) {
			if isProfileLink(link) {
				found = append(found, link)
			}
		}
	}
	return dedupe(found)
}

// isProfileLink returns true for likely social-profile URLs.
func isProfileLink(link string) bool {
	social := []string{
		"linkedin.com/in/", "twitter.com/", "x.com/",
		"instagram.com/", "facebook.com/", "youtube.com/channel/",
		"youtube.com/user/", "tiktok.com/@",
	}
	for _, s := range social {
		if strings.Contains(link, s) {
			return true
		}
	}
	return false
}

// extractSnippets returns up to maxCount short context windows around needle in text.
func extractSnippets(text, needle string, windowSize int) []string {
	var out []string
	lower := strings.ToLower(text)
	lowerNeedle := strings.ToLower(needle)
	start := 0
	for {
		idx := strings.Index(lower[start:], lowerNeedle)
		if idx == -1 {
			break
		}
		abs := start + idx
		lo := abs - windowSize
		if lo < 0 {
			lo = 0
		}
		hi := abs + len(needle) + windowSize
		if hi > len(text) {
			hi = len(text)
		}
		snippet := strings.TrimSpace(text[lo:hi])
		// Strip HTML tags crudely
		snippet = stripHTML(snippet)
		if snippet != "" {
			out = append(out, snippet)
		}
		start = abs + 1
		if len(out) >= 3 {
			break
		}
	}
	return out
}

// extractSearchSnippets pulls the first n result snippets from a DuckDuckGo HTML page.
func extractSearchSnippets(body string, n int) []string {
	// DuckDuckGo HTML result snippets are in <a class="result__snippet">…</a>
	// We do a simple string extraction rather than a full HTML parser.
	var out []string
	lower := body
	marker := `class="result__snippet"`
	start := 0
	for len(out) < n {
		idx := strings.Index(lower[start:], marker)
		if idx == -1 {
			break
		}
		abs := start + idx
		// find end of the tag content
		tagEnd := strings.Index(lower[abs:], ">")
		if tagEnd == -1 {
			break
		}
		contentStart := abs + tagEnd + 1
		closeTag := strings.Index(lower[contentStart:], "</a>")
		if closeTag == -1 {
			break
		}
		snippet := stripHTML(body[contentStart : contentStart+closeTag])
		if snippet != "" {
			out = append(out, snippet)
		}
		start = contentStart + closeTag + 1
	}
	return out
}

// stripHTML removes HTML tags from a string.
func stripHTML(s string) string {
	var b strings.Builder
	inTag := false
	for _, r := range s {
		switch {
		case r == '<':
			inTag = true
		case r == '>':
			inTag = false
		case !inTag:
			b.WriteRune(r)
		}
	}
	return strings.TrimSpace(b.String())
}
