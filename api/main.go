package main

import (
	"cloud.google.com/go/datastore"
	"context"
	"encoding/json"
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

	keys := []*datastore.Key{
		datastore.NameKey("ApiReportMin", "aws_us_west_2_s3_read_file:1555340160", nil),
		datastore.NameKey("ApiReportMin", "aws_us_west_2_s3_read_file:1555340220", nil),
	}

	var reports = make([]report.ApiReport, len(keys))
	err = dsClient.GetMulti(ctx, keys, reports)
	if err != nil {
		panic(err)
	}

	var presentable = make([]report.ApiReportJson, len(keys))
	for i, _ := range reports {
		presentable[i] = reports[i].PresentJson()
	}

	logger.Info(ctx, "count", fmt.Sprintf("%+v", reports))

	//_, writeErr := fmt.Fprint(w, reports)
	//if writeErr != nil {
	//	fmt.Println("failed to write", writeErr)
	//}
	SerializeData(ctx, w, presentable, true)
}

type Response struct {
	Status  string      `json:"status,omitempty"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

var serverError = Response{
	Status:  "error",
	Message: "internal server error",
}
var serverErrorJSONBytes, _ = marshal(serverError, true)
var serverErrorJSON = string(serverErrorJSONBytes)

func marshal(data interface{}, prettyJSON bool) ([]byte, error) {
	if prettyJSON {
		return json.MarshalIndent(data, "", "    ")
	}
	return json.Marshal(data)
}
func SerializeData(ctx context.Context, w http.ResponseWriter, data interface{}, isPrettyJSON bool) {
	response := Response{Status: "ok", Data: data}
	b, err := marshal(response, isPrettyJSON)
	if err != nil {
		logger.Error(ctx, "failed to serialize json", err, "for", response)
		http.Error(w, serverErrorJSON, 500)
		return
	}
	_, err = w.Write(b)
	if err != nil {
		logger.Error(ctx, "failed to serialize json", err, "for", response)
		http.Error(w, serverErrorJSON, 500)
		return
	}
	w.Header().Set("Content-Type", "application/json")
}
