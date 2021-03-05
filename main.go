package main

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"log"
	"net/http"
	"os"
	"strings"
)

func main() {
	http.HandleFunc("/", checkAuth(http.HandlerFunc(scrapGoogle)))
	port := os.Getenv("PORT")

	if port == "" {
		port = "8080"
		log.Printf("Defaulting to port %s", port)
	}

	log.Printf("Listening on port %s", port)

	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}

}

func checkAuth(next http.HandlerFunc) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
	}
}

var googleDomains = map[string]string{
	"com": "https://www.google.com/search?q=",
	"uk":  "https://www.google.co.uk/search?q=",
	"ru":  "https://www.google.ru/search?q=",
	"fr":  "https://www.google.fr/search?q=",
	"":    "https://www.google.com/search?q=",
}

type GoogleResult struct {
	ResultRank  int
	ResultURL   string
}

func buildGoogleUrl(searchTerm string, countryCode string, languageCode string) string {
	searchTerm = strings.Trim(searchTerm, " ")
	searchTerm = strings.Replace(searchTerm, " ", "+", -1)
	if googleBase, found := googleDomains[countryCode]; found {
		return fmt.Sprintf("%s%s&num=100&hl=%s", googleBase, searchTerm, languageCode)
	} else {
		return fmt.Sprintf("%s%s&num=100&hl=%s", googleDomains["com"], searchTerm, languageCode)
	}
}

func googleRequest(searchURL string) (*http.Response, error) {

	baseClient := &http.Client{}

	req, _ := http.NewRequest("GET", searchURL, nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/61.0.3163.100 Safari/537.36")

	res, err := baseClient.Do(req)

	if err != nil {
		return nil, err
	} else {
		return res, nil
	}
}

func googleResultParser(response *http.Response, siteName string) (GoogleResult, error) {
	doc, err := goquery.NewDocumentFromReader(response.Body)
	if err != nil {
		return GoogleResult{}, err
	}
	sel := doc.Find("div.g")
	rank := 1
	for i := range sel.Nodes {
		item := sel.Eq(i)
		linkTag := item.Find("a")
		link, _ := linkTag.Attr("href")
		link = strings.Trim(link, " ")
		if link != "" && link != "#" {
			result := GoogleResult{
				rank,
				link,
			}
			if strings.Contains(link, siteName) {
				return result, nil
			}
			rank += 1
		}
	}
	return GoogleResult{}, err
}

func scrapGoogle(writer http.ResponseWriter, request *http.Request) {
	urlQuery := request.URL.Query()

	query, ok := urlQuery["query"]
	if !ok || len(query[0]) < 1 {
		log.Println("Unable to parse query URL param.")
		return
	}

	languageParam, ok := urlQuery["language"]
	var langCode string
	if !ok || len(languageParam[0]) < 1 {
		log.Println("Unable to parse language URL param.")
		langCode=""
	} else {
		langCode = languageParam[0]
	}


	countryParam, ok := urlQuery["country"]
	var countryCode string
	if !ok || len(countryParam[0]) < 1 {
		log.Println("Unable to parse country URL param.")
		countryCode = ""
	} else {
		countryCode = countryParam[0]
	}

	site, ok := urlQuery["site"]
	if !ok || len(site[0]) < 1 {
		log.Println("Unable to parse site URL param.")
		return
	}

	googleUrl := buildGoogleUrl(query[0], countryCode, langCode)
	res, err := googleRequest(googleUrl)
	if err != nil {
		log.Printf("Unable to make google request, err: %v", err)
	}
	scrape, err := googleResultParser(res, site[0])
	if err != nil {
		log.Printf("Unable to parse google response, err: %v", err)
	} else {
		fmt.Fprintf(writer, "")
	}

}
