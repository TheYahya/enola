package enola

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"

	"github.com/theyahya/enola/internal/checker"
	"golang.org/x/sync/semaphore"
)

// Enola searches for a username across configured websites.
type Enola struct {
	Data   map[string]Website
	Site   string
	client *http.Client
	sem    *semaphore.Weighted
}

//go:embed data.json
var embeddedData []byte

// New loads the embedded site database and applies options.
func New(opts ...Option) (*Enola, error) {
	var data map[string]Website
	if err := json.Unmarshal(embeddedData, &data); err != nil {
		return nil, fmt.Errorf("%w", ErrDataFileIsNotAValidJson)
	}

	e := &Enola{
		Data:   data,
		client: defaultHTTPClient(),
		sem:    semaphore.NewWeighted(defaultConcurrency),
	}
	for _, opt := range opts {
		opt(e)
	}
	return e, nil
}

// SetSite filters checks to sites whose name contains the given string.
func (e *Enola) SetSite(site string) *Enola {
	e.Site = site
	return e
}

// ListCount returns the number of configured sites.
func (e *Enola) ListCount() int { return len(e.Data) }

// List returns the configured site database.
func (e *Enola) List() map[string]Website { return e.Data }

// Check probes all selected sites for the given username.
// The returned channel is closed when all probes finish.
func (e *Enola) Check(ctx context.Context, username string) (<-chan Result, error) {
	sites, err := e.selectSites()
	if err != nil {
		return nil, err
	}

	ch := make(chan Result)
	go func() {
		defer close(ch)
		var wg sync.WaitGroup
		for name, site := range sites {
			if err := e.sem.Acquire(ctx, 1); err != nil {
				return
			}
			wg.Add(1)
			go func(name string, site Website) {
				defer wg.Done()
				defer e.sem.Release(1)
				ch <- e.checkOne(ctx, name, site, username)
			}(name, site)
		}
		wg.Wait()
	}()
	return ch, nil
}

func (e *Enola) selectSites() (map[string]Website, error) {
	if e.Site == "" {
		return e.Data, nil
	}

	filter := strings.ToLower(strings.TrimSpace(e.Site))
	selected := make(map[string]Website)
	for name, site := range e.Data {
		if strings.Contains(strings.ToLower(name), filter) {
			selected[name] = site
		}
	}
	if len(selected) == 0 {
		return nil, fmt.Errorf("%w", ErrSiteNotFound)
	}
	return selected, nil
}

func (e *Enola) checkOne(ctx context.Context, name string, site Website, username string) Result {
	url := strings.ReplaceAll(site.URL, "{}", username)
	res := Result{Name: name, URL: url, Found: false, Status: StatusNotFound}

	if site.RegexCheck != "" {
		ok, err := matchUsernameRegex(site.RegexCheck, username)
		if err != nil {
			res.Status = StatusErrored
			return res
		}
		if !ok {
			res.Status = StatusInvalid
			return res
		}
	}

	detector, ok := checker.Lookup(site.ErrorType)
	if !ok {
		res.Status = StatusErrored
		return res
	}

	resp, err := e.probe(ctx, site, username)
	if err != nil {
		res.Status = StatusErrored
		return res
	}

	target := checker.Target{
		ErrorMessages: site.ErrorMessages,
		ErrorURL:      site.ErrorURL,
		ErrorCodes:    site.ErrorCodes,
	}
	found, err := detector.Detect(resp, target)
	if err != nil {
		res.Status = StatusErrored
		return res
	}
	if found {
		res.Found = true
		res.Status = StatusFound
	}
	return res
}

func (e *Enola) probe(ctx context.Context, site Website, username string) (checker.Response, error) {
	req, err := e.buildRequest(ctx, site, username)
	if err != nil {
		return checker.Response{}, err
	}

	httpResp, err := e.client.Do(req)
	if err != nil {
		return checker.Response{}, err
	}
	defer httpResp.Body.Close()

	body, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return checker.Response{}, err
	}

	finalURL := ""
	if httpResp.Request != nil && httpResp.Request.URL != nil {
		finalURL = httpResp.Request.URL.String()
	}

	return checker.Response{
		StatusCode: httpResp.StatusCode,
		Body:       body,
		FinalURL:   finalURL,
	}, nil
}

func (e *Enola) buildRequest(ctx context.Context, site Website, username string) (*http.Request, error) {
	probeURL := site.URLProbe
	if probeURL == "" {
		probeURL = site.URL
	}
	probeURL = strings.ReplaceAll(probeURL, "{}", username)

	method := site.RequestMethod
	if method == "" {
		method = http.MethodGet
	}

	var body io.Reader
	if site.RequestPayload != nil {
		payload := substituteUsername(site.RequestPayload, username)
		data, err := json.Marshal(payload)
		if err != nil {
			return nil, err
		}
		body = bytes.NewReader(data)
	}

	req, err := http.NewRequestWithContext(ctx, method, probeURL, body)
	if err != nil {
		return nil, err
	}
	for key, value := range site.Headers {
		req.Header.Set(key, value)
	}
	if site.RequestPayload != nil && req.Header.Get("Content-Type") == "" {
		req.Header.Set("Content-Type", "application/json")
	}
	return req, nil
}

func substituteUsername(value any, username string) any {
	switch v := value.(type) {
	case string:
		return strings.ReplaceAll(v, "{}", username)
	case map[string]any:
		out := make(map[string]any, len(v))
		for key, val := range v {
			out[key] = substituteUsername(val, username)
		}
		return out
	case []any:
		out := make([]any, len(v))
		for i, val := range v {
			out[i] = substituteUsername(val, username)
		}
		return out
	default:
		return value
	}
}
