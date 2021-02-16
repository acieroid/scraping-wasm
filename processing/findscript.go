package main

import (
	"os"
	"net/url"
	"bufio"
	"fmt"
	"log"
	"context"
	"time"
	"io/ioutil"
	"crypto/sha256"
	"encoding/hex"
	"encoding/csv"
	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/debugger"
	"github.com/chromedp/cdproto/runtime"
	"github.com/chromedp/chromedp"
)

const TIMEOUT_SECONDS = 60
var currentItem = 1
var totalItems = 0

type ScriptInfo struct {
	Domain string
	PageURL string
	ScriptURL string
	ScriptHash string
	CalledFrom string
	CalledFromHash string
}

type ScriptInternalInfo struct {
	URL string
	ScriptID runtime.ScriptID
	ExecutionContextID runtime.ExecutionContextID
	StackTrace *runtime.StackTrace
}

func SaveToFile(content []byte, file string) {
	err := ioutil.WriteFile(file, content, 0644)
	if err != nil {
		log.Fatalf("Can't write to %s: %v", file, err)
	}
}

func HashOf(script []byte) string {
	hash := sha256.Sum256(script)
	return hex.EncodeToString(hash[:])
}

func SaveSourceGetHash(source string, bytecode []byte) string {
	if (len(source) > 0) {
		bytes := []byte(source)
		hash := HashOf(bytes)
		SaveToFile([]byte(source), fmt.Sprintf("source/%s.js", hash))
		return hash
	}
	if (len(bytecode) > 0) {
		hash := HashOf(bytecode)
		SaveToFile(bytecode, fmt.Sprintf("bytecode/%s.wasm", hash))
		return hash
	}
	return ""
}

func URLDomain(link string) string {
	u, err := url.Parse(link)
	if err != nil {
		log.Fatalf("Malformed URL: %v", err)
	}
	return u.Host
}

func ExtractScriptInfo(url string) []ScriptInfo {
	log.Printf("[%d/%d] Extracting scripts from %v\n", currentItem, totalItems, url)
	result := make([]ScriptInfo, 0)

	opts := chromedp.DefaultExecAllocatorOptions[:]
	opts = append(opts, chromedp.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/87.0.4280.88 Safari/537.36"))
	opts = append(opts, chromedp.WindowSize(1920, 1080))
	opts = append(opts, chromedp.NoFirstRun)
	opts = append(opts, chromedp.NoDefaultBrowserCheck)
	opts = append(opts, chromedp.Headless)
	cx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()
	// Create context
	ctx, cancel := chromedp.NewContext(cx)
	defer cancel()

	// Allocate the context (actually runs the browser)
	if err := chromedp.Run(ctx); err != nil {
		log.Printf("Unexpected error in ExtractScripts when allocating context: %v\n", err)
	}
	chromeContext := ctx

	// Create new tab
	ctxTab, cancel := chromedp.NewContext(chromeContext)
	defer cancel()

	ctx, cancel = context.WithTimeout(ctxTab, (TIMEOUT_SECONDS+5) * time.Second)
	defer cancel()

	// Allocate the context (actually runs the browser)
	if err := chromedp.Run(ctx); err != nil {
		log.Printf("Unexpected error in ExtractScripts when allocating context: %v\n", err)
		return result
	}

	// Enable the debugger
	c := chromedp.FromContext(ctx)
	if _, err := debugger.Enable().Do(cdp.WithExecutor(ctx, c.Target)); err != nil {
		log.Printf("Unexpected error in ExtractScripts when enabling debugger: %v\n", err)
		return result
	}

	var scriptsFound []ScriptInternalInfo
	// Listen for EventScriptParsed events
	chromedp.ListenTarget(ctx, func(ev interface{}) {
		if ev, ok := ev.(*debugger.EventScriptParsed); ok {
			if ev.ScriptLanguage == "WebAssembly" {
				// log.Printf("Script found: %v", ev.URL)
				script := ScriptInternalInfo{ev.URL, ev.ScriptID, ev.ExecutionContextID, ev.StackTrace}
				scriptsFound = append(scriptsFound, script)
				// log.Printf("executionContextAuxData: %v", string(ev.ExecutionContextAuxData))
			}
		}
	})

	executionContexts := make(map[runtime.ExecutionContextID]runtime.ExecutionContextDescription)
	chromedp.ListenTarget(ctx, func(ev interface{}) {
		if ev, ok := ev.(*runtime.EventExecutionContextCreated); ok {
			executionContexts[ev.Context.ID] = *ev.Context;
		}
	})

	// Setup a timeout
	ctxWithTimeout, cancel := context.WithTimeout(ctx, TIMEOUT_SECONDS * time.Second)
	defer cancel()

	// Actually visits the page
	var realurl string
	if err := chromedp.Run(ctxWithTimeout,
		chromedp.Navigate(url),
		chromedp.WaitReady("body"),
		// Wait 5 extra seconds after page has loaded to ensure other scripts have loaded as well
		chromedp.Sleep(5 * time.Second),
		chromedp.Location(&realurl),
	); err != nil {
		if err == context.DeadlineExceeded {
			log.Printf("Deadline exceeded (timeout) when visiting: %v\n", url)
		} else if fmt.Sprintf("%v", err) == "page load error net::ERR_NAME_NOT_RESOLVED" { // Ugly, but I don't see how else to do it
			log.Printf("DNS error in ExtractScripts for %v: %v\n", url, err)
		} else {
			log.Printf("Unexpected error in ExtractScripts when visiting %v: %v\n", url, err)
		}
		return result
	}

	log.Printf("[%d/%d] Finished extracting scripts from %v, now gathering extra info", currentItem, totalItems, url)

	for _, script := range(scriptsFound) {
		scriptInfo := ScriptInfo{URLDomain(url), url, script.URL, "", "", ""}
		source, bytecode, err := debugger.GetScriptSource(script.ScriptID).Do(cdp.WithExecutor(ctx, c.Target))
		if err != nil {
			log.Printf("Unexpected error when extracting info from page %s: %v", url, err)
		}
		scriptInfo.ScriptHash = SaveSourceGetHash(source, bytecode)

		if script.StackTrace != nil && len((*script.StackTrace).CallFrames) > 0 {
			callFrame := (*script.StackTrace).CallFrames[0]
			// log.Printf("called from: %v, %v, %d:%d", callFrame.FunctionName, callFrame.URL, callFrame.LineNumber, callFrame.ColumnNumber)
			source, bytecode, err := debugger.GetScriptSource(callFrame.ScriptID).Do(cdp.WithExecutor(ctx, c.Target))
			if err != nil {
				log.Printf("Unexpected error: %v", err)
			}
			scriptInfo.CalledFromHash = SaveSourceGetHash(source, bytecode)
			scriptInfo.CalledFrom = fmt.Sprintf("%s:%d:%d:%s", callFrame.FunctionName, callFrame.LineNumber, callFrame.ColumnNumber, callFrame.URL)
		} else {
			log.Printf("no stack trace")
		}

		executionContext := executionContexts[script.ExecutionContextID]
		log.Printf("ex context: origin=%s, name=%s, auxdata=%s", executionContext.Origin, executionContext.Name, string(executionContext.AuxData))
		result = append(result, scriptInfo)
	}
	return result
}

func ReadLines(path string) []string {
	file, err := os.Open(path)
	if err != nil {
		log.Fatalf("Cannot open file %s: %v", path, err)
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if scanner.Err() != nil {
		log.Fatalf("Cannot read from file %s: %v", path, err)
	}
	return lines
}

func AddCSVLine(path string, line []string) {
	file, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		log.Fatalf("Cannot open file %s: %v", path, err)
	}
	defer file.Close()

	w := csv.NewWriter(file)
	err = w.Write(line)
	if err != nil {
		log.Fatalf("Cannot write to file %s: %v", path, err)
	}
	w.Flush()
}

func Mkdir(dir string) {
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		log.Fatalf("Can't create dir directory: %v", dir, err)
	}
}

func ReadEntries(path string) [][]string {
	file, err := os.Open(path)
	if err != nil {
		log.Fatalf("Cannot open file %s: %v", path, err)
	}
	defer file.Close()

	r := csv.NewReader(file)
	records, err := r.ReadAll()
	if err != nil {
		log.Fatalf("Cannot read from file %s: %v", path, err)
	}
	return records
}

func AlreadyHandled(script string, entries [][]string) bool {
	for _, entry := range(entries) {
		if entry[1] == script {
			return true
		}
	}
	return false
}

func main() {
	Mkdir("source")
	Mkdir("bytecode")
	scripts := ReadLines("scripts.log")
	totalItems = len(scripts)
	found := 0
	alreadyInCSV := ReadEntries("results.csv")
	for _, script := range scripts {
		if !AlreadyHandled(script, alreadyInCSV) {
			for _, info := range ExtractScriptInfo(script) {
				found = found + 1
				AddCSVLine("results.csv", []string{info.Domain, info.PageURL, info.ScriptURL, info.ScriptHash, info.CalledFrom, info.CalledFromHash})
			}
		}
		currentItem += 1
	}
	log.Printf("Found %d new scripts", found)
}
