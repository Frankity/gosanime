package tests

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
)

const (
	baseURL   = "http://localhost:3000"
	authToken = "Bearer 260716eba2c68ba3a15dece65337f99c"
)

type SearchAnimeResponse struct {
	Data    []Anime     `json:"data"`
	Status  string      `json:"status"`
	Message string      `json:"message"`
	Page    interface{} `json:"page"`
}

type Anime struct {
	ID       string   `json:"id"`
	Name     string   `json:"name"`
	Poster   string   `json:"poster"`
	State    string   `json:"state"`
	Type     string   `json:"type"`
	Synopsis string   `json:"synopsis"`
	Genre    []string `json:"genre"`
	Episodes string   `json:"episodes"`
}

func TestSearchAnime(t *testing.T) {
	url := fmt.Sprintf("%s/api/v1/search?anime=naruto&page=1", baseURL)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	req.Header.Set("Authorization", authToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("request failed (is the server running at %s?): %v", baseURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	var result SearchAnimeResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	t.Logf("status:  %s", result.Status)
	t.Logf("message: %s", result.Message)
	t.Logf("page:    %v", result.Page)
	t.Logf("results: %d anime(s) found", len(result.Data))

	for i, anime := range result.Data {
		t.Logf("[%d] %s (id: %s, type: %s, state: %s)", i+1, anime.Name, anime.ID, anime.Type, anime.State)
	}

	if len(result.Data) == 0 {
		t.Error("expected at least one result for 'naruto'")
	}
}
