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
)

// tags fetches a list of anime based on a specific tag and page number.
// It expects 'tag' and 'page' as query parameters in the provided http.Request.
// The function scrapes the source website using a URL constructed from these parameters.
// It returns an interface{} containing a models.SearchAnimeResponse, which includes
// the list of anime, pagination information, and status. An error is returned if issues
// occur during parameter parsing, HTTP requests, or data processing.
//
// Parameters:
//   r: The *http.Request containing 'tag' and 'page' query parameters.
//
// Returns:
//   An interface{} which will be a models.SearchAnimeResponse on success.
//   An error if any step of the process fails.
//
// Note: The function uses os.Exit(1) or log.Fatal in case of certain errors,
// which will terminate the application. Ideally, errors should be returned to the caller.
func tags(r *http.Request) (interface{}, error) {
	var err error
	if err := r.ParseForm(); err != nil {
		fmt.Println(err) // Consider logging instead of printing to stdout.
		// os.Exit(1) is generally not recommended in HTTP handlers; returning an error is preferred.
		return nil, fmt.Errorf("error parsing form: %w", err)
	}

	animes := []models.Anime{}
	tag := r.Form.Get("tag")   // Tag to filter anime by.
	page := r.Form.Get("page") // Page number for pagination.

	// Using http.DefaultClient. Consider consistency with utils.NewHTTPClient() if used elsewhere.
	client := http.DefaultClient

	// config.New is likely a custom transport or round tripper.
	client.Transport, err = config.New(client.Transport)
	if err != nil {
		log.Fatal(err) // Terminates application; consider returning error.
	}
	// Constructs URL like: {Rooturl}{Genreurl}/{tag}/{page}
	res, err := client.Get(fmt.Sprintf("%v%s/%s/%s", config.Rooturl, config.Genreurl, tag, page))
	if err != nil {
		log.Fatal(err) // Terminates application; consider returning error.
	}
	defer res.Body.Close() // Ensure response body is closed.

	responseData, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Fatal(err) // Terminates application; consider returning error.
	}

	responseString := string(responseData)
	doc := soup.HTMLParse(responseString)

	mainContent := doc.FindAll("div", "class", "anime__item")

	for _, p := range mainContent {
		anime := models.Anime{
			ID:     strings.Split(p.Find("h5").Find("a").Attrs()["href"], "/")[3],
			Name:   p.Find("h5").Find("a").Text(),
			Poster: p.Find("a").Find("div", "class", "anime__item__pic").Attrs()["data-setbg"],
			State:  p.Find("div", "class", "anime__item__text").Find("ul").FindAll("li")[0].Text(),
			Type:   strings.TrimSpace(p.Find("div", "class", "anime__item__text").Find("ul").FindAll("li")[1].Text()),
			// Synopsis, Genre, Episodes are not populated for these anime items.
		}
		animes = append(animes, anime)
	}

	pg, errConv := strconv.Atoi(page)
	if errConv != nil {
		// Printing to stdout is not ideal for server applications. Log or return error.
		print(fmt.Sprintf("%v %s", errConv, "<- Error converting page to int"))
		// If Atoi fails, pg will be 0. This might impact pagination logic.
		// The original 'err' variable is shadowed by the loop variable 'err' if it existed.
		// Assigning errConv to err to return it.
		err = fmt.Errorf("error converting page to integer: %w", errConv)
	}

	// Pagination logic: if 24 items are found (specific to source site layout),
	// assume there is a next page. Otherwise, mark as last page (-1).
	if len(mainContent) == 24 {
		pg = pg + 1
	} else {
		pg = -1
	}

	ar := models.SearchAnimeResponse{
		Data:    animes,
		Status:  "200", // Consider using actual HTTP status codes.
		Message: "Success",
		Page:    pg,
	}

	return ar, err // Returns the original Atoi conversion error if it occurred.
}
