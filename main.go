package main

import (
	"flag"
	"github.com/PuerkitoBio/goquery"
	"github.com/onrik/logrus/filename"
	"fmt"
	log "github.com/sirupsen/logrus"
)

func main(){
	log.AddHook(filename.NewHook())

	url := flag.String("url", "http://foo-bar.com", "base url to scrape")
	from := flag.Int("from", 1, "page from")
	to := flag.Int("to", 10, "page to")
	concurrency := flag.Int("concurrency", 3, "amount of simultaneous workers")
	flag.Parse()
	tasks := make(chan task)
	go func(){
		for i:= *from; i<=*to; i++ {
			tasks <- task{*url, i}
		}
		close(tasks)
	}()
	results := make(chan result)
	for i := 1; i<= *concurrency; i++ {
		workerLogger := log.WithField("concurrency", concurrency)
		go func(workerLogger *log.Entry){
			for t := range tasks {
				title, err := getTitle(fmt.Sprintf("%s/%d", *url, t.Page))
				if err != nil {
					workerLogger.Error(err)
				}
				results <- result{Task: t, Title: title, Error: err}
			}
		}(workerLogger)
	}

	for r := range results {
		resultLogger := log.WithField("page", r.Task.Page)
		if r.Error != nil {
			resultLogger.Error(r.Error)
			continue
		}
		resultLogger.Println(r.Title)
 	}

}

func getTitle(url string) (string, error) {
	log.Println("requesting", url)
	d, err := goquery.NewDocument(url)
	if err != nil {
		return "", fmt.Errorf("unable to create document form url %+v: %+v", url, err)
	}
	return d.Find("title").First().Text(), nil
}

type task struct {
	Url string
	Page int
}

type result struct {
	Task task
	Title string
	Error error
}