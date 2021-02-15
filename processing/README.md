1. Put all scraping results in zip format in this directory
2. Combine them: `./combine.sh`
3. Validate DNSes `go run checkdns.go` (will heavily use the network, requires dnserrors.log, scripts.log and noscripts.log)
4. Extract domains and pages to reschedule: `go run extract.go`
This results in:
  - two new files of URLS for which the scraping failed
    - `urls-toplevel.log` for top-level URLS (for which links have to be followed)
    - `urls-nottoplevel.log` for other URLS
  - the `scripts.log` file listing all pages that contain a WebAssembly script
  - the `noscripts.log` file listing all page that do not contain WebAssembly

Scripts can then be extracted from the `scripts.log` file using `go run findscripts.go`
