package main

import (
	"cloud.google.com/go/datastore"
	"fmt"
	"github.com/cevaris/status/report"
	"github.com/cevaris/timber"
	"log"
	"net/http"
	"os"

	"google.golang.org/appengine"
)

var logger = timber.NewAppEngineLogger()

func main() {
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/reports.json", getReports)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
		log.Printf("Defaulting to port %s", port)
	}

	log.Printf("Listening on port %s", port)
	appengine.Main()
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	_, err := fmt.Fprint(w, "Hello, World!")
	if err != nil {
		fmt.Println("failed to write", err)
	}
}


func getReports(w http.ResponseWriter, r *http.Request) {
	projectID := os.Getenv("PROJECT_ID")
	ctx := appengine.NewContext(r)
	dsClient, err := datastore.NewClient(ctx, projectID)
	if err != nil {
		panic(err)
	}


	keys := []*datastore.Key {
		datastore.NameKey("ApiReportMin", "aws_us_west_2_s3_read_file:1555340160", nil),
		datastore.NameKey("ApiReportMin", "aws_us_west_2_s3_read_file:1555340220", nil),
	}

	var reports = make([]report.ApiReport, 2)
	err = dsClient.GetMulti(ctx, keys, reports)
	if err != nil {
		panic(err)
	}

	logger.Info(ctx, "count", fmt.Sprintf("%+v", reports))

	_, writeErr := fmt.Fprint(w, reports)
	if writeErr != nil {
		fmt.Println("failed to write", writeErr)
	}
}
