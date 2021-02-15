package main

import (
	"bufio"
	"os"
	"log"
	"sort"
	"fmt"
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
	domains = LinesInFile("../urls.txt")
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

// Go over top-level domains to scrape and reschedule the ones that were not successfully scraped, to urls-toplevel.txt
func RescheduleTopLevelDomains() {
	for i, domain := range domains {
		if i % 1000 == 0 {
			// Display progress
			fmt.Printf("\r%d", i)
		}
		if !IsSuccessfullyScraped(domain) {
			AddDomainToFile(domain, "urls-toplevel.txt")
		}
	}
	fmt.Printf("\rTop levels: done\n")
}

// Reschedule failures that are not top-level domains from file into urls-nottoplevel.txt
func RescheduleFailures(file string) {
	failures := LinesInFile(file)
	for i, url := range failures {
		fmt.Printf("\r%s %d/%d", file, i, len(failures))
		if !IsSuccessfullyScraped(url) && !IsTopLevel(url) {
			AddDomainToFile(url, "urls-nottoplevel.txt")
		}
	}
}

func main() {
	Init()
	RescheduleTopLevelDomains()
	RescheduleFailures("failures.log")
	RescheduleFailures("timeouts.log")
	RescheduleFailures("dnserrors-valid.log")
}
