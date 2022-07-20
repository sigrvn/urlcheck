package main

import (
	"bytes"
	"io"
	"net/http"
	"strings"
	"time"
)

type urlStatus struct {
	isBroken     bool
	statusCode   int
	responseTime time.Duration
}

type worker struct {
	urls <-chan string
}

func (w *worker) execute() {
	for url := range w.urls {
		logger.Printf("Fetching links for url '%s'\n", url)

		if !strings.HasPrefix(url, "http") {
			logger.Printf("No protocol specified for url '%s', assuming HTTPS\n", url)
			url = "https://" + url
		}

		resp, err := http.Get(url)
		if err != nil {
			logger.Print(err)
			continue
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			logger.Print(err)
			continue
		}
		resp.Body.Close()

		var totalTime time.Duration
		totalURLs, brokenURLs := 0, 0
		hrefLines := bytes.Split(body, []byte("href=\""))
		for _, line := range hrefLines[1:] {
			end := bytes.IndexRune(line, '"')
			if end == -1 {
				logger.Printf("invalid href in body of HTML\n")
				continue
			}

			href := string(line[:end])
			nextURL := extractURL(url, href)
			status, err := getURLStatus(nextURL)
			if err != nil {
				logger.Print(err)
				continue
			}

			if status.isBroken {
				logger.Printf("BROKEN(%d) '%s' (response time: %s)\n",
					status.statusCode, nextURL, status.responseTime.String())
				brokenURLs++
			} else {
				logger.Printf("OK(%d) '%s' (response time: %s)\n",
					status.statusCode, nextURL, status.responseTime.String())
			}
			totalURLs++
			totalTime += status.responseTime
		}

		logger.Printf("Finished checking urls for '%s' in %s.\n", url, totalTime.String())
		logger.Printf("\tChecked %d urls, %d OK, %d BROKEN\n", totalURLs, totalURLs-brokenURLs, brokenURLs)
	}
}

func extractURL(rawURL, href string) string {
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
