// go-callvis: a tool to help visualize the call graph of a Go program.
//
package main

import (
	"flag"
	"fmt"
	"go/build"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/pkg/browser"
	"golang.org/x/tools/go/buildutil"
)

var (
	Version = "v0.4-dev"
)

var (
	focusFlag       = flag.String("focus", "main", "Focus specific package using name or import path.")
	focusMethodFlag = flag.String("method", "", "Focus on specific method and filter out all others in the focus package")
	groupFlag       = flag.String("group", "", "Grouping functions by packages and/or types [pkg, type] (separated by comma)")
	limitFlag       = flag.String("limit", "", "Limit package paths to given prefixes (separated by comma)")
	ignoreFlag      = flag.String("ignore", "", "Ignore package paths containing given prefixes (separated by comma)")
	includeFlag     = flag.String("include", "", "Include package paths with given prefixes (separated by comma)")
	nostdFlag       = flag.Bool("nostd", false, "Omit calls to/from packages in standard library.")
	nointerFlag     = flag.Bool("nointer", false, "Omit calls to unexported functions.")
	testFlag        = flag.Bool("tests", false, "Include test code.")
	debugFlag       = flag.Bool("debug", false, "Enable verbose log.")
	versionFlag     = flag.Bool("version", false, "Show version and exit.")
	httpFlag        = flag.String("http", ":7878", "HTTP service address.")
	skipBrowser     = flag.Bool("skipbrowser", false, "Skip opening browser.")
)

func init() {
	flag.Var((*buildutil.TagsFlag)(&build.Default.BuildTags), "tags", buildutil.TagsFlagDoc)
	// Graphviz options
	flag.UintVar(&minlen, "minlen", 2, "Minimum edge length (for wider output).")
	flag.Float64Var(&nodesep, "nodesep", 0.35, "Minimum space between two adjacent nodes in the same rank (for taller output).")
}

const Usage = `go-callvis: visualize call graph of a Go program.

Usage:

  go-callvis [flags] package

  Package should be main package, otherwise -tests flag must be used.

Flags:

`

func main() {
	flag.Parse()

	if *versionFlag {
		fmt.Fprintf(os.Stderr, "go-callvis %s\n", Version)
		os.Exit(0)
	}
	if *debugFlag {
		log.SetFlags(log.Lmicroseconds)
	}

	args := flag.Args()
	if flag.NArg() != 1 {
		fmt.Fprintf(os.Stderr, Usage)
		flag.PrintDefaults()
		os.Exit(2)
	}

	tests := *testFlag
	httpAddr := *httpFlag
	urlAddr := parseHTTPAddr(httpAddr)

	doAnalysis(&build.Default, tests, args)

	http.HandleFunc("/", handler)

	log.Printf("http serving at %s", urlAddr)

	if !*skipBrowser {
		go openBrowser(urlAddr)
	}

	if err := http.ListenAndServe(httpAddr, nil); err != nil {
		log.Fatal(err)
	}
}

func parseHTTPAddr(addr string) string {
	host, port, _ := net.SplitHostPort(addr)
	if host == "" {
		host = "localhost"
	}
	if port == "" {
		port = "80"
	}
	u := url.URL{
		Scheme: "http",
		Host:   fmt.Sprintf("%s:%s", host, port),
	}
	return u.String()
}

func openBrowser(url string) {
	time.Sleep(time.Millisecond * 100)
	if err := browser.OpenURL(url); err != nil {
		log.Printf("OpenURL error: %v", err)
	}
}

func logf(f string, a ...interface{}) {
	if *debugFlag {
		log.Printf(f, a...)
	}
}
