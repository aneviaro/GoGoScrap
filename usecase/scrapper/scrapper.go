package scrapper

import (
	"errors"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"net/http"
	"strings"
)

var googleDomains = map[string]string{
	"com": "https://www.google.com/search?q=",
	"uk":  "https://www.google.co.uk/search?q=",
	"ru":  "https://www.google.ru/search?q=",
	"fr":  "https://www.google.fr/search?q=",
	"by":  "https://www.google.com/search?q=",
	"":    "https://www.google.com/search?q=",
}

type URLRank struct {
	Rank int
	URL  string
}

func buildGoogleQuery(searchTerm string, countryCode string, languageCode string) string {
	searchTerm = strings.Trim(searchTerm, " ")
	searchTerm = strings.Replace(searchTerm, " ", "+", -1)
	if googleBase, found := googleDomains[countryCode]; found {
		return fmt.Sprintf("%s%s&num=100&hl=%s", googleBase, searchTerm, languageCode)
	} else {
		return fmt.Sprintf("%s%s&num=100&hl=%s", googleDomains["com"], searchTerm, languageCode)
	}
}

func performGoogleRequest(searchURL string) (*http.Response, error) {
	baseClient := &http.Client{}

	req, _ := http.NewRequest("GET", searchURL, nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/61.0.3163.100 Safari/537.36")

	res, err := baseClient.Do(req)

	return res, err
}

func parseGoogleResponse(response *http.Response, siteName string) (*URLRank, error) {
	doc, err := goquery.NewDocumentFromReader(response.Body)
	if err != nil {
		return &URLRank{}, err
	}
	sel := doc.Find("div.g")
	rank := 1
	for i := range sel.Nodes {
		item := sel.Eq(i)
		linkTag := item.Find("a")
		link, _ := linkTag.Attr("href")
		link = strings.Trim(link, " ")
		if link != "" && link != "#" {
			result := &URLRank{
				rank,
				link,
			}
			if strings.Contains(link, siteName) {
				return result, nil
			}
			rank += 1
		}
	}
	return &URLRank{}, err
}

func GetWebsitePositionForQuery(query string, langCode string, website string) (*URLRank, error) {
	googleQueryURL := buildGoogleQuery(query, "by", langCode)
	res, err := performGoogleRequest(googleQueryURL)
	if err != nil {
		return nil, err
	}
	if res.StatusCode != 200 {
		return nil, errors.New("not expected google response")
	}

	return parseGoogleResponse(res, website)
}
