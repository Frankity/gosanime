package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/anaskhan96/soup"
)

func maicn() {
	name := "https://jkanime.net/top/"

	resp, err := soup.Get(name)
	if err != nil {
		os.Exit(1)
	}

	doc := soup.HTMLParse(resp)
	main := doc.Find("section", "class", "contenido").Find("div", "class", "container").Find("div", "class", "row").Find("div", "class", "col-lg-12").FindAll("div", "class", "list")
	for _, p := range main {
		parent := p.Find("div", "id", "conb")
		fmt.Println(strings.Split(parent.Find("a").Attrs()["href"], "/")[3])
		fmt.Println(parent.Find("a").Attrs()["title"])
		fmt.Println(p.Find("a").Find("img").Attrs()["src"])
		fmt.Println(strings.TrimSpace(p.Find("div", "id", "animinfo").Find("h2", "class", "portada-title").Text()))
		fmt.Println(strings.TrimSpace(strings.Split(p.Find("div", "id", "animinfo").Find("span", "class", "title").Text(), "/")[0]))

	}
	// animes, nil

}
