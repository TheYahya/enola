package checker

// Response holds HTTP probe data passed to detectors.
type Response struct {
	StatusCode int
	Body       []byte
	FinalURL   string // resp.Request.URL.String() after redirects
}

// Target holds site-specific detection configuration.
type Target struct {
	ErrorMessages []string
	ErrorURL      string
	ErrorCodes    []int
}

// Detector decides whether a username exists on a site given a probe response.
type Detector interface {
	Detect(resp Response, target Target) (found bool, err error)
}
