package checker

import "net/http"

func init() {
	Register("status_code", func() Detector { return statusCodeDetector{} })
}

type statusCodeDetector struct{}

func (statusCodeDetector) Detect(resp Response, target Target) (bool, error) {
	for _, code := range target.ErrorCodes {
		if resp.StatusCode == code {
			return false, nil
		}
	}
	return resp.StatusCode == http.StatusOK, nil
}
