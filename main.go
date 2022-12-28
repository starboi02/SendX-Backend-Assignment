package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

type data struct {
	Url         string `json:"url"`
	retry_limit int
	Filename    string `json:"filename"`
	Timestamp   int64  `json:"timestamp"`
}

var pages []data
var jobs chan data

func seachPageInPages(url string) int {
	for index, page := range pages {
		if page.Url == url {
			return index
		}
	}
	return -1
}

func main() {
	// Create the downloads directory if it doesn't exist
	if _, err := os.Stat("files"); os.IsNotExist(err) {
		os.Mkdir("files", 0755)
	}
	// Create a channel for downloads jobs and a pool of workers to handle the requests
	jobs = make(chan data, 1000)
	for w := 1; w <= 5; w++ {
		go downloadWorker(jobs)
	}

	//Writing the downloads handler to handle the request
	http.HandleFunc("/pagesource", func (w http.ResponseWriter, r *http.Request) {
		// Get the URL from the request
		url := r.URL.Query().Get("url")
		if url == "" {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintln(w, "url is required")
			return
		}
	
		// Get the retry limit from the request
		retryLimit := r.URL.Query().Get("retry_limit")
	
		if retryLimit == "" {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintln(w, "retry_limit is required")
			return
		}
	
		retryLimitInt, err := strconv.Atoi(retryLimit)
	
		if err != nil {
			log.Fatal(err)
		}
	
		filename := "files/" + strconv.Itoa(len(pages)+1) + ".html"
		timestamp := time.Now().Unix()
		requestedDownload := data{Url: url, retry_limit: retryLimitInt, Filename: filename, Timestamp: timestamp}
		index := seachPageInPages(url)
	
		if index != -1 && timestamp-pages[index].Timestamp < 86400 {
			// Serve the file from the cache
			filename = pages[index].Filename
			pages[index].Timestamp = timestamp
		} else {
			// Schedule downloads of file
			jobs <- requestedDownload
			// Add it to cache
			pages = append(pages, requestedDownload)
			index = len(pages) - 1
		}
		// Constantly check if file has been downloaded and serve it
		for {
			if _, err := os.Stat(filename); os.IsNotExist(err) {
				time.Sleep(1 * time.Second)
			} else {
				fmt.Println("Serving file: " + filename)
				break
			}
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		pageJson, err := json.Marshal(pages[index])
		fmt.Println(string(pageJson))
		if err == nil {
			w.Write(pageJson)
		}
	})

	log.Fatal(http.ListenAndServe("localhost:8080", nil))
}

// Worker to handle the downloads request
func downloadWorker(jobs <-chan data) {
	// If a job is received, recieve from channel and downloads the file
	for requestedDownload := range jobs {
		fmt.Println("Downloading file: " + requestedDownload.Filename)
		downloadFile(requestedDownload)
		fmt.Println("Downloaded file: " + requestedDownload.Filename)
	}
}

func downloadFile(requestedDownload data) {
	for i := 0; i < requestedDownload.retry_limit; i++ {
		out, err := os.Create(requestedDownload.Filename)
		if err != nil {
			log.Fatal(err)
		}
		defer out.Close()

		// Getting the data
		resp, err := http.Get(requestedDownload.Url)
		if err != nil {
			log.Fatal(err)
		}
		defer resp.Body.Close()

		// Writing the body to file
		_, err = io.Copy(out, resp.Body)

		if err != nil {
			log.Fatal(err)
			continue
		}
		break
	}
}
