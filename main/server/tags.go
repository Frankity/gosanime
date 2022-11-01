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

func tags(r *http.Request) (interface{}, error) {

	if err := r.ParseForm(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	animes := []models.Anime{}
	tag := r.Form.Get("tag")
	page := r.Form.Get("page")

	client := http.DefaultClient

	http.DefaultClient.Transport = config.AddCloudFlareByPass(http.DefaultClient.Transport)

	res, err := client.Get(fmt.Sprintf("%v%s/%s/%s", config.Rooturl, config.Genreurl, tag, page))
	if err != nil {
		log.Fatal(err)
	}

	responseData, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Fatal(err)
	}

	responseString := string(responseData)

	doc := soup.HTMLParse(responseString)

	main := doc.FindAll("div", "class", "anime__item")

	for _, p := range main {
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
		print(fmt.Sprintf("%v %s", err, "<- Error"))
	}

	if len(main) == 24 {
		pg = pg + 1
	} else {
		pg = -1
	}

	ar := models.SearchAnimeResponse{
		Data:    animes,
		Status:  "200",
		Message: "Success",
		Page:    pg,
	}

	return ar, err
}
