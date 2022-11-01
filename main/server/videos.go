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
	"xyz.frankity/gosanime/main/config"
	"xyz.frankity/gosanime/main/models"
)

func videosByServer(r *http.Request) (interface{}, error) {
	if err := r.ParseForm(); err != nil {
		os.Exit(1)
	}
	anime := r.Form.Get("anime")
	episode := r.Form.Get("episode")

	client := http.DefaultClient

	http.DefaultClient.Transport = config.AddCloudFlareByPass(http.DefaultClient.Transport)

	res, err := client.Get(fmt.Sprintf("%v/%s/%s/", config.Rooturl, anime, episode))
	if err != nil {
		log.Fatal(err)
	}

	responseData, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Fatal(err)
	}

	responseString := string(responseData)

	doc := soup.HTMLParse(responseString)

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
						ur := strings.Replace(urli, "/jk.php?u=", "https://jkanime.net/", -1)
						urls = append(urls, GetHiddenUrl(ur))
						break
					}

					doc, err := soup.Get(urli)

					if err != nil {
						fmt.Print(err)
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
