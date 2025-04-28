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

func videosByServer(r *http.Request) (interface{}, error) {
	var err error
	if err := r.ParseForm(); err != nil {
		return nil, fmt.Errorf("error parsing form: %v", err)
	}
	anime := r.Form.Get("anime")
	episode := r.Form.Get("episode")

	client := utils.NewHTTPClient()

	res, err := client.R().Get(fmt.Sprintf("%v/%s/%s/", config.Rooturl, anime, episode))
	if err != nil {
		log.Fatal(err)
	}

	responseString := res.String()

	doc := soup.HTMLParse(responseString)

	var serversJSON string
	scriptRegex := regexp.MustCompile(`var servers = (\[.*?\]);`)
	for _, script := range doc.FindAll("script") {
		matches := scriptRegex.FindStringSubmatch(script.Text())
		if len(matches) > 1 {
			serversJSON = matches[1]
			break
		}
	}

	if serversJSON == "" {
		return nil, fmt.Errorf("no se encontró el JSON de servidores")
	}

	var serversData []struct {
		Remote string `json:"remote"`
		Server string `json:"server"`
		Lang   int    `json:"lang"`
		Size   string `json:"size"`
	}
	if err := json.Unmarshal([]byte(serversJSON), &serversData); err != nil {
		return nil, fmt.Errorf("error al decodificar el JSON de servidores: %v", err)
	}

	servers := []models.Server{}
	for _, server := range serversData {
		decodedURL, err := base64.StdEncoding.DecodeString(server.Remote)
		if err != nil {
			log.Printf("error al decodificar la URL en Base64: %v", err)
			continue
		}

		servers = append(servers, models.Server{
			Name: fmt.Sprintf("%s (%s)", server.Server, server.Size),
			Url:  string(decodedURL),
		})
	}

	return servers, nil
}
