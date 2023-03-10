package enola

import (
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"golang.org/x/sync/semaphore"
)

const RequestTimeout time.Duration = time.Second * 20

type Website struct {
	ErrorType         string      `json:"errorType"`
	ErrorMessage      interface{} `json:"errorMsg"`
	URL               string      `json:"url"`
	UrlMain           string      `json:"urlMain"`
	UsernameClaimed   string      `json:"username_claimed"`
	UsernameUnclaimed string      `json:"username_unclaimed"`
}

type Enola struct {
	Data map[string]Website
	Site string
	Ctx  context.Context
}

type Result struct {
	Name  string `json:"name"`
	URL   string `json:"url"`
	Found bool   `json:"found"`
}

//go:embed data.json
var d []byte

func New(ctx context.Context) (Enola, error) {
	var err error
	var data map[string]Website
	err = json.Unmarshal(d, &data)
	if err != nil {
		return Enola{}, errors.New(ErrDataFileIsNotAValidJson)
	}

	return Enola{
		Data: data,
		Ctx:  ctx,
	}, nil
}

func (s *Enola) SetSite(site string) *Enola {
	s.Site = site
	return s
}

func (s *Enola) ListCount() int           { return len(s.Data) }
func (s *Enola) List() map[string]Website { return s.Data }

func (s *Enola) Check(username string) (<-chan Result, error) {
	ch := make(chan Result)
	data := map[string]Website{}

	if s.Site != "" {
		for k, v := range s.Data {
			if strings.Contains(strings.ToLower(k), strings.Trim(strings.ToLower(s.Site), " ")) {
				data[k] = v
			}
		}

		// if site is not found in the list
		if len(data) == 0 {
			return nil, errors.New(ErrSiteNotFound)
		}
	}

	if len(data) == 0 {
		data = s.Data
	}

	ctx := context.Background()
	sem := semaphore.NewWeighted(20)

	go func() {
		for key, value := range data {
			if err := sem.Acquire(ctx, 1); err != nil {
				fmt.Println(err)
			}
			go func(key string, value Website) {
				defer sem.Release(1)
				url := strings.ReplaceAll(value.URL, "{}", username)

				res := Result{
					Name:  key,
					URL:   url,
					Found: false,
				}

				client := http.DefaultClient
				client.Timeout = RequestTimeout
				if value.ErrorType == "status_code" {
					resp, err := client.Get(url)
					if err != nil {
						ch <- res
						return
					}
					resp.Body.Close()

					if resp.StatusCode == http.StatusOK {
						res.Found = true
						ch <- res
						return
					}
					ch <- res
					return

				}

				if value.ErrorType == "message" {
					resp, err := client.Get(url)
					if err != nil {
						ch <- res
						return
					}
					defer resp.Body.Close()

					bodyBytes, err := io.ReadAll(resp.Body)
					if err != nil {
						ch <- res
						return
					}

					valueString, ok := value.ErrorMessage.(string)
					if !ok {
						ch <- res
						return
					}

					if !strings.Contains(string(bodyBytes), valueString) {
						res.Found = true
						ch <- res
						return
					}
					ch <- res
				}
			}(key, value)
		}
	}()

	return ch, nil
}
