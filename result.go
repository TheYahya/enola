package enola

// Status describes the outcome of a username check.
type Status int

const (
	StatusNotFound Status = iota
	StatusFound
	StatusInvalid
	StatusErrored
)

// Result is a single site check outcome.
type Result struct {
	Name   string `json:"name"`
	URL    string `json:"url"`
	Found  bool   `json:"found"`
	Status Status `json:"-"`
}
