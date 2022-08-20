package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"strings"
)

// gets the results, processes and returns
type FixerFunction func(string) string

type GenericStep struct {
	regex                regexp.Regexp
	fct                  FixerFunction
	additionalPageBefore *GenericStep
}

type GenericFinding struct {
	regex regexp.Regexp
	fct   FixerFunction
}

const (
	crawlOnCategory    int = 0
	crawlOnSubcategory     = 1
	crawlOnProduct         = 2
)

type CrawlMetadata struct {
	crawlType            int
	categoryStep         GenericStep
	categoryLastPage     int
	subcategoryStep      GenericStep
	subcategoryLastPage  int
	productStep          GenericStep
	productName          GenericFinding
	productDownloadCount GenericFinding
	productSize          GenericFinding
	productLanguage      GenericFinding
}

func RetrieveWebsite(website string) string {
	resp, err := http.Get(website)

	if err != nil {
		log.Fatal(err)
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)

	if err != nil {
		log.Fatal(err)
	}
	return string(body)
}

func CrawlProduct(rootWebsite string, metadata CrawlMetadata) {
	rootWebsiteContent := RetrieveWebsite(rootWebsite)
	productRegex := metadata.productStep.regex
	allMatches := productRegex.FindAllStringSubmatch(rootWebsiteContent, -1)
	for _, element := range allMatches {
		var productDownloadCount string
		productWebsite := string(element[1])
		productContent := RetrieveWebsite(productWebsite)
		productName := metadata.productName.regex.FindStringSubmatch(productContent)[1]

		productDownloadCount = metadata.productDownloadCount.regex.FindStringSubmatch(productContent)[1]
		productDownloadCountFct := metadata.productDownloadCount.fct

		if productDownloadCountFct != nil {
			productDownloadCount = productDownloadCountFct(productDownloadCount)
		}

		fmt.Println(productWebsite)
		fmt.Println(productName)
		fmt.Println(productDownloadCount)

		break
	}
}

func Crawl(rootWebsite string, metadata CrawlMetadata) {
	if metadata.crawlType == crawlOnProduct {
		CrawlProduct(rootWebsite, metadata)
	}
}

func main() {
	Crawl("https://oldergeeks.com/sitemap.xml",
		CrawlMetadata{crawlOnProduct,
			GenericStep{},
			0,
			GenericStep{},
			0,
			GenericStep{
				*regexp.MustCompile(`>([^<]+file\.php[^<]+)`),
				nil,
				nil},
			GenericFinding{*regexp.MustCompile(`File\s-\sDownload\s([^<]+)`), nil},
			GenericFinding{*regexp.MustCompile(`<tr>\s+.+>([\d,]+)</td>`), func(s string) string { return strings.Replace(s, ",", "", 1) }},
			GenericFinding{*regexp.MustCompile(`row\">([\d.]+[GMKB]+)`), nil},
			GenericFinding{}})
}
