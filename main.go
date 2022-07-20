package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"sync"
	"time"
)

const (
	DefaultOutputFilename string        = "links.txt"
	DefaultNumWorkers     int           = 1
	DefaultTimeout        time.Duration = time.Second * 30
	LoggerPrefix          string        = "[urlcheck] | "
)

var (
	silent          bool
	followRedirects bool // Not yet implemented
	output          string
	timeout         time.Duration // Not yet implemented
	numWorkers      int
	logger          *log.Logger
)

func main() {
	args := parseArgs()

	logger.Printf("Using %d worker(s)\n", numWorkers)
	var wg sync.WaitGroup
	wg.Add(numWorkers)

	urls := make(chan string)
	for id := 0; id < numWorkers; id++ {
		go func() {
			w := &worker{urls}
			w.execute()
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
		fmt.Println("error: no urls were provided")
		os.Exit(1)
	}

	if timeout == 0 {
		timeout = DefaultTimeout
	}

	if numWorkers == 0 {
		numWorkers = DefaultNumWorkers
	}

	outputFile, err := os.Create(output)
	if err != nil {
		fmt.Printf("error: couldn't create output file '%s': %s\n", output, err.Error())
		os.Exit(1)
	}

	var w io.Writer
	if silent {
		w = outputFile
	} else {
		w = io.MultiWriter(outputFile, os.Stdout)
	}
	logger = log.New(w, LoggerPrefix, log.LstdFlags|log.Lmsgprefix)

	return flag.Args()
}
