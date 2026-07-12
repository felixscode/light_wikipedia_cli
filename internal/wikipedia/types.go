package wikipedia

import "encoding/json"

type SearchResult struct {
	Title   string `json:"title"`
	Snippet string `json:"snippet"`
}

type Summary struct {
	Title   string `json:"title"`
	Extract string `json:"extract"`
	URL     string `json:"url"`
}

type Page struct {
	Title   string `json:"title"`
	URL     string `json:"url"`
	Content string `json:"content"`
}

type actionSearchResponse struct {
	Query struct {
		Search []struct {
			Title     string `json:"title"`
			Snippet   string `json:"snippet"`
			PageID    int    `json:"pageid"`
			WordCount int    `json:"wordcount"`
			Size      int    `json:"size"`
		} `json:"search"`
	} `json:"query"`
}

type summaryResponse struct {
	Title       string `json:"title"`
	Extract     string `json:"extract"`
	URL         string
	Description string `json:"description"`
	PageID      int    `json:"pageid"`
}

type contentURLs struct {
	Desktop struct {
		Page string `json:"page"`
	} `json:"desktop"`
	Mobile struct {
		Page string `json:"page"`
	} `json:"mobile"`
}

func (sr *summaryResponse) UnmarshalJSON(data []byte) error {
	var raw struct {
		Title       string      `json:"title"`
		Extract     string      `json:"extract"`
		URLs        contentURLs `json:"content_urls"`
		Description string      `json:"description"`
		PageID      int         `json:"pageid"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	sr.Title = raw.Title
	sr.Extract = raw.Extract
	sr.URL = raw.URLs.Desktop.Page
	sr.Description = raw.Description
	sr.PageID = raw.PageID
	return nil
}
