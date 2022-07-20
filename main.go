package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

const (
	LoggerPrefix          string        = "[urlcheck] | "
	DefaultOutputFilename string        = "links.txt"
	DefaultNumWorkers     int           = 1
	DefaultTimeout        time.Duration = time.Second * 30
)

var (
	silent          bool
	followRedirects bool // Not yet implemented
	output          string
	timeout         time.Duration // Not yet implemented
	numWorkers      int
	logger          *log.Logger
)

type urlStatus struct {
	isBroken     bool
	statusCode   int
	responseTime time.Duration
}

func main() {
	args := parseArgs()
	initLogger()

	logger.Printf("Using %d worker(s)", numWorkers)

	var wg sync.WaitGroup
	wg.Add(numWorkers)

	urls := make(chan string)
	for i := 0; i < numWorkers; i++ {
		go func() {
			worker(urls)
			wg.Done()
		}()
	}

	for _, arg := range args {
		urls <- arg
	}
	close(urls)

	wg.Wait()
}

func parseArgs() []string {
	flag.BoolVar(&silent, "silent", false, "Suppress logging to stdout")
	flag.BoolVar(&followRedirects, "follow-redirects", false, "Follow links that respond with a redirect status code")
	flag.StringVar(&output, "output", DefaultOutputFilename, "Set output file for data")
	flag.DurationVar(&timeout, "timeout", DefaultTimeout, "Set timeout duration for HTTP requests")
	flag.IntVar(&numWorkers, "workers", DefaultNumWorkers, "Set number of workers to process data")
	flag.Parse()

	if len(flag.Args()) == 0 {
		log.Fatal("error: no urls were provided")
	}

	if timeout == 0 {
		timeout = DefaultTimeout
	}

	if numWorkers == 0 {
		numWorkers = DefaultNumWorkers
	}

	return flag.Args()
}

// TODO: Find a way to synchronize output for logging when workers are overlapping
func initLogger() {
	outputFile, err := os.Create(output)
	if err != nil {
		log.Fatalf("error: couldn't create output file '%s': %s\n", output, err.Error())
	}

	var w io.Writer
	if silent {
		w = outputFile
	} else {
		w = io.MultiWriter(outputFile, os.Stdout)
	}

	logger = log.New(w, LoggerPrefix, log.LstdFlags|log.Lmsgprefix)
}

func worker(urls <-chan string) {
	for url := range urls {
		err := checkURL(url)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func checkURL(rawURL string) error {
	logger.Printf("Fetching links for url '%s'\n", rawURL)

	if !strings.HasPrefix(rawURL, "http") {
		logger.Printf("No protocol specified for url '%s', assuming HTTPS", rawURL)
		rawURL = "https://" + rawURL
	}

	resp, err := http.Get(rawURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var totalTime time.Duration
	totalURLs, brokenURLs := 0, 0
	hrefLines := bytes.Split(body, []byte("href=\""))
	for _, line := range hrefLines[1:] {
		end := bytes.IndexRune(line, '"')
		if end == -1 {
			return fmt.Errorf("invalid href in body of HTML")
		}

		href := string(line[:end])
		url := extractURL(rawURL, href)
		status, err := getURLStatus(url)
		if err != nil {
			return err
		}

		if status.isBroken {
			logger.Printf("BROKEN(%d) '%s' (response time: %s)\n",
				status.statusCode, url, status.responseTime.String())
			brokenURLs++
		} else {
			logger.Printf("OK(%d) '%s' (response time: %s)\n",
				status.statusCode, url, status.responseTime.String())
		}
		totalURLs++
		totalTime += status.responseTime
	}

	logger.Printf("Finished checking urls for '%s' in %s.\n", rawURL, totalTime.String())
	logger.Printf("\tChecked %d urls, %d OK, %d BROKEN\n", totalURLs, totalURLs-brokenURLs, brokenURLs)
	return nil
}

func extractURL(rawURL, href string) string {
	// fmt.Printf("rawURL = %s, href = %s\n", rawURL, href)
	if strings.HasPrefix(href, "http") {
		return href
	}

	if strings.HasPrefix(href, "//") {
		protocol := rawURL[:strings.Index(rawURL, ":")+1]
		return protocol + href
	}

	if strings.HasPrefix(href, "/") {
		dotIdx := strings.LastIndex(rawURL, ".")
		slashIdx := strings.Index(rawURL[dotIdx:], "/")
		if slashIdx == -1 {
			slashIdx = 0
		}
		baseURL := rawURL[:dotIdx+slashIdx]

		return baseURL + href
	}

	if strings.HasPrefix(href, "#") || strings.HasPrefix(href, "?") {
		return rawURL + href
	}

	if strings.HasSuffix(rawURL, "/") {
		return rawURL + href
	}

	return rawURL + "/" + href
}

func getURLStatus(url string) (urlStatus, error) {
	start := time.Now()
	resp, err := http.Get(url)
	if err != nil {
		return urlStatus{}, err
	}
	defer resp.Body.Close()
	elapsed := time.Since(start)

	return urlStatus{
		isBroken:     resp.StatusCode > 399,
		statusCode:   resp.StatusCode,
		responseTime: elapsed,
	}, nil
}
