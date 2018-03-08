package main

import (
	"flag"
	"fmt"
	"sync"

	"github.com/PuerkitoBio/goquery"
	"github.com/onrik/logrus/filename"
	log "github.com/sirupsen/logrus"
)

func main() {
	log.AddHook(filename.NewHook())

	url := flag.String("url", "http://foo-bar.com", "base url to scrape")
	from := flag.Int("from", 1, "page from")
	to := flag.Int("to", 10, "page to")
	concurrency := flag.Int("concurrency", 3, "amount of simultaneous workers")
	flag.Parse()
	wg := sync.WaitGroup{}
	wg.Add(*concurrency)
	tasks := make(chan task)
	go func() {
		for i := *from; i <= *to; i++ {
			tasks <- task{*url, i}
		}
		close(tasks)
	}()
	results := make(chan result)
	for i := 1; i <= *concurrency; i++ {
		workerLogger := log.WithField("concurrency", concurrency)
		go func(workerLogger *log.Entry) {
			for t := range tasks {
				title, err := getTitle(fmt.Sprintf("%s/%d", *url, t.Page))
				if err != nil {
					workerLogger.Error(err)
				}
				results <- result{Task: t, Title: title, Error: err}
			}
			wg.Done()
		}(workerLogger)
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	var counter = 0
	for r := range results {
		counter++
		resultLogger := log.WithFields(log.Fields{"page": r.Task.Page, "i": counter})
		if r.Error != nil {
			resultLogger.Error(r.Error)
			continue
		}
		resultLogger.Println(r.Title)
	}
}

func getTitle(url string) (string, error) {
	log.Debug("requesting", url)
	d, err := goquery.NewDocument(url)
	if err != nil {
		return "", fmt.Errorf("unable to create document form url %+v: %+v", url, err)
	}
	return d.Find("title").First().Text(), nil
}

type task struct {
	Url  string
	Page int
}

type result struct {
	Task  task
	Title string
	Error error
}
