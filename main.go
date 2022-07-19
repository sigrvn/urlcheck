package main

import (
  "bytes"
  "io"
  "os"
  "fmt"
  "log"
  "net/http"
  "strings"
  "flag"
  "time"
)

const (
  LoggerPrefix          string = "[urlcheck] | "
  DefaultOutputFilename string = "links.txt"
  DefaultNumWorkers        int = 8
  DefaultTimeout time.Duration = time.Second * 30
)

var (
  silent bool           = false
  followRedirects bool  = false
  output string
  urls []string
  timeout time.Duration
  numWorkers int
  logger *log.Logger
)

type urlStatus struct {
  isBroken     bool
  statusCode   int
  responseTime time.Duration
}

func main() {
  parseOptions()

  for _, url := range urls {
    err := checkURL(url)
    if err != nil {
      log.Fatal(err)
    }
  }
}

func parseOptions() {
  flag.BoolVar(&silent, "silent", false, "Suppress logging to stdout")
  flag.BoolVar(&followRedirects, "follow-redirects", false, "Follow links that respond with a redirect status code")
  flag.StringVar(&output, "output", DefaultOutputFilename, "Set output file for data")
  flag.DurationVar(&timeout, "timeout", DefaultTimeout, "Set timeout duration for HTTP requests")
  flag.IntVar(&numWorkers, "workers", DefaultNumWorkers, "Set number of workers to process data")
  flag.Parse()

  for _, arg := range flag.Args() {
    urls = append(urls, arg)
  }

  outputFile, err := os.Create(output)
  if err != nil {
    log.Fatalf("couldn't create output file '%s': %s\n", output, err.Error())
  }

  var w io.Writer
  if silent {
    w = outputFile
  } else {
    w = io.MultiWriter(outputFile, os.Stdout)
  }

  logger = log.New(w, LoggerPrefix, log.LstdFlags | log.Lmsgprefix)
}

func checkURL(rawURL string) error {
  logger.Printf("Fetching links for url '%s' ...\n", rawURL)

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

  total, broken := 0, 0
  hrefLines := bytes.Split(body, []byte("href=\""))
  for _, line := range hrefLines[1:] {
    end := bytes.IndexRune(line, '"')
    if end == -1 {
      return fmt.Errorf("Invalid href in body of HTML")
    }

    href := string(line[:end])
    url := extractURL(rawURL, href)
    status, err := getURLStatus(url)
    if err != nil {
      return err
    }

    if status.isBroken {
      logger.Printf("BROKEN '%s' (code: %d, response time: %s)\n", 
        url, status.statusCode, status.responseTime.String())
      broken++
    } else {
      logger.Printf("OK '%s' (code: %d, response time: %s)\n", 
        url, status.statusCode, status.responseTime.String())
    }
    total++
  }

  logger.Printf("Finished checking urls for url '%s'.\n", rawURL)
  logger.Printf("\tChecked %d urls, %d OK, %d BROKEN\n", total, total - broken, broken)
  return nil 
}

func extractURL(rawURL, href string) string {
  // fmt.Printf("rawURL = %s, href = %s\n", rawURL, href)
  if strings.HasPrefix(href, "http") {
    return href
  } 

  if strings.HasPrefix(href, "//") {
    protocol := rawURL[:strings.Index(rawURL, ":") + 1]
    return protocol + href
  } 

  if strings.HasPrefix(href, "/") {
    dotIdx := strings.LastIndex(rawURL, ".")
    slashIdx := strings.Index(rawURL[dotIdx:], "/")
    if slashIdx == -1 {
      slashIdx = 0
    }
    baseURL := rawURL[:dotIdx + slashIdx]

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
    isBroken: resp.StatusCode > 399,
    statusCode: resp.StatusCode,
    responseTime: elapsed,
  }, nil
}

