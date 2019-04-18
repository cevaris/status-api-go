package main

import (
	"cloud.google.com/go/datastore"
	"errors"
	"fmt"
	"github.com/cevaris/status/report"
	"github.com/cevaris/timber"
	"log"
	"net/http"
	"os"
	"time"

	"google.golang.org/appengine"
)

var logger = timber.NewAppEngineLogger()
var projectID = os.Getenv("PROJECT_ID")

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
	ctx := appengine.NewContext(r)
	reportName, err := GetString(ctx, r, "name")
	if err != nil {
		serializeErr(ctx, w, err)
		return
	}

	fromTime, err := GetTime(ctx, r, "from")
	if err != nil {
		serializeErr(ctx, w, err)
		return
	}

	toTime, err := GetTime(ctx, r, "to")
	if err != nil {
		serializeErr(ctx, w, err)
		return
	}

	if fromTime.After(toTime) {
		serializeErr(ctx, w, errors.New(fmt.Sprintf("'from' param value (%s) must be before 'to' param value (%s)", fromTime, toTime)))
		return
	}

	logger.Info(ctx, reportName, fromTime, toTime)

	dsClient, err := datastore.NewClient(ctx, projectID)
	if err != nil {
		serializeErr(ctx, w, err)
		return
	}

	//newKeys := make([]*datastore.Key, 0)
	//for d := fromTime; d.Day() == toTime.Day(); d = d.Add(time.Hour * 24) {
	//	logger.Info(ctx, "day", d.Day())
	//	for h := fromTime; h.Hour() == toTime.Hour(); h = h.Add(time.Hour) {
	//		logger.Info(ctx, "hour", d.Hour())
	//		for m := fromTime; m.Minute() == toTime.Minute(); m = m.Add(time.Minute) {
	//			logger.Info(ctx, "min", d.Minute())
	//			key := datastore.NameKey(
	//				report.KindApiReportMin,
	//				fmt.Sprintf("%s:%d", reportName, report.UTCMinute(m)),
	//				nil,
	//			)
	//			newKeys = append(newKeys, key)
	//		}
	//	}
	//}

	keys := make([]*datastore.Key, 0)
	for d := fromTime; d.Unix() < toTime.Unix(); d = d.Add(time.Minute) {
		key := datastore.NameKey(
			report.KindApiReportMin,
			fmt.Sprintf("%s:%d", reportName, report.UTCMinute(d).Unix()),
			nil,
		)
		keys = append(keys, key)
	}

	var reports = make([]report.ApiReport, len(keys))
	err = dsClient.GetMulti(ctx, keys, reports)
	if err != nil {
		if me, ok := err.(datastore.MultiError); ok {
			logger.Error(ctx, "got here", err.Error())
			for i, merr := range me {
				if merr == datastore.ErrNoSuchEntity {
					reports[i] = report.ApiReport{}
				}
			}
		} else {
			serializeErr(ctx, w, err)
			return
		}
	}

	var presentable = make([]report.ApiReportJson, 0)
	for _, x := range reports {
		if x.Name != "" { // nil api report
			presentable = append(presentable, x.PresentJson())
		}
	}

	//logger.Info(ctx, "count", fmt.Sprintf("%+v", reports))
	serializeData(ctx, w, presentable, true)
}
