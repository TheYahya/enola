package exporter

type ExportType string

const (
	JSON         ExportType = "json"
	CSV          ExportType = "csv"
	NOTSUPPORTED ExportType = "notsupported"
)

type Item struct {
	Title string `json:"title"`
	URL   string `json:"url"`
	Found bool   `json:"found"`
}

type Writer interface {
	Write()
}
