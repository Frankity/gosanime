package server

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"regexp"

	"github.com/anaskhan96/soup"
	"xyz.frankity/gosanime/main/config"
	"xyz.frankity/gosanime/main/models"
	"xyz.frankity/gosanime/main/utils"
)

// videosByServer fetches available video server links for a specific anime episode.
// It expects 'anime' (anime slug) and 'episode' (episode number) as query parameters
// in the provided http.Request.
// The function scrapes the anime episode page, extracts a JSON string containing server
// information from a script tag, decodes it, and then decodes the Base64 encoded
// video URLs.
//
// Parameters:
//   r: The *http.Request containing 'anime' and 'episode' query parameters.
//
// Returns:
//   An interface{} which will be a slice of models.Server on success. Each Server
//   struct contains the name of the server and the direct video URL.
//   An error if any step of the process fails (e.g., parsing form, HTTP request,
//   JSON extraction/unmarshalling, Base64 decoding).
//
// Note: Uses log.Fatal on HTTP request errors, which will terminate the application.
// Errors related to JSON or Base64 decoding are logged but the function continues if possible.
func videosByServer(r *http.Request) (interface{}, error) {
	var err error
	if err := r.ParseForm(); err != nil {
		return nil, fmt.Errorf("error parsing form: %w", err)
	}
	animeSlug := r.Form.Get("anime")   // Anime slug, e.g., "spy-x-family/"
	episode := r.Form.Get("episode") // Episode number, e.g., "1"

	client := utils.NewHTTPClient()

	// Constructs URL like: {Rooturl}/{animeSlug}/{episode}/
	res, err := client.R().Get(fmt.Sprintf("%v/%s/%s/", config.Rooturl, animeSlug, episode))
	if err != nil {
		log.Fatal(err) // Terminates application. Consider returning error.
	}
	// Note: res.Body is not explicitly closed here. The NewHTTPClient or Resty might handle it.

	responseString := res.String()
	doc := soup.HTMLParse(responseString)

	var serversJSON string
	// Regex to find 'var servers = [...];' in script tags.
	scriptRegex := regexp.MustCompile(`var servers = (\[.*?\]);`)
	for _, script := range doc.FindAll("script") {
		matches := scriptRegex.FindStringSubmatch(script.Text())
		if len(matches) > 1 {
			serversJSON = matches[1] // Extract the JSON array string.
			break
		}
	}

	if serversJSON == "" {
		return nil, fmt.Errorf("no se encontró el JSON de servidores") // "Servers JSON not found"
	}

	// serversData is used to temporarily hold the unmarshalled JSON from the script.
	var serversData []struct {
		Remote string `json:"remote"` // Base64 encoded URL
		Server string `json:"server"` // Server name (e.g., "MEGA", "Stape")
		Lang   int    `json:"lang"`   // Language ID (not used in final output)
		Size   string `json:"size"`   // Video size/quality (e.g., "HD", "SD")
	}
	if err := json.Unmarshal([]byte(serversJSON), &serversData); err != nil {
		return nil, fmt.Errorf("error al decodificar el JSON de servidores: %w", err) // "Error decoding servers JSON"
	}

	servers := []models.Server{}
	for _, serverData := range serversData {
		decodedURLBytes, err := base64.StdEncoding.DecodeString(serverData.Remote)
		if err != nil {
			// Log the error and skip this server, rather than failing the whole request.
			log.Printf("error al decodificar la URL en Base64 para el servidor %s: %v", serverData.Server, err)
			continue
		}

		servers = append(servers, models.Server{
			Name: fmt.Sprintf("%s (%s)", serverData.Server, serverData.Size), // e.g., "MEGA (HD)"
			Url:  string(decodedURLBytes),
		})
	}

	return servers, nil
}
