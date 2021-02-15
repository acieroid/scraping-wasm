package main

import (
	"bufio"
	"net"
	"net/url"
	"os"
	"log"
	"fmt"
	"time"
	"sort"
)

// Load all files of a file
func LinesInFile(file string) []string {
	f, err := os.Open(file)
	if err != nil {
		log.Fatalf("Can't open file: %v", err)
	}
	scanner := bufio.NewScanner(f)
	result := []string{}
	for scanner.Scan() {
		line := scanner.Text()
		result = append(result, line)
	}
	return result
}

var domains []string
var scraped []string
// Load scraped URLS into the `scraped` variable and top-level domains that needed to be scraped in `domains`
func Init() {
	scraped = append(LinesInFile("scripts.log"), LinesInFile("noscripts.log")...)
	sort.Strings(scraped)
	domains = LinesInFile("urls.txt")
	sort.Strings(domains)
}

// Check if an URL has been scraped
func IsSuccessfullyScraped(url string) bool {
	idx := sort.SearchStrings(scraped, url)
	return idx < len(scraped) && scraped[idx] == url
}

// Check if an URL is a top-level domain to scrape
func IsTopLevel(url string) bool {
	idx := sort.SearchStrings(domains, url)
	return idx < len(domains) && scraped[idx] == url
}


// Add a domain to a file
func AddDomainToFile(domain string, file string) {
	f, err := os.OpenFile(file, os.O_APPEND | os.O_CREATE | os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("Can't open urls.txt file to write: %v", err)
	}
	defer f.Close()
	if _, err := f.WriteString(fmt.Sprintf("%s\n", domain)); err != nil {
		log.Fatal("Can't add domain to urls.txt: %v", err)
	}
}

// Check if a DNS is valid by performing a DNS lookup
func ValidDNS(link string) bool {
	u, err := url.Parse(link)
	if err != nil {
		log.Fatal("Malformed URL: %v", err)
	}
	ips, err := net.LookupIP(u.Host)
	if err != nil {
		// Could not find an IP
		return false
	}
	if len(ips) == 0 {
		return false
	}
	return true
}

func FilterValidDNS() {
	urls := LinesInFile("dnserrors.log")
	concurrency := 0
	validChan := make(chan string, 0)
	invalidChan := make(chan string, 0)
	go func() {
		for {
			AddDomainToFile(<- validChan, "dnserrors-valid.log")
		}
	}()
	go func() {
		for {
			AddDomainToFile(<- invalidChan, "dnserrors-invalid.log")
		}
	}()
	for i, url := range(urls) {
		if !IsSuccessfullyScraped(url) {
			if true {
				fmt.Printf("\r%d/%d", i, len(urls))
			}
			scheduled := false
			for !scheduled {
				<-time.After(100 * time.Millisecond)
				if concurrency < 10 {
					concurrency = concurrency + 1
					scheduled = true
					go func(url string) {
						if ValidDNS(url) {
							validChan <- url
						} else {
							invalidChan <- url
						}
						concurrency = concurrency - 1
					}(url)
				}
			}
		} else {
			fmt.Printf("Ignoring already scraped url: %s\n", url)
		}
	}
}

// Read urls which resulted in a DNS error from dnserrors.log
// First check if the url has actually been scraped in another scraping run, by going through the scripts.log and noscripts.log files
// Output the URLs for which the DNS is actually valid in dnserrors-valid.log
// Output the other ones in dnserrors-invalid.log
func main() {
	Init()
	FilterValidDNS()
}
