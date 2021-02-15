package main

import (
	"os/signal"
	"syscall"
	"fmt"
	"log"
	"os"
	"time"
	"strings"
	"net"
	"net/http"
	"net/rpc"
	"math/rand"
	"io/ioutil"
	"scraping"
)

const MAX_URLS_PER_DOMAIN = 4

// Configuration of the coordinator node
type Config struct {
	batchSize int // Size of the batches sent to the nodes
	myAddress string // The IP and port on which the coordinator is listening
	myPort string // Only the port
}

type State struct {
	config Config
	nodes []scraping.Node
	queueChan chan scraping.Request
	shutdownChan chan bool
	nodeReadyChan chan scraping.Node
	startTime time.Time
	lastReadyTime time.Time
	totalURLsToRequest int
	totalScraped int
	totalTimeouts int
	totalDNSErrors int
	totalFailures int
	totalScripts int
	batchesDispatched int
	resultsReceived int
}
var state State

// The RPC server for the main server
type Server int

func (t *Server) Results(args *scraping.BatchResult, reply *bool) error {
	if (*args).NotQueried != nil {
		log.Printf("Received partial results from %s", (*args).Node.URL)
	} else {
		log.Printf("Received results from %s", (*args).Node.URL)
	}
	for _, req := range (*args).NotQueried {
		state.queueChan <- req // Reschedule failed requests
	}
	if len((*args).Results) > 0 {
		for _, result := range (*args).Results {
			state.resultsReceived += 1
			StoreResult(result)
			for _, url := range result.URLs {
				state.queueChan <- scraping.Request{url, false} // Not a toplevel url
			}
			state.totalURLsToRequest += len(result.URLs)
		}
	}
	*reply = true
	return nil
}

func (t *Server) NodeReady(args *scraping.Node, reply *bool) error {
	log.Printf("Node is ready: %s", (*args).URL)
	MarkReady(*args)
	nodeAlreadySeen := false
	for _, node := range state.nodes {
		if node == *args {
			nodeAlreadySeen = true
		}
	}
	if !nodeAlreadySeen {
		state.nodes = append(state.nodes, *args)
	}
	*reply = true
	return nil
}

func (t *Server) Shutdown(args *bool, reply *bool) error {
	state.shutdownChan <- true
	return nil
}

func SetupSIGTERMHandler() {
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		fmt.Println("Shutting down upon request from the terminal...")
		EndScraping()
		state.shutdownChan <- true
	}()
}

func StartServer() {
	server := new(Server)
	rpc.Register(server)
	rpc.HandleHTTP()
	l, err := net.Listen("tcp", ":" + state.config.myPort)
	if err != nil {
		log.Fatalf("Could not listen on port %s: %v", state.config.myPort, err)
	}
	log.Printf("Listening on port %s\n", state.config.myPort)
	go http.Serve(l, nil)
}

func ExtractPort(url string) string {
	return strings.Split(url, ":")[1]
}

func main() {
	rand.Seed(time.Now().UnixNano())
	if len(os.Args) != 2 {
		log.Fatalf("Expected 1 argument, got %d", len(os.Args)-1)
	}
	urls := LoadURLs("urls.txt")
	queueBufferSize := MAX_URLS_PER_DOMAIN * len(urls)

	state.config.myAddress = os.Args[1]
	state.config.myPort = ExtractPort(os.Args[1])
	state.config.batchSize = 100
	state.nodes = make([]scraping.Node, 0)
	state.queueChan = make(chan scraping.Request, queueBufferSize)
	state.shutdownChan = make(chan bool, 0)
	state.nodeReadyChan = make(chan scraping.Node, 100)
	state.startTime = time.Now()
	state.lastReadyTime = state.startTime
	// All other counters are initialized to 0 by default

	StartServer()
	go ServeBatches()
	go FrequentlyPrintStats()
	Initialize(urls)
	SetupSIGTERMHandler()
	<- state.shutdownChan
}

// Returns the URLs to scrape
func LoadURLs(urlsFile string) []string {
	// Fetch URLs from the given file, which should be formatted as one URL per line
	// NOTE: This reads the entire file and stores it in memory.
	//       This can be memory-intensive if the file is large, but
	//       majestic_1million.csv is 78MB, so that should be fine
	content, err := ioutil.ReadFile(urlsFile)
	if err != nil {
		log.Fatalf("Could not read URLs file: %v", err)
	}
	lines := strings.Split(string(content), "\n")
	return lines
}

// Initialize the scraping, queuing all top level urls to scrape
func Initialize(lines []string) {
	// Put all URLs in the queue
	for _, line := range lines {
		if line != "" { // Filter empty lines, in case there are any
			state.queueChan <- scraping.Request{line, true} // This is a top level request
			state.totalURLsToRequest += 1
		}
	}
}

// Mark a node as ready
func MarkReady(node scraping.Node) {
	log.Printf("Node ready: %s ", node.URL)
	state.lastReadyTime = time.Now()
	state.nodeReadyChan <- node
}

// Terminate scraping
func EndScraping() {
	log.Println("Terminating scraping nodes")
	// Notify all nodes to terminate
	for _, node := range state.nodes {
		client, err := rpc.DialHTTP("tcp", node.URL)
		if err != nil {
			log.Fatalf("Could not notify node %s of shutdown: %v", node.URL, err)
		}
		var reply bool
		err = client.Call("NodeServer.Shutdown", true, &reply)
		if err != nil {
			log.Fatalf("Could not notify node %s of shutdown: %v", node.URL, err)
		}
	}
	// Terminate main server
	state.shutdownChan <- true
}

// Dispatch a batch of request to a node that is ready, wait for one if needed
func DispatchBatch(requests []scraping.Request) {
	for {
		log.Println("Waiting for a node to be ready")
		node := <- state.nodeReadyChan
		state.batchesDispatched += 1
		log.Printf("Dispatching batch %d/%d to %s", state.batchesDispatched, state.totalURLsToRequest / state.config.batchSize, node.URL)
		var reply bool
		client, err := rpc.DialHTTP("tcp", node.URL)
		if err != nil {
			log.Fatalf("Could not send batch to node %s: %v", node.URL, err)
		}
		err = client.Call("NodeServer.Batch", scraping.Batch{requests}, &reply)
		if err != nil {
			log.Fatalf("Could not send batch to node %s: %v", node.URL, err)
		}
		return
	}
}

// Serve the batches by looking for requests on queueChan
func ServeBatches() {
	for {
		batch := make([]scraping.Request, 0, state.config.batchSize)
		// Get enough URLs
		for i := 0; i < state.config.batchSize; {
			select {
			case request := <- state.queueChan:
				batch = append(batch, request)
				i++
			case <-time.After(10 * time.Second):
				// No more responses to expect?
				if state.batchesDispatched == state.resultsReceived {
					// Yes, stop looking for more URLs for this batch
					i = state.config.batchSize // no way to break out of the for loop without this
				} // Otherwise, really wait for next URL
			}
		}
		if len(batch) == 0 {
			// No URLs to dispatch, check if all nodes are finished
			if state.batchesDispatched == state.resultsReceived {
				// If so, scraping is done
				EndScraping()
				return;
			} // Otherwise, do nothing, there might be later batches coming
		} else {
			// Dispatch the batch to one of the nodes that are ready
			DispatchBatch(batch)
		}
	}
}

// Frequently print the stats of the scraping process
func FrequentlyPrintStats() {
	for {
		<-time.After(1 * time.Minute) // Every minute
		PrintStats()
	}
}
// Print the stats of the scraping process
func PrintStats() {
	var t = time.Now().Sub(state.startTime)
	var rate float64
	if t.Seconds() == 0 {
		rate = 0
	} else {
		rate = float64(state.totalScraped) / t.Seconds()
	}
	log.Printf("Scraped %d URLs (on %d to scrape so far) in %s [%v URL/s]:", state.totalScraped, state.totalURLsToRequest, t.String(), rate)
	log.Printf("\t%d scripts found, %d failures, %d DNS errors, %d timeouts", state.totalScripts, state.totalFailures, state.totalDNSErrors, state.totalTimeouts)
	log.Printf("\tLast node ready was %s ago", time.Now().Sub(state.lastReadyTime).String())
	if rate != 0 {
		log.Printf("\tRemaining time: between %s and %s",
			(time.Duration(float64(state.totalURLsToRequest - state.totalScraped) / rate) * time.Second).String(),
			(time.Duration(float64((state.totalURLsToRequest * 4) - state.totalScraped) / rate) * time.Second).String())
	}
}

// Add a line to a given file
func AddLine(file string, line string) {
	f, err := os.OpenFile(file, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("Cannot write to file %s: %v", file, err)
	}
	defer f.Close()
	if _, err := f.WriteString(line + "\n"); err != nil {
		log.Fatalf("Cannot write to file %s: %v", file, err)
	}
}

// Store the result of a query
func StoreResult(result scraping.Result) {
	state.totalScraped += 1
	if result.Timeout {
		AddLine("/tmp/out/timeouts.log", result.URL)
		state.totalTimeouts += 1
	} else if result.DNSError {
		AddLine("/tmp/out/dnserrors.log", result.URL)
		state.totalDNSErrors += 1
	} else if result.Failure {
		AddLine("/tmp/out/failures.log", result.URL)
		state.totalFailures += 1
	} else if (len(result.Scripts) == 0) {
		AddLine("/tmp/out/noscripts.log", result.URL)
	} else {
		log.Printf("Found a script! On page %s, scripts are %v", result.URL, result.Scripts)
		AddLine("/tmp/out/scripts.log", fmt.Sprintf("%s %v", result.URL, result.Scripts))
		state.totalScripts += 1
	}
}
