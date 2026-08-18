// Package tiktok provides TikTok account geolocation analysis.
// It scrapes the profile, harvests linked pages, finds Gmail addresses,
// and scores each country using pattern matching tuned for Jamaica and the
// wider Caribbean region.
package tiktok

import (
	"context"
	"fmt"
	"regexp"
	"sort"
	"strings"
)

// Signal represents one piece of geolocation evidence.
type Signal struct {
	Country  string
	Evidence string // human-readable description of what was found
	Source   string // where it was found (bio, web search, gmail, etc.)
	Weight   float64
}

// GeoResult is the final output of a full analysis.
type GeoResult struct {
	Username      string
	TikTokURL     string
	Profile       *TikTokProfile
	Signals       []Signal
	CountryScores map[string]float64
	BestGuess     string
	Confidence    string // High / Medium / Low
	Emails        []EmailResult
	LinkedURLs    []string
	Videos        []VideoItem
	VideoGeos     []VideoGeoResult
	Comments      []CommentItem // comments BY the target user on their own videos
}

// EmailResult bundles a discovered email with all findings about it.
type EmailResult struct {
	Email          string
	FoundAt        []string
	Clues          []string
	LinkedAccounts []string
}

// GeoLocator orchestrates the full geolocation pipeline.
type GeoLocator struct {
	scraper      *Scraper
	gmailFinder  *GmailFinder
	apiClient    *APIClient
	videoAnalyzer *VideoGeoAnalyzer
}

// NewGeoLocator returns a GeoLocator with default HTTP settings.
func NewGeoLocator() *GeoLocator {
	s := DefaultScraper()
	return &GeoLocator{
		scraper:       s,
		gmailFinder:   NewGmailFinder(s),
		apiClient:     NewAPIClient(s),
		videoAnalyzer: NewVideoGeoAnalyzer(s),
	}
}

// Analyze runs the full pipeline for the given TikTok handle or URL.
func (g *GeoLocator) Analyze(ctx context.Context, input string) (*GeoResult, error) {
	username := ParseUsername(input)
	if username == "" {
		return nil, fmt.Errorf("could not parse username from: %s", input)
	}

	result := &GeoResult{
		Username:      username,
		TikTokURL:     fmt.Sprintf("https://www.tiktok.com/@%s", username),
		CountryScores: make(map[string]float64),
	}

	// ── Step 1: TikTok profile ──────────────────────────────────────────────
	profile, err := g.scraper.FetchTikTokProfile(ctx, username)
	if err == nil {
		result.Profile = profile
		result.LinkedURLs = profile.Links
		g.scoreText(profile.Bio, "TikTok bio", 2.0, result)
		g.scoreText(profile.Nickname, "TikTok nickname", 1.5, result)
		g.scoreText(profile.Region, "TikTok region field", 3.0, result)
		g.scoreHashtags(profile.Hashtags, "TikTok profile hashtags", result)
		g.scoreText(strings.Join(profile.VideoTitles, " "), "video titles", 1.0, result)
	}

	// Username itself often carries regional clues
	g.scoreText(username, "TikTok username", 1.0, result)

	// ── Step 2: Web searches for the username ───────────────────────────────
	g.webSearchUsername(ctx, username, result)

	// ── Step 3: Linked pages ────────────────────────────────────────────────
	var extraLinks []string
	for _, link := range result.LinkedURLs {
		body, err := g.scraper.FetchPage(ctx, link)
		if err != nil {
			continue
		}
		g.scoreText(body, "linked page: "+link, 1.5, result)

		// Collect emails
		for _, email := range ExtractEmails(body) {
			extraLinks = append(extraLinks, email)
		}

		// Follow sub-links one level deep (linktree, allmylinks, etc.)
		for _, subLink := range reLinks.FindAllString(body, -1) {
			subLink = strings.TrimRight(subLink, ".,;!?")
			if isUsefulLink(subLink) && !contains(result.LinkedURLs, subLink) {
				result.LinkedURLs = append(result.LinkedURLs, subLink)
				subBody, err := g.scraper.FetchPage(ctx, subLink)
				if err == nil {
					g.scoreText(subBody, "sub-linked page: "+subLink, 1.0, result)
					for _, email := range ExtractEmails(subBody) {
						extraLinks = append(extraLinks, email)
					}
				}
			}
		}
	}

	// ── Step 4: Gmail discovery & analysis ─────────────────────────────────
	// Search pages we already have
	gmailMap := g.gmailFinder.DiscoverFromPages(ctx, result.LinkedURLs)
	// Also search DDG for any emails linked to the username
	g.searchEmailsForUsername(ctx, username, gmailMap, result)

	for email, info := range gmailMap {
		// Aggressive DDG search for the email
		clues := g.gmailFinder.SearchForEmail(ctx, email)
		info.CluesText = append(info.CluesText, clues...)

		// Look for linked accounts
		linked := g.gmailFinder.SearchLinkedAccounts(ctx, email)
		for _, l := range linked {
			body, err := g.scraper.FetchPage(ctx, l)
			if err == nil {
				g.scoreText(body, "linked account (via Gmail "+email+"): "+l, 1.5, result)
			}
		}

		// Score clues text
		for _, clue := range info.CluesText {
			g.scoreText(clue, "Gmail search result for "+email, 1.2, result)
		}

		er := EmailResult{
			Email:          email,
			FoundAt:        info.Sources,
			Clues:          info.CluesText,
			LinkedAccounts: linked,
		}
		result.Emails = append(result.Emails, er)
	}

	// ── Step 5: Video + comment geolocation ────────────────────────────────
	g.analyzeVideosAndComments(ctx, username, result)

	// ── Step 6: Finalise ────────────────────────────────────────────────────
	g.finalise(result)
	return result, nil
}

// analyzeVideosAndComments fetches the user's videos, extracts location tags
// and CDN server IPs, fetches comments on each video, logs comments left by
// the target user, and folds all signals into the score.
func (g *GeoLocator) analyzeVideosAndComments(ctx context.Context, username string, result *GeoResult) {
	// Resolve secUid (needed for the video-list API)
	secUID := ""
	if result.Profile != nil {
		secUID = extractSecUID(result.Profile.RawText)
	}
	if secUID == "" {
		var err error
		secUID, err = g.apiClient.FetchUserSecUID(ctx, username)
		if err != nil || secUID == "" {
			return // can't proceed without secUid
		}
	}

	// Fetch up to 30 most recent videos
	videos, err := g.apiClient.FetchVideos(ctx, secUID, 30)
	if err != nil || len(videos) == 0 {
		return
	}
	result.Videos = videos

	for _, video := range videos {
		// Score description / hashtags
		g.scoreText(video.Desc, fmt.Sprintf("video desc (id %s)", video.VideoID), 1.2, result)

		// Score the author region code if present
		if video.AuthorRegion != "" {
			g.scoreRegionCode(video.AuthorRegion, fmt.Sprintf("video author region (id %s)", video.VideoID), result)
		}

		// Score tagged location (POI)
		if video.POI != nil {
			loc := video.POI.Name + " " + video.POI.Address
			g.scoreText(loc, fmt.Sprintf("video location tag (id %s)", video.VideoID), 3.5, result)
		}

		// CDN IP resolution + GeoIP
		vgr := g.videoAnalyzer.Analyze(ctx, video)
		result.VideoGeos = append(result.VideoGeos, *vgr)

		if vgr.CDNToken != "" {
			country := CountryFromCDNToken(vgr.CDNToken)
			if country != "" {
				g.addSignal(result, country,
					fmt.Sprintf("CDN upload region %q (%s)", vgr.CDNToken, vgr.CDNRegion),
					fmt.Sprintf("video CDN (id %s)", video.VideoID), 2.0)
			}
		}
		if vgr.IPGeo != nil && vgr.IPGeo.Country != "" {
			g.addSignal(result, vgr.IPGeo.Country,
				fmt.Sprintf("CDN IP %s → %s, %s (ISP: %s)", vgr.CDNIP, vgr.IPGeo.City, vgr.IPGeo.Country, vgr.IPGeo.ISP),
				fmt.Sprintf("GeoIP lookup (video %s)", video.VideoID), 1.5)
		}

		// Fetch comments on this video (up to 100)
		comments, err := g.apiClient.FetchComments(ctx, video.VideoID, 100)
		if err != nil {
			continue
		}
		for i := range comments {
			comments[i].VideoID = video.VideoID
			if strings.EqualFold(comments[i].Username, username) {
				comments[i].IsOwner = true
				result.Comments = append(result.Comments, comments[i])
				// Score the text of own comments – high value signal
				g.scoreText(comments[i].Text,
					fmt.Sprintf("own comment on video %s", video.VideoID), 2.0, result)
			}
			// Also score region codes embedded in any commenter's profile
			if comments[i].Region != "" && comments[i].IsOwner {
				g.scoreRegionCode(comments[i].Region,
					fmt.Sprintf("commenter region field (video %s)", video.VideoID), result)
			}
		}
	}
}

// scoreRegionCode maps a 2-letter ISO country code to a country name and scores it.
func (g *GeoLocator) scoreRegionCode(code, source string, result *GeoResult) {
	country := isoToCountry(code)
	if country == "" {
		return
	}
	g.addSignal(result, country, fmt.Sprintf("ISO region code %q", strings.ToUpper(code)), source, 4.0)
}

// isoToCountry maps common ISO 3166-1 alpha-2 codes to country names.
func isoToCountry(code string) string {
	m := map[string]string{
		"JM": "Jamaica", "TT": "Trinidad and Tobago", "BB": "Barbados",
		"GY": "Guyana", "HT": "Haiti", "LC": "Saint Lucia",
		"VC": "Saint Vincent", "AG": "Antigua", "DM": "Dominica",
		"GD": "Grenada", "KN": "Saint Kitts", "BS": "Bahamas",
		"CU": "Cuba", "DO": "Dominican Republic", "PR": "Puerto Rico",
		"GB": "United Kingdom", "US": "United States", "CA": "Canada",
		"NG": "Nigeria", "GH": "Ghana", "KE": "Kenya",
		"ZA": "South Africa", "DE": "Germany", "FR": "France",
		"SG": "Singapore", "AU": "Australia", "NZ": "New Zealand",
		"BR": "Brazil", "MX": "Mexico", "IN": "India",
	}
	return m[strings.ToUpper(code)]
}

// ─── internal scoring helpers ───────────────────────────────────────────────

// scoreText runs all country patterns against a piece of text.
func (g *GeoLocator) scoreText(text, source string, baseWeight float64, result *GeoResult) {
	if text == "" {
		return
	}
	for _, cp := range Countries {
		w := baseWeight * cp.Weight

		if ok, matched := ContainsAny(text, cp.PhoneCodes); ok {
			g.addSignal(result, cp.Name, fmt.Sprintf("phone code %q", matched), source, w*3.0)
		}
		if ok, matched := ContainsAny(text, cp.Cities); ok {
			g.addSignal(result, cp.Name, fmt.Sprintf("city %q", matched), source, w*2.5)
		}
		if ok, matched := ContainsAny(text, cp.Keywords); ok {
			g.addSignal(result, cp.Name, fmt.Sprintf("keyword %q", matched), source, w*2.0)
		}
		if ok, matched := ContainsAny(text, cp.Slang); ok {
			g.addSignal(result, cp.Name, fmt.Sprintf("slang %q", matched), source, w*1.5)
		}
		if ok, matched := ContainsAny(text, cp.MusicGenres); ok {
			g.addSignal(result, cp.Name, fmt.Sprintf("music genre %q", matched), source, w*1.2)
		}
		if ok, matched := ContainsAny(text, cp.TLDs); ok {
			g.addSignal(result, cp.Name, fmt.Sprintf("TLD %q", matched), source, w*2.0)
		}
		if ok, matched := ContainsAny(text, cp.Currencies); ok {
			g.addSignal(result, cp.Name, fmt.Sprintf("currency %q", matched), source, w*2.5)
		}
		// Hashtags embedded in free text
		if ok, matched := ContainsAny(text, cp.Hashtags); ok {
			g.addSignal(result, cp.Name, fmt.Sprintf("hashtag %q", matched), source, w*1.8)
		}
	}

	// Phone number regex scan (any format)
	g.scorePhoneNumbers(text, source, baseWeight, result)
	// Email domain TLD scan
	g.scoreEmailTLDs(text, source, baseWeight, result)
}

var rePhone = regexp.MustCompile(`(?:\+|00)(\d{1,3})[\s\-\.]?\(?\d{1,4}\)?[\s\-\.]?\d{2,4}`)

// scorePhoneNumbers looks for E.164-style phone numbers and maps country codes.
func (g *GeoLocator) scorePhoneNumbers(text, source string, base float64, result *GeoResult) {
	for _, m := range rePhone.FindAllStringSubmatch(text, -1) {
		if len(m) < 2 {
			continue
		}
		cc := m[1]
		for _, cp := range Countries {
			for _, code := range cp.PhoneCodes {
				clean := strings.TrimPrefix(strings.TrimPrefix(code, "+"), "00")
				if strings.HasPrefix(cc, clean) || cc == clean {
					g.addSignal(result, cp.Name, fmt.Sprintf("phone number with country code +%s", cc), source, base*cp.Weight*2.5)
				}
			}
		}
	}
}

var reEmailAddr = regexp.MustCompile(`@[\w.\-]+\.([a-zA-Z]{2,})`)

// scoreEmailTLDs checks email TLDs in text (e.g. .jm = Jamaica).
func (g *GeoLocator) scoreEmailTLDs(text, source string, base float64, result *GeoResult) {
	for _, m := range reEmailAddr.FindAllStringSubmatch(text, -1) {
		if len(m) < 2 {
			continue
		}
		tld := "." + strings.ToLower(m[1])
		for _, cp := range Countries {
			for _, t := range cp.TLDs {
				if t == tld && tld != ".com" {
					g.addSignal(result, cp.Name, fmt.Sprintf("email TLD %s", tld), source, base*cp.Weight*2.0)
				}
			}
		}
	}
}

// scoreHashtags scores a list of #hashtag strings.
func (g *GeoLocator) scoreHashtags(tags []string, source string, result *GeoResult) {
	joined := strings.Join(tags, " ")
	for _, cp := range Countries {
		for _, ht := range cp.Hashtags {
			needle := "#" + strings.TrimPrefix(ht, "#")
			if strings.Contains(joined, needle) {
				g.addSignal(result, cp.Name, fmt.Sprintf("hashtag %q", needle), source, 1.8*cp.Weight)
			}
		}
	}
}

// webSearchUsername fires multiple targeted searches and scores the snippets.
func (g *GeoLocator) webSearchUsername(ctx context.Context, username string, result *GeoResult) {
	queries := []string{
		fmt.Sprintf(`tiktok @%s location country`, username),
		fmt.Sprintf(`"%s" tiktok jamaica OR caribbean OR trinidad OR barbados`, username),
		fmt.Sprintf(`tiktok @%s where are they from`, username),
		fmt.Sprintf(`@%s tiktok bio email phone`, username),
		fmt.Sprintf(`"%s" instagram OR twitter OR facebook location`, username),
	}
	for _, q := range queries {
		body, err := g.scraper.SearchDDG(ctx, q)
		if err != nil {
			continue
		}
		g.scoreText(body, "web search: "+q, 1.0, result)
		// Also harvest any emails/links from search results
		for _, email := range ExtractGmails(body) {
			result.Emails = append(result.Emails, EmailResult{Email: strings.ToLower(email), FoundAt: []string{"web search"}})
		}
	}
}

// searchEmailsForUsername searches DuckDuckGo for email addresses tied to the username.
func (g *GeoLocator) searchEmailsForUsername(ctx context.Context, username string, gmailMap map[string]*GmailInfo, result *GeoResult) {
	queries := []string{
		fmt.Sprintf(`"%s" gmail.com`, username),
		fmt.Sprintf(`tiktok @%s contact email`, username),
	}
	for _, q := range queries {
		body, err := g.scraper.SearchDDG(ctx, q)
		if err != nil {
			continue
		}
		for _, email := range ExtractGmails(body) {
			email = strings.ToLower(email)
			if _, ok := gmailMap[email]; !ok {
				gmailMap[email] = &GmailInfo{Email: email, Sources: []string{"web search"}}
			}
		}
	}
}

// addSignal records a scoring signal and updates the country tally.
func (g *GeoLocator) addSignal(result *GeoResult, country, evidence, source string, weight float64) {
	result.Signals = append(result.Signals, Signal{
		Country:  country,
		Evidence: evidence,
		Source:   source,
		Weight:   weight,
	})
	result.CountryScores[country] += weight
}

// finalise picks the best guess and labels confidence.
func (g *GeoLocator) finalise(result *GeoResult) {
	if len(result.CountryScores) == 0 {
		result.BestGuess = "Unknown"
		result.Confidence = "Low"
		return
	}

	type scored struct {
		country string
		score   float64
	}
	var ranked []scored
	for c, s := range result.CountryScores {
		ranked = append(ranked, scored{c, s})
	}
	sort.Slice(ranked, func(i, j int) bool { return ranked[i].score > ranked[j].score })

	result.BestGuess = ranked[0].country
	top := ranked[0].score

	var total float64
	for _, r := range ranked {
		total += r.score
	}

	pct := top / total
	switch {
	case pct >= 0.70:
		result.Confidence = "High"
	case pct >= 0.45:
		result.Confidence = "Medium"
	default:
		result.Confidence = "Low"
	}
}

// contains checks slice membership.
func contains(ss []string, s string) bool {
	for _, v := range ss {
		if v == s {
			return true
		}
	}
	return false
}
