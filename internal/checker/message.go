package checker

import "strings"

func init() {
	Register("message", func() Detector { return messageDetector{} })
}

type messageDetector struct{}

func (messageDetector) Detect(resp Response, target Target) (bool, error) {
	body := string(resp.Body)
	for _, msg := range target.ErrorMessages {
		if strings.Contains(body, msg) {
			return false, nil
		}
	}
	return true, nil
}
