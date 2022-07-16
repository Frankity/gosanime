# Gosanime

Gosanime is a kind of web scrapper written in [Go](https://go.dev/), the purpose of it is read the content of [Jkanime](https://jkanime.net/) site and serve as an API.

**Available endpoints:**

|  |  |
|--|--|
| root | {host}/ |
| main | {host}/api/v1/main |
| ovas | {host}/api/v1/ovas |
| search | {host}/api/v1/search?anime=naruto&page=1n |
| anime | {host}/api/v1/anime?id=naruto |
| episode | {host}/api/v1/video?anime=spy-x-family/&episode=1 |
| tags | {host}/api/v1/tags?tag=shounen&page=1 |
 

Insomnia collection [here](Insomnia_2022-07-15.json)

wanna try it?

    git clone https://github.com/Frankity/gosanime.git
    cd gosanime
    go get ./
    go run main.go

and heck your *localhost* in the port *3000*.

---
**To Do:**

- Decode correctly another video server due the only one i played with is the jk one.
- Fix the pagination logic.
- Paginate tags correctly.
- Increase the search response range.
- Add more endpoints for other pages of the site.

willing to help?, fork and send some pr's.
