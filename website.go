package enola

import (
	"encoding/json"
	"fmt"
)

// Website holds per-site configuration from data.json.
type Website struct {
	ErrorType         string            `json:"errorType"`
	ErrorMessages     []string          `json:"-"`
	ErrorURL          string            `json:"errorUrl"`
	ErrorCodes        []int             `json:"-"`
	URL               string            `json:"url"`
	URLProbe          string            `json:"urlProbe"`
	URLMain           string            `json:"urlMain"`
	RegexCheck        string            `json:"regexCheck"`
	RequestMethod     string            `json:"request_method"`
	RequestPayload    map[string]any    `json:"request_payload"`
	Headers           map[string]string `json:"headers"`
	IsNSFW            bool              `json:"isNSFW"`
	UsernameClaimed   string            `json:"username_claimed"`
	UsernameUnclaimed string            `json:"username_unclaimed"`
}

type websiteJSON struct {
	ErrorType         string            `json:"errorType"`
	ErrorMsg          json.RawMessage   `json:"errorMsg"`
	ErrorURL          string            `json:"errorUrl"`
	ErrorCode         json.RawMessage   `json:"errorCode"`
	URL               string            `json:"url"`
	URLProbe          string            `json:"urlProbe"`
	URLMain           string            `json:"urlMain"`
	RegexCheck        string            `json:"regexCheck"`
	RequestMethod     string            `json:"request_method"`
	RequestPayload    map[string]any    `json:"request_payload"`
	Headers           map[string]string `json:"headers"`
	IsNSFW            bool              `json:"isNSFW"`
	UsernameClaimed   string            `json:"username_claimed"`
	UsernameUnclaimed string            `json:"username_unclaimed"`
}

// UnmarshalJSON normalizes errorMsg (string | []string) and errorCode (int | []int).
func (w *Website) UnmarshalJSON(data []byte) error {
	var raw websiteJSON
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	msgs, err := parseErrorMessages(raw.ErrorMsg)
	if err != nil {
		return fmt.Errorf("errorMsg: %w", err)
	}
	codes, err := parseErrorCodes(raw.ErrorCode)
	if err != nil {
		return fmt.Errorf("errorCode: %w", err)
	}

	w.ErrorType = raw.ErrorType
	w.ErrorMessages = msgs
	w.ErrorURL = raw.ErrorURL
	w.ErrorCodes = codes
	w.URL = raw.URL
	w.URLProbe = raw.URLProbe
	w.URLMain = raw.URLMain
	w.RegexCheck = raw.RegexCheck
	w.RequestMethod = raw.RequestMethod
	w.RequestPayload = raw.RequestPayload
	w.Headers = raw.Headers
	w.IsNSFW = raw.IsNSFW
	w.UsernameClaimed = raw.UsernameClaimed
	w.UsernameUnclaimed = raw.UsernameUnclaimed
	return nil
}

func parseErrorMessages(raw json.RawMessage) ([]string, error) {
	if len(raw) == 0 {
		return nil, nil
	}
	var s string
	if err := json.Unmarshal(raw, &s); err == nil {
		return []string{s}, nil
	}
	var list []string
	if err := json.Unmarshal(raw, &list); err != nil {
		return nil, err
	}
	return list, nil
}

func parseErrorCodes(raw json.RawMessage) ([]int, error) {
	if len(raw) == 0 {
		return nil, nil
	}
	var n int
	if err := json.Unmarshal(raw, &n); err == nil {
		return []int{n}, nil
	}
	var list []int
	if err := json.Unmarshal(raw, &list); err != nil {
		return nil, err
	}
	return list, nil
}
