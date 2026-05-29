package checker

import (
	"net/http"
	"strings"
)

func init() {
	Register("response_url", func() Detector { return responseURLDetector{} })
}

type responseURLDetector struct{}

func (responseURLDetector) Detect(resp Response, target Target) (bool, error) {
	if target.ErrorURL == "" {
		return resp.StatusCode == http.StatusOK, nil
	}
	return !strings.HasPrefix(resp.FinalURL, target.ErrorURL), nil
}
