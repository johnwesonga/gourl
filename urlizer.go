package main

import (
    "bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"
)

const apiurl = "https://www.googleapis.com/urlshortener/v1/url"

var (
	domain_file = flag.String("domain_file", "domains.txt", "file containing all the domains")
)

type UrlMap struct {
	Url          string
	ShortenedUrl string
}

type ShortenedUrl struct {
	Kind    string
	Id      string
	Longurl string
}

func retrieveUrls() []string {
	domain_file, err := ioutil.ReadFile(*domain_file)
	if err != nil {
		log.Fatal(" error ", err)
	}
	return strings.Split(string(domain_file), "\n")
}

func shorten(url string) (id string, err error) {
	buf := bytes.NewBuffer(nil)
	buf.WriteString(fmt.Sprintf(`{"longUrl": "%s"}`, url))
	res, err := http.Post(apiurl, "application/json", buf)
	if err != nil {
		log.Fatal("error", err)
		return "", err
	}
	body, _ := ioutil.ReadAll(res.Body)
	var shortened ShortenedUrl
	json.Unmarshal(body, &shortened)
	return shortened.Id, nil
}

func shortenUrl(returnChannel chan string, url string) (shortenedUrls []string, err error) {
	buf := bytes.NewBuffer(nil)
	buf.WriteString(fmt.Sprintf(`{"longUrl": "%s"}`, url))
	res, err := http.Post(apiurl, "application/json", buf)
	if err != nil {
		log.Fatal("error", err)
		return nil, err
	}
	body, _ := ioutil.ReadAll(res.Body)
	var shortened ShortenedUrl
	json.Unmarshal(body, &shortened)
	shortenedUrls = append(shortenedUrls, shortened.Id)
	returnChannel <- shortened.Id
	return shortenedUrls, nil
}

func waitForDomains(responseChannel chan string, numberOfDomains int) (domainMapping []string) {
	returnedCount := 0
	for {
		domainMapping = append(domainMapping, <-responseChannel)
		returnedCount++

		if returnedCount >= numberOfDomains {
			break
		}
	}

	return
}

func main() {
	start := time.Now()
	flag.Parse()
	//create a channel where the shortened url will be sent to
	responseChannel := make(chan string)
	if len(*domain_file) == 0 {
		log.Fatal("Please specify the path to the domain file")
	}

	longurls := retrieveUrls()
	for _, url := range longurls {
		go shortenUrl(responseChannel, url)
	}

	domainMapping := waitForDomains(responseChannel, len(longurls))

	fmt.Println(domainMapping)
	elapsed := time.Since(start)
	fmt.Println(elapsed)

}
