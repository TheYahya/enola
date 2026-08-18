package tiktok

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

// ─── API response types ───────────────────────────────────────────────────────

type apiUserResponse struct {
	UserInfo struct {
		User struct {
			ID        string `json:"id"`
			UniqueID  string `json:"uniqueId"`
			SecUID    string `json:"secUid"`
			Nickname  string `json:"nickname"`
			Signature string `json:"signature"` // bio
			Region    string `json:"region"`
		} `json:"user"`
	} `json:"userInfo"`
	StatusCode int `json:"status_code"`
}

type apiVideoListResponse struct {
	// TikTok web API uses "itemList"
	ItemList []apiVideoItem `json:"itemList"`
	// Mobile/fallback uses "aweme_list"
	AwemeList []apiVideoItem `json:"aweme_list"`
	HasMore   bool           `json:"hasMore"`
	Cursor    int64          `json:"cursor"`
	StatusCode int           `json:"status_code"`
}

type apiVideoItem struct {
	ID         string `json:"id"`
	AwemeID    string `json:"aweme_id"`  // mobile API field name
	Desc       string `json:"desc"`
	CreateTime int64  `json:"createTime"`

	Author struct {
		UniqueID string `json:"uniqueId"`
		Region   string `json:"region"`
	} `json:"author"`

	// Location sticker / POI
	POI *struct {
		Name      string  `json:"name"`
		Address   string  `json:"address"`
		Latitude  float64 `json:"latitude"`
		Longitude float64 `json:"longitude"`
	} `json:"poi"`

	// Some API versions embed location here
	StickersOnItem []struct {
		StickerType int      `json:"stickerType"`
		StickerText []string `json:"stickerText"`
	} `json:"stickersOnItem"`

	Video struct {
		// Web API
		PlayAddr    string `json:"playAddr"`
		DownloadAddr string `json:"downloadAddr"`
		// Mobile API
		PlayAddrObj struct {
			URLList []string `json:"url_list"`
		} `json:"play_addr"`
		DownloadAddrObj struct {
			URLList []string `json:"url_list"`
		} `json:"download_addr"`
	} `json:"video"`

	// Mobile API location
	PoiInfo *struct {
		PoiName string  `json:"poi_name"`
		Address string  `json:"address"`
		Latitude float64 `json:"latitude"`
		Longitude float64 `json:"longitude"`
	} `json:"poi_info"`
}

type apiCommentListResponse struct {
	Comments   []apiComment `json:"comments"`
	HasMore    int          `json:"has_more"`
	Cursor     int64        `json:"cursor"`
	StatusCode int          `json:"status_code"`
}

type apiComment struct {
	CID        string `json:"cid"`
	Text       string `json:"text"`
	CreateTime int64  `json:"create_time"`
	User       struct {
		UID      string `json:"uid"`
		UniqueID string `json:"unique_id"`
		Nickname string `json:"nickname"`
		Region   string `json:"region"`
	} `json:"user"`
	ReplyCommentTotal int `json:"reply_comment_total"`
}

// ─── Exported types ───────────────────────────────────────────────────────────

// VideoItem is a normalised TikTok video.
type VideoItem struct {
	VideoID    string
	Desc       string
	CreateTime int64
	POI        *POIInfo // location tag, if any
	PlayURLs   []string
	AuthorRegion string
}

// POIInfo holds a tagged location.
type POIInfo struct {
	Name      string
	Address   string
	Latitude  float64
	Longitude float64
}

// CommentItem is a normalised TikTok comment.
type CommentItem struct {
	CID        string
	VideoID    string // video the comment was made on
	Text       string
	Username   string
	Nickname   string
	Region     string
	CreateTime int64
	IsOwner    bool // true when the comment was left by the account under analysis
}

// ─── API client ───────────────────────────────────────────────────────────────

// APIClient calls TikTok's unofficial web endpoints.
type APIClient struct {
	scraper *Scraper
}

// NewAPIClient creates an API client using the given scraper.
func NewAPIClient(scraper *Scraper) *APIClient {
	return &APIClient{scraper: scraper}
}

// FetchUserSecUID fetches the secUid needed for the video-list endpoint.
// It tries the user-detail API first, then falls back to regex on the profile page.
func (c *APIClient) FetchUserSecUID(ctx context.Context, username string) (string, error) {
	// Try the unofficial user-detail API
	apiURL := fmt.Sprintf(
		"https://www.tiktok.com/api/user/detail/?uniqueId=%s&aid=1988&app_language=en&device_platform=web_pc",
		username,
	)
	body, err := c.scraper.fetchWithBrowserHeaders(ctx, apiURL)
	if err == nil {
		var resp apiUserResponse
		if json.Unmarshal([]byte(body), &resp) == nil && resp.UserInfo.User.SecUID != "" {
			return resp.UserInfo.User.SecUID, nil
		}
	}

	// Fallback: scrape the profile page
	pageBody, err := c.scraper.fetchWithBrowserHeaders(ctx,
		fmt.Sprintf("https://www.tiktok.com/@%s", username))
	if err != nil {
		return "", err
	}
	return extractSecUID(pageBody), nil
}

var reSecUID = regexp.MustCompile(`"secUid"\s*:\s*"(MS4[^"]{20,})"`)

func extractSecUID(html string) string {
	if m := reSecUID.FindStringSubmatch(html); len(m) > 1 {
		return m[1]
	}
	return ""
}

// FetchVideos retrieves up to maxVideos of the user's TikToks.
func (c *APIClient) FetchVideos(ctx context.Context, secUID string, maxVideos int) ([]VideoItem, error) {
	if maxVideos <= 0 {
		maxVideos = 20
	}
	var all []VideoItem
	cursor := int64(0)

	for len(all) < maxVideos {
		count := 30
		if remaining := maxVideos - len(all); remaining < count {
			count = remaining
		}
		apiURL := fmt.Sprintf(
			"https://www.tiktok.com/api/post/item_list/?secUid=%s&count=%d&cursor=%d&aid=1988&app_language=en&device_platform=web_pc",
			secUID, count, cursor,
		)
		body, err := c.scraper.fetchWithBrowserHeaders(ctx, apiURL)
		if err != nil {
			break
		}
		var resp apiVideoListResponse
		if err := json.Unmarshal([]byte(body), &resp); err != nil {
			break
		}

		items := resp.ItemList
		if len(items) == 0 {
			items = resp.AwemeList
		}
		if len(items) == 0 {
			break
		}

		for _, raw := range items {
			all = append(all, normaliseVideo(raw))
		}

		if !resp.HasMore {
			break
		}
		cursor = resp.Cursor
	}
	return all, nil
}

// FetchComments retrieves up to maxComments comments for a single video.
func (c *APIClient) FetchComments(ctx context.Context, videoID string, maxComments int) ([]CommentItem, error) {
	if maxComments <= 0 {
		maxComments = 100
	}
	var all []CommentItem
	cursor := int64(0)

	for len(all) < maxComments {
		count := 50
		if remaining := maxComments - len(all); remaining < count {
			count = remaining
		}
		apiURL := fmt.Sprintf(
			"https://www.tiktok.com/api/comment/list/?aweme_id=%s&count=%d&cursor=%d&aid=1988&app_language=en&device_platform=web_pc",
			videoID, count, cursor,
		)
		body, err := c.scraper.fetchWithBrowserHeaders(ctx, apiURL)
		if err != nil {
			break
		}
		var resp apiCommentListResponse
		if err := json.Unmarshal([]byte(body), &resp); err != nil {
			break
		}
		if len(resp.Comments) == 0 {
			break
		}
		for _, raw := range resp.Comments {
			all = append(all, CommentItem{
				CID:        raw.CID,
				Text:       raw.Text,
				Username:   raw.User.UniqueID,
				Nickname:   raw.User.Nickname,
				Region:     raw.User.Region,
				CreateTime: raw.CreateTime,
			})
		}
		if resp.HasMore == 0 {
			break
		}
		cursor = resp.Cursor
	}
	return all, nil
}

// normaliseVideo converts the raw API struct into a clean VideoItem.
func normaliseVideo(raw apiVideoItem) VideoItem {
	id := raw.ID
	if id == "" {
		id = raw.AwemeID
	}

	v := VideoItem{
		VideoID:      id,
		Desc:         raw.Desc,
		CreateTime:   raw.CreateTime,
		AuthorRegion: raw.Author.Region,
	}

	// Collect play URLs
	if raw.Video.PlayAddr != "" {
		v.PlayURLs = append(v.PlayURLs, raw.Video.PlayAddr)
	}
	v.PlayURLs = append(v.PlayURLs, raw.Video.PlayAddrObj.URLList...)
	if raw.Video.DownloadAddr != "" {
		v.PlayURLs = append(v.PlayURLs, raw.Video.DownloadAddr)
	}
	v.PlayURLs = append(v.PlayURLs, raw.Video.DownloadAddrObj.URLList...)
	v.PlayURLs = dedupe(v.PlayURLs)

	// POI / location sticker
	if raw.POI != nil && raw.POI.Name != "" {
		v.POI = &POIInfo{
			Name:      raw.POI.Name,
			Address:   raw.POI.Address,
			Latitude:  raw.POI.Latitude,
			Longitude: raw.POI.Longitude,
		}
	} else if raw.PoiInfo != nil && raw.PoiInfo.PoiName != "" {
		v.POI = &POIInfo{
			Name:      raw.PoiInfo.PoiName,
			Address:   raw.PoiInfo.Address,
			Latitude:  raw.PoiInfo.Latitude,
			Longitude: raw.PoiInfo.Longitude,
		}
	} else {
		// Try location stickers (type 15 = location)
		for _, sticker := range raw.StickersOnItem {
			if sticker.StickerType == 15 && len(sticker.StickerText) > 0 {
				v.POI = &POIInfo{Name: strings.Join(sticker.StickerText, ", ")}
				break
			}
		}
	}

	return v
}
