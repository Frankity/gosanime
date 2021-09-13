package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

func maisn() {

	name := "https://jkanime.net/jk.php?u=stream/jkmedia/0721571b35ca6f70e30d55e7b6bd38fd/a28e5f284a491ba9f012bd30c66f58ee/1/b99019d0f5f293ea80c39be58ddce738"

	if strings.Contains(name, "jk.php") {

		fmt.Printf("HTML code of %s ...\n", name)
		resp, err := http.Get(name)
		// handle the error if there is one
		if err != nil {
			panic(err)
		}
		// do this now so it won't be forgotten
		defer resp.Body.Close()
		// reads html as a slice of bytes
		html, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			panic(err)
		}
		// show the HTML code as a string %s
		fmt.Printf("%s\n", html)

	}
	/*f, err := soup.Get(a)

	i := soup.HTMLParse(f)

	print(i.HTML())*/

	//log.Println(string(body))

}
