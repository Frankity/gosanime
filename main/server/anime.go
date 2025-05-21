package server

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/anaskhan96/soup"
	"xyz.frankity/gosanime/main/config"
	"xyz.frankity/gosanime/main/models"
	"xyz.frankity/gosanime/main/utils"
)

// top fetches a list of top anime from the source.
// It scrapes the configured TopUrl to gather anime information.
// Returns a slice of models.Anime and an error if any issues occur during scraping or processing.
func top() ([]models.Anime, error) {
	animes := []models.Anime{}
	var err error

	client := utils.NewHTTPClient()

	res, err := client.R().Get(fmt.Sprintf("%v%s", config.Rooturl, config.TopUrl))
	if err != nil {
		// Using log.Fatal will exit the program, consider returning the error instead
		// for better error handling by the caller.
		log.Fatal(err)
	}

	responseString := res.String()

	doc := soup.HTMLParse(responseString)

	main := doc.Find("section", "class", "contenido").Find("div", "class", "container").Find("div", "class", "row").Find("div", "class", "col-lg-12").FindAll("div", "class", "list")

	for _, p := range main {
		parent := p.Find("div", "id", "conb")
		anime := models.Anime{
			ID:       strings.Split(parent.Find("a").Attrs()["href"], "/")[3],
			Name:     parent.Find("a").Attrs()["title"],
			Poster:   p.Find("a").Find("img").Attrs()["src"],
			Synopsis: strings.TrimSpace(strings.Split(p.Find("div", "id", "animinfo").Find("span", "class", "title").Text(), "/")[1]),
			State:    strings.TrimSpace(p.Find("div", "id", "animinfo").Find("h2", "class", "portada-title").Text()),
			Type:     strings.TrimSpace(strings.Split(p.Find("div", "id", "animinfo").Find("span", "class", "title").Text(), "/")[0]),
		}
		animes = append(animes, anime)
	}
	return animes, nil // err is always nil here due to log.Fatal above
}

// anime fetches detailed information for a specific anime based on its ID.
// It expects an 'id' query parameter in the http.Request.
// The ID is used to construct the URL for scraping the anime's details page.
// Returns an interface{} containing a models.Anime object and an error if any issues occur.
func anime(r *http.Request) (interface{}, error) {
	var err error

	if err := r.ParseForm(); err != nil {
		// Exiting here might be too drastic for a request handler.
		// Consider returning an error and letting the caller decide how to handle it.
		os.Exit(1)
	}
	x := r.Form.Get("id") // ID of the anime to fetch

	client := utils.NewHTTPClient()

	res, err := client.R().Get(fmt.Sprintf("%v/%s", config.Rooturl, x))
	if err != nil {
		log.Fatal(err) // Consider returning error
	}

	responseString := res.String()

	doc := soup.HTMLParse(responseString)

	var episodesData string

	episodes := doc.Find("div", "class", "anime__pagination").FindAll("a", "class", "numbers")

	episodeNumber := strings.TrimSpace(doc.Find("div", "class", "anime__details__widget").Find("ul").FindAll("li")[5].Text())

	lastEp := strings.Split(strings.TrimSpace(episodes[len(episodes)-1].Text()), "-")[1]

	genres := doc.Find("div", "class", "anime__details__widget").Find("ul").FindAll("li")[1].FindAll("a")

	result := []string{}

	for a := range genres {
		result = append(result, strings.ToLower(genres[a].Text()))
	}

	strings.Join(result, ",") // The result of Join is not assigned, this line has no effect.

	if episodeNumber == "Desconocido" {
		episodesData = lastEp
	} else {
		episodesData = episodeNumber
	}

	an := models.Anime{
		ID:       x,
		Name:     doc.Find("div", "class", "anime__details__content").Find("h3").Text(),
		Poster:   doc.Find("div", "class", "anime__details__pic").Attrs()["data-setbg"],
		Type:     strings.TrimSpace(doc.Find("div", "class", "anime__details__widget").Find("ul").FindAll("li")[0].Text()),
		Synopsis: doc.Find("div", "class", "anime__details__text").Find("p").Text(),
		Genre:    result,
		State:    "unknown", // State seems to be hardcoded or placeholder.
		Episodes: strings.TrimSpace(episodesData),
	}

	return an, err // err is likely nil here if previous errors cause fatal exits.
}

// searchAnime handles requests to search for anime.
// It expects 'anime' (search query) and 'page' (page number) as query parameters in the http.Request.
// It scrapes the search results from the configured URL.
// Returns an interface{} containing a models.SearchAnimeResponse object and an error if any.
func searchAnime(r *http.Request) (interface{}, error) {
	var err error

	if err := r.ParseForm(); err != nil {
		fmt.Println(err) // Consider logging instead of just printing.
		os.Exit(1)       // Consider returning an error.
	}
	animeQuery := r.Form.Get("anime") // The search term for anime.
	page := r.Form.Get("page")        // The page number for search results.

	url := fmt.Sprintf("%v/buscar/%s/%s/", config.Rooturl, strings.Replace(animeQuery, "-", "_", -1), page)

	// DefaultClient is used here, consider consistency with utils.NewHTTPClient()
	client := http.DefaultClient

	// The config.New round tripper might be for specific proxy or transport settings.
	client.Transport, err = config.New(client.Transport)
	if err != nil {
		log.Fatal(err) // Consider returning error.
	}

	res, err := client.Get(url)
	if err != nil {
		log.Fatal(err) // Consider returning error.
	}
	defer res.Body.Close() // Ensure the response body is closed.

	responseData, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Fatal(err) // Consider returning error.
	}

	responseString := string(responseData)

	doc := soup.HTMLParse(responseString)

	elements := doc.FindAll("div", "class", "anime__item")

	animes := []models.Anime{}

	for _, p := range elements {
		anime := models.Anime{
			ID:     strings.Split(p.Find("h5").Find("a").Attrs()["href"], "/")[3],
			Name:   p.Find("h5").Find("a").Text(),
			Poster: p.Find("a").Find("div", "class", "anime__item__pic").Attrs()["data-setbg"],
			State:  p.Find("div", "class", "anime__item__text").Find("ul").FindAll("li")[0].Text(),
			Type:   strings.TrimSpace(p.Find("div", "class", "anime__item__text").Find("ul").FindAll("li")[1].Text()),
		}
		animes = append(animes, anime)
	}

	pg, err := strconv.Atoi(page)
	if err != nil {
		// Printing to stdout in a server handler is generally not recommended.
		// Log this error or return it in the response.
		print(fmt.Sprintf("%v %s", err, "<- Error"))
		// If Atoi fails, pg will be 0, which might lead to incorrect page logic.
	}

	// Pagination logic: if 12 elements are found, assume there's a next page.
	// Otherwise, set page to -1 to indicate no more pages.
	if len(elements) == 12 {
		pg = pg + 1
	} else {
		pg = -1
	}

	ar := models.SearchAnimeResponse{
		Data:    animes,
		Status:  "200", // Status should ideally reflect actual HTTP status codes.
		Message: "Success",
		Page:    pg,
	}

	return ar, err // err might be from strconv.Atoi if not handled, or nil.
}
