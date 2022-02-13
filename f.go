package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/anaskhan96/soup"
)

func maain() {
	name := "https://jkanime.net/chobits/"

	resp, err := soup.Get(name)
	if err != nil {
		os.Exit(1)
	}

	doc := soup.HTMLParse(resp)
	genres := doc.Find("div", "class", "anime__details__widget").Find("ul").FindAll("li")[1].FindAll("a")

	result := []string{}

	for a := range genres {
		result = append(result, genres[a].Text())
	}

	strings.Join(result, ",")
	// animes, nil

	fmt.Println(result)

}
