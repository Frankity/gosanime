package server

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/anaskhan96/soup"
	"xyz.frankity/gosanime/main/models"
)

var bearer string

func lastAnimes() ([]models.Anime, error) {
	animes := []models.Anime{}
	resp, err := soup.Get(ROOTURL.url)
	if err != nil {
		os.Exit(1)
	}

	doc := soup.HTMLParse(resp)

	main := doc.Find("div", "class", "trending__anime").FindAll("div", "class", "anime__item")

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
	return animes, nil
}

func ovas() ([]models.Anime, error) {
	animes := []models.Anime{}
	resp, err := soup.Get(fmt.Sprintf("%v%s", ROOTURL.url, OVASURL.url))
	if err != nil {
		os.Exit(1)
	}

	doc := soup.HTMLParse(resp)

	main := doc.Find("div", "class", "anime__page__content").FindAll("div", "class", "anime__item")

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

	return animes, nil
}

func anime(r *http.Request) (interface{}, error) {

	if err := r.ParseForm(); err != nil {
		os.Exit(1)
	}
	x := r.Form.Get("id")

	resp, err := soup.Get(fmt.Sprintf("%v/%s", ROOTURL.url, x))
	if err != nil {
		os.Exit(1)
	}

	print(fmt.Sprintf("%v/%s", ROOTURL.url, x))

	doc := soup.HTMLParse(resp)

	episodes := doc.Find("div", "class", "anime__pagination").FindAll("a", "class", "numbers")

	lastEp := strings.Split(strings.TrimSpace(episodes[len(episodes)-1].Text()), "-")[1]
	lastEpIntVal, err := strconv.Atoi(strings.TrimSpace(lastEp))

	eplist := []models.Episode{}
	for i := 0; i < lastEpIntVal; i++ {
		ep := models.Episode{
			Id:      strconv.Itoa(i),
			Episode: x,
		}
		eplist = append(eplist, ep)
	}

	an := models.Anime{
		ID:       x,
		Name:     doc.Find("div", "class", "anime__details__content").Find("h3").Text(),
		Poster:   doc.Find("div", "class", "anime__details__pic").Attrs()["data-setbg"],
		Type:     strings.TrimSpace(doc.Find("div", "class", "anime__details__widget").Find("ul").FindAll("li")[0].Text()),
		Synopsis: doc.Find("div", "class", "anime__details__content").Find("p").Text(),
		State:    strings.TrimSpace(doc.Find("div", "class", "anime__details__widget").Find("ul").FindAll("li")[6].Find("span", "class", "enemision").Text()),
		Episodes: eplist,
	}

	return an, err
}

func videosByServer(r *http.Request) (interface{}, error) {
	if err := r.ParseForm(); err != nil {
		os.Exit(1)
	}
	anime := r.Form.Get("anime")
	episode := r.Form.Get("episode")

	resp, err := soup.Get(fmt.Sprintf("%v/%s/%s/", ROOTURL.url, anime, episode))
	if err != nil {
		os.Exit(1)
	}

	doc := soup.HTMLParse(resp)

	urls := []string{}
	for _, in := range doc.FindAll("script") {
		if strings.Contains(in.Text(), "var video = [];") {
			d := in.Children()[0]
			arr := strings.Split(d.NodeValue, "\n")
			for i := 0; i < len(arr); i++ {
				if strings.Contains(arr[i], "player_conte") {

					html := soup.HTMLParse(arr[i])
					urli := html.Find("iframe", "class", "player_conte").Attrs()["src"]

					if strings.Contains(urli, "jk.php") {
						ur := strings.Replace(urli, "jk.php?u=", "", -1)
						urls = append(urls, getHiddenUrl(ur))
						break
					}

					doc, err := soup.Get(urli)

					if err != nil {
						os.Exit(1)
					}

					datas := soup.HTMLParse(doc)

					if strings.Contains(datas.HTML(), "input") {

						redirUrl := datas.FindAll("input")[0].Attrs()["value"]

						data := url.Values{}
						data.Set("data", redirUrl)

						client := &http.Client{}
						r, err := http.NewRequest("POST", "https://jkanime.net/gsplay/redirect_post.php", strings.NewReader(data.Encode())) // URL-encoded payload
						if err != nil {
							log.Fatal(err)
						}
						r.Header.Add("Content-Type", "application/x-www-form-urlencoded")
						r.Header.Add("Content-Length", strconv.Itoa(len(data.Encode())))

						res, err := client.Do(r)
						if err != nil {
							log.Fatal(err)
						}

						defer res.Body.Close()

						if err != nil {
							log.Fatal(err)
						}

						vUrl := strings.Split(string(fmt.Sprint(res.Request)), " ")[1]

						if strings.Contains(vUrl, "#") {
							hash := strings.Split(vUrl, "#")

							d := url.Values{}
							d.Set("v", hash[1])

							p, err := http.PostForm("https://jkanime.net/gsplay/api.php", url.Values{"v": {hash[1]}})
							if err != nil {
								log.Fatal(err)
							}

							r.Header.Add("Content-Type", "application/x-www-form-urlencoded")
							r.Header.Add("Content-Length", strconv.Itoa(len(d.Encode())))

							if nil != err {
								fmt.Println("error in action happened getting the response", err)
							}

							defer p.Body.Close()
							body, err := ioutil.ReadAll(p.Body)

							if nil != err {
								fmt.Println("error in acion happened reading the body", err)
							}

							var url Url
							_ = json.Unmarshal([]byte(body), &url)

							urls = append(urls, url.File)

						} else {
							urls = append(urls, *&vUrl)
						}
					}
				}
			}
		}
	}

	servers := []models.Server{}
	for i := 0; i < len(urls); i++ {
		s := models.Server{
			Name: fmt.Sprintf("Servidor %v", i+1),
			Url:  urls[i],
		}
		servers = append(servers, s)
	}

	return servers, err
}

func searchAnime(r *http.Request) (interface{}, error) {
	if err := r.ParseForm(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	anime := r.Form.Get("anime")
	page := r.Form.Get("page")

	url := fmt.Sprintf("%v/buscar/%s/%s/", ROOTURL.url, strings.Replace(anime, "-", "_", -1), page)

	resp, err := soup.Get(url)
	if err != nil {
		fmt.Println(err)
		return err, err
	}

	doc := soup.HTMLParse(resp)

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
		print(fmt.Sprintf("%v %s", err, "<- Error"))
	}

	if len(elements) == 12 {
		pg = pg + 1
	}

	ar := models.SearchAnimeResponse{
		Data:    animes,
		Status:  "200",
		Message: "Success",
		Page:    pg,
	}

	return ar, err
}

func (a *Server) IndexHandler() http.HandlerFunc {
	bearer = Config().Bearer
	fmt.Println(bearer)

	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "Gosanime Api Running")
	}
}

func (a *Server) GetTopAnimes() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		if r.Header.Get("Authorization") != fmt.Sprintf("Bearer %v", bearer) {
			sendResponse(w, r, getNoBearer, http.StatusUnauthorized)
		} else {
			animes, err := lastAnimes()

			if err != nil {
				log.Printf("cant get latest animes err=%v \n", err)
				sendResponse(w, r, nil, http.StatusInternalServerError)
				return
			}

			var resp = make([]models.Anime, len(animes))
			for ifx, anime := range animes {
				resp[ifx] = mapToJson(&anime)
			}

			sendResponse(w, r, makeArrayResponse(resp), http.StatusOK)
		}
	}
}

func (a *Server) GetOvas() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != fmt.Sprintf("Bearer %v", bearer) {
			sendResponse(w, r, getNoBearer, http.StatusUnauthorized)
		} else {
			animes, err := ovas()
			if err != nil {
				log.Printf("cant get latest animes err=%v \n", err)
				sendResponse(w, r, nil, http.StatusInternalServerError)
				return
			}

			var resp = make([]models.Anime, len(animes))
			for ifx, anime := range animes {
				resp[ifx] = mapToJson(&anime)
			}

			sendResponse(w, r, makeArrayResponse(resp), http.StatusOK)
		}
	}
}

func (a *Server) GetAnime() http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != fmt.Sprintf("Bearer %v", bearer) {
			sendResponse(w, r, getNoBearer, http.StatusUnauthorized)
		} else {
			anime, err := anime(r)
			if err != nil {
				log.Printf("cant get latest animes err=%v \n", err)
				sendResponse(w, r, nil, http.StatusInternalServerError)
				return
			}

			var response = models.Response{
				Data:    anime,
				Status:  "200",
				Message: "Success",
			}

			sendResponse(w, r, response, http.StatusOK)
		}
	}
}

func (a *Server) GetVideoServers() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != fmt.Sprintf("Bearer %v", bearer) {
			sendResponse(w, r, getNoBearer, http.StatusUnauthorized)
		} else {
			episodes, err := videosByServer(r)

			if err != nil {
				log.Printf("cant get latest animes err=%v \n", err)
				sendResponse(w, r, nil, http.StatusInternalServerError)
				return
			}

			var response = models.Response{
				Data:    episodes,
				Status:  "200",
				Message: "Success",
			}

			sendResponse(w, r, response, http.StatusOK)

		}
	}
}

func (a *Server) GetSearchAnime() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != fmt.Sprintf("Bearer %v", bearer) {
			sendResponse(w, r, getNoBearer, http.StatusUnauthorized)
		} else {
			animes, err := searchAnime(r)

			errorResponse := models.SearchAnimeResponse{
				Data:    nil,
				Status:  "401",
				Message: "Not Found",
				Page:    -1,
			}

			if err != nil {
				sendResponse(w, r, errorResponse, http.StatusNotFound)
			} else {
				sendResponse(w, r, animes, http.StatusOK)
			}
		}
	}
}

func getHiddenUrl(url string) string {
	resp, err := http.Get(url)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	return fmt.Sprint(resp.Request.URL)
}

type Url struct {
	File string
}

type Bearer struct {
	Message string `json:"message"`
	Status  string `json:"status"`
	Code    int    `json:"code"`
}

func makeArrayResponse(animes []models.Anime) interface{} {
	return models.ArrayResponse{
		Data:    animes,
		Status:  "200",
		Message: "Success",
	}
}

func getNoBearer() interface{} {
	return Bearer{
		Message: "Bearer token not present",
		Status:  "Unauthorized",
		Code:    401,
	}
}
