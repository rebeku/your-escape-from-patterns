package client

type Pattern struct {
	ID               int              `json:"id"`
	Name             string           `json:"name"`
	DownloadLocation DownloadLocation `json:"download_location"`
}

type DownloadLocation struct {
	URL string `json:"url"`
}

type PatternDetailSearchResult struct {
	Patterns map[string]Pattern
}
