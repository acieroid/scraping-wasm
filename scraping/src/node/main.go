package main

import (
	"fmt"
	"os/signal"
	"syscall"
	"os"
	"log"
	"context"
	"time"
	"net"
	"net/http"
	"net/url"
	"net/rpc"
	"math/rand"
	"strings"
	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/cdproto/debugger"
	"github.com/chromedp/chromedp"
	tld "github.com/jpillora/go-tld"
	"scraping"
)

const (
	NWORKERS = 4
	URLS_TO_EXTRACT = 3
	PREVENT_HEADLESS_DETECTION = false
	USE_TOR = false
	TIMEOUT_SECONDS = 35
)

type Config struct {
	serverAddress string
	myself scraping.Node
	port string
}

type State struct {
	config Config
	batchChan chan []scraping.Request
	shutdownChan chan bool
	gracefulShutdownChan chan bool
	chromeContext context.Context
}
var state State

// The RPC server for a node
type NodeServer int

// Receive a new batch of requests
func (t *NodeServer) Batch(args *scraping.Batch, reply *bool) error {
	state.batchChan <- (*args).Requests
	*reply = true
	return nil
}

// Receive the shutdown signal
func (t *NodeServer) Shutdown(args *bool, reply *bool) error {
	log.Println("Shutting down upon request from server")
	state.shutdownChan <- true
	*reply = true
	return nil
}

func SetupSIGTERMHandler() {
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		fmt.Println("Shutting down upon request from the terminal...")
		state.gracefulShutdownChan <- true
	}()
}

func StartServer() {
	// Register the RPC endpoint
	nodeServer := new(NodeServer)
	rpc.Register(nodeServer)
	rpc.HandleHTTP()
	l, err := net.Listen("tcp", ":" + state.config.port)
	if err != nil {
		log.Fatalf("Cannot listen on port %s: %v", state.config.port, err)
	}
	log.Printf("Listening on port %s\n", state.config.port)
	go http.Serve(l, nil)
}

func ExtractPort(url string) string {
	return strings.Split(url, ":")[1]
}

func main() {
	rand.Seed(time.Now().UnixNano())
	if len(os.Args) != 3 {
		log.Fatalf("Expected 2 argument, got %d", len(os.Args)-1)
	}
	state.config.serverAddress = os.Args[1]
	state.config.myself = scraping.Node{os.Args[2]}
	state.config.port = ExtractPort(os.Args[2])
	// Allocate channels
	state.batchChan = make(chan []scraping.Request, 0)
	state.shutdownChan = make(chan bool, 0)
	state.gracefulShutdownChan = make(chan bool, 0)
	SpawnChrome()
	StartServer()
	go HandleBatches()
	SetupSIGTERMHandler()
	NotifyImReady()
	// Wait until termination
	<- state.shutdownChan
}

// Notify that we're ready to receive requests
func NotifyImReady() {
	var reply bool
	client, err := rpc.DialHTTP("tcp", state.config.serverAddress)
	if err != nil {
		log.Fatalf("Cannot connect to server: %v", err)
	}
	err = client.Call("Server.NodeReady", state.config.myself, &reply)
	if err != nil {
		log.Fatalf("Cannot connect to server: %v", err)
	}
}
// Handle batches of requests
func HandleBatches() {
	for {
		// Take one batch
		select {
		case batch := <- state.batchChan:
			// Perform the requests and send result to server
			log.Printf("Received a new batch of %v requests\n", len(batch))
			start := time.Now()
			results := PerformRequests(batch)
			end := time.Now()
			elapsed := end.Sub(start)
			log.Printf("Performed all requests in %v\n", elapsed.String())
			SendResultsToServer(results)
			if len(results.NotQueried) > 0 {
				state.shutdownChan <- true // graceful shutdown has been requested
			} else {
				NotifyImReady()
			}
		case <- state.gracefulShutdownChan:
			state.shutdownChan <- true
		}
	}
}

// Send a batch result to the server
func SendResultsToServer(results scraping.BatchResult) {
	var reply bool
	// Send results
	client, err := rpc.DialHTTP("tcp", state.config.serverAddress)
	if err != nil {
		log.Fatalf("Cannot connect to server: %v", err)
	}
	err = client.Call("Server.Results", results, &reply)
	if err != nil {
		log.Fatalf("Cannot connect to server: %v", err)
	}
	log.Printf("Sent results to server")
}


// Wait a random amount of time between two subsequent requests
func WaitBetweenRequests(min int, max int) <- chan time.Time {
	// Wait between 0 and max ms
	return time.After(time.Duration(min + (rand.Int() % (max - min))) * time.Millisecond)
}

// Performs all the requests from a given queue of requests
func PerformRequests(queue []scraping.Request) scraping.BatchResult {
	results := make([]scraping.Result, 0, len(queue))
	requestChan := make(chan scraping.Request, 0)
	finished := make(chan bool)
	for i := 0; i < NWORKERS; i++ {
		// Launch the worker
		workerId := i
		go func() {
			for {
				log.Printf("[worker-%d] Waiting for a request", workerId)
				request, morerequests := <-requestChan // Read a request from the channel
				if !morerequests {
					log.Printf("[worker-%d] No more requests", workerId)
					finished <- true
					return; // Channel has been closed, stop the worker
				} else {
					// Perform the request and store the result
					result, err := ExtractScripts(workerId, request)
					if err != nil {
						state.gracefulShutdownChan <- true
					} else {
						results = append(results, result)
					}
				}
			}
		}()
	}
	// Schedule all requests
	log.Printf("Scheduling %v requests\n", len(queue))
	for i, request := range queue {

		select {
		case <-state.gracefulShutdownChan:
			// Send partial results
			log.Printf("Asking for graceful shutdown")
			return scraping.BatchResult{results, state.config.myself, queue[i:]}
		case <-WaitBetweenRequests(250, 500):
			requestChan <- request
			log.Printf("Scheduled request %d/%d", i, len(queue))
		}
	}
	// Close the channel to let the workers know that there are no more requests
	close(requestChan)
	log.Println("Waiting for workers to finish")
	// Wait for all workers to finish
	for i := 0; i < NWORKERS; i++ {
		<-finished
	}
	log.Println("Finished performing requests")
	return scraping.BatchResult{results, state.config.myself, nil}
}

// see: https://intoli.com/blog/not-possible-to-block-chrome-headless/
const script = `(function(w, n, wn) {
  // Pass the Webdriver Test.
  Object.defineProperty(n, 'webdriver', {
    get: () => false,
  });

  // Pass the Plugins Length Test.
  // Overwrite the plugins property to use a custom getter.
  Object.defineProperty(n, 'plugins', {
    // This just needs to have length > 0 for the current test,
    // but we could mock the plugins too if necessary.
    get: () => [1, 2, 3, 4, 5],
  });

  // Pass the Languages Test.
  // Overwrite the plugins property to use a custom getter.
  Object.defineProperty(n, 'languages', {
    get: () => ['en-US', 'en'],
  });

  // Pass the Chrome Test.
  // We can mock this in as much depth as we need for the test.
  w.chrome = {
    runtime: {},
  };

  // Pass the Permissions Test.
  const originalQuery = wn.permissions.query;
  return wn.permissions.query = (parameters) => (
    parameters.name === 'notifications' ?
      Promise.resolve({ state: Notification.permission }) :
      originalQuery(parameters)
  );

})(window, navigator, window.navigator);`
func SpawnChrome() error {
	opts := chromedp.DefaultExecAllocatorOptions[:]
	opts = append(opts, chromedp.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/87.0.4280.88 Safari/537.36"))
	opts = append(opts, chromedp.WindowSize(1920, 1080))
	opts = append(opts, chromedp.NoFirstRun)
	opts = append(opts, chromedp.NoDefaultBrowserCheck)
	opts = append(opts, chromedp.Headless)
	// Use a tor proxy to perform requests
	if USE_TOR {
		opts = append(opts, chromedp.ProxyServer("socks5://localhost:9050"))
	}
	cx, _ := chromedp.NewExecAllocator(context.Background(), opts...)
	// defer cancel()
	// Create context
	ctx, _ := chromedp.NewContext(cx)
	// defer cancel()

	log.Printf("Allocating context")
	// Allocate the context (actually runs the browser)
	if err := chromedp.Run(ctx); err != nil {
		log.Fatalf("Unexpected error in ExtractScripts when allocating context: %v\n", err)
	}
	state.chromeContext = ctx
	return nil
}

func ExtractScripts(worker int, request scraping.Request) (scraping.Result, error) {
	log.Printf("[worker-%d] Extracting scripts from %v\n", worker, request.URL)
	result := scraping.Result{request.URL, false, false, false, make([]string, 0), make([]string, 0)}

	// Create new tab
	ctxTab, cancel := chromedp.NewContext(state.chromeContext)
	defer cancel()

	ctx, cancel := context.WithTimeout(ctxTab, (TIMEOUT_SECONDS+5) * time.Second)
	defer cancel()

	log.Printf("[worker-%d] Allocating tab", worker)
	// Allocate the context (actually runs the browser)
	if err := chromedp.Run(ctx); err != nil {
		log.Printf("[worker-%d] Unexpected error in ExtractScripts when allocating context: %v\n", worker, err)
		return result, err
	}

	log.Printf("[worker-%d] Enable debugger", worker)
	// Enable the debugger
	c := chromedp.FromContext(ctx)
	if _, err := debugger.Enable().Do(cdp.WithExecutor(ctx, c.Target)); err != nil {
		log.Printf("[worker-%d] Unexpected error in ExtractScripts when enabling debugger: %v\n", worker, err)
		return result, err
	}

	// Setup headless detection prevention mechanism
	if PREVENT_HEADLESS_DETECTION {
		if err := chromedp.Run(ctx,
			chromedp.ActionFunc(func(ctx context.Context) error {
				_, err := page.AddScriptToEvaluateOnNewDocument(script).Do(ctx)
				if err != nil {
					return err
				}
				return err
			}),
		); err != nil {
			log.Printf("Cannot instantiate headless chrome as needed: %v", err)
			result.Failure = true
			return result, err
		}
	}


	log.Printf("[worker-%d] Listen for EvenScriptParsed events", worker)
	// Listen for EventScriptParsed events
	chromedp.ListenTarget(ctx, func(ev interface{}) {
		if ev, ok := ev.(*debugger.EventScriptParsed); ok {
			if ev.ScriptLanguage == "WebAssembly" {
				log.Printf("Script found: %v", ev.URL)
				result.Scripts = append(result.Scripts, ev.URL)
			}
		}
	})

	log.Printf("[worker-%d] Setup timeout", worker)
	// Setup a timeout
	ctxWithTimeout, cancel := context.WithTimeout(ctx, TIMEOUT_SECONDS * time.Second)
	defer cancel()

	log.Printf("[worker-%d] Visit the page", worker)
	// Actually visits the page
	var realurl string
	if err := chromedp.Run(ctxWithTimeout,
		chromedp.Navigate(request.URL),
		chromedp.WaitReady("body"),
		// Wait 5 extra seconds after page has loaded to ensure other scripts have loaded as well
		chromedp.Sleep(5 * time.Second),
		chromedp.Location(&realurl),
	); err != nil {
		if err == context.DeadlineExceeded {
			log.Printf("[worker-%d] Deadline exceeded (timeout) when visiting: %v\n", worker, request.URL)
			result.Timeout = true
		} else if fmt.Sprintf("%v", err) == "page load error net::ERR_NAME_NOT_RESOLVED" { // Ugly, but I don't see how else to do it
			log.Printf("[worker-%d] DNS error in ExtractScripts for %v: %v\n", worker, request.URL, err)
			result.DNSError = true
		} else {
			log.Printf("[worker-%d] Unexpected error in ExtractScripts when visiting %v: %v\n", worker, request.URL, err)
			result.Failure = true
		}
		return result, nil
	}

	log.Printf("[worker-%d] Extract URLs", worker)
	if request.TopLevel {
		realURL, err := url.Parse(realurl)
		if err != nil {
			log.Printf("[worker-%d] Unexpected error in ExtractScripts when retrieving url of %v: %v\n", worker, request.URL, err)
			result.Failure = true
			return result, nil
		}
		realURLTLD, err := tld.Parse(realurl)
		if err != nil {
			log.Printf("[worker-%d] Unexpected error in ExtractScripts when parsing TLD of %v: %v\n", worker, request.URL, err)
			result.Failure = true
			return result, nil
		}
		var nodes []*cdp.Node
		ctxWithTimeout, cancel := context.WithTimeout(ctx, 2 * time.Second) // Put an extra timeout here, it seems that this sometimes blocks?
		defer cancel()
		if err := chromedp.Run(ctxWithTimeout,
			chromedp.Nodes("a", &nodes, chromedp.ByQueryAll, chromedp.AtLeast(0)),
		); err != nil {
			log.Printf("[worker-%d] Unexpected error in ExtractScripts when extracting links of %v: %v\n", worker, request.URL, err)
			result.Failure = true
			return result, nil
		}

		var urls []string
		// Extract links to the same domain and prefix them if necessary
		for _, node := range nodes {
			href, present := node.Attribute("href")
			if present {
				parsed, err := url.Parse(href)
				if err != nil {
					// Invalid links in web pages are plausible, so ignore it
					continue
				}
				if parsed.Host == "" {
					// Prefix the URL if there's no host (i.e., it is a relative path)
					parsed.Scheme = realURL.Scheme
					parsed.Host = realURL.Host
				}
				tld, err := tld.Parse(parsed.String())
				if err != nil {
					continue
				}
				if tld.Domain == realURLTLD.Domain && tld.TLD == realURLTLD.TLD {
					// Only add the URL to our list if it is to the same domain
					urls = append(urls, parsed.String())
				}
			}
		}
		// Select 3 URLs at random
		if len(urls) <= URLS_TO_EXTRACT {
			// Did not parse more than 3 URLS, return everything parsed instead of selecting
			result.URLs = urls
		} else {
			vs := rand.Perm(len(urls))
			for _, i := range vs[:URLS_TO_EXTRACT] {
				result.URLs = append(result.URLs, urls[i])
			}
		}
	}
	log.Printf("[worker-%d] Finished extracting scripts from %v", worker, request.URL)
	return result, nil
}
