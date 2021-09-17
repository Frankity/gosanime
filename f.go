package main

import (
	"fmt"
	"net/http"
)

func masin() {
	name := "https://jkanime.net/stream/jkmedia/1ad82f40afcedec7a5d6c873986e7c15/a28e5f284a491ba9f012bd30c66f58ee/1/a3c36c9b6023d7318403b2f6da7c013e/"

	resp, err := http.Get(name)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	fmt.Println(resp.Request.URL)

}
