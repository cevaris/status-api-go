package main

import (
	"cloud.google.com/go/datastore"
	"errors"
	"fmt"
	"github.com/cevaris/status/report"
	"github.com/cevaris/timber"
	"github.com/julienschmidt/httprouter"
	"log"
	"net/http"
	"net/http/pprof"
	"os"
	"time"

	"google.golang.org/appengine"
)

var logger = timber.NewAppEngineLogger()
var projectID = os.Getenv("PROJECT_ID")

func main() {

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
		log.Printf("Defaulting to port %s", port)
	}

	router := httprouter.New()
	router.GET("/reports/:ID", getReports)
	router.GET("/debug/pprof/goroutine", func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) { pprof.Index(w, r) })
	http.Handle("/", router)

	log.Printf("Listening on port %s", port)
	appengine.Main()
}

func getReports(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	ctx := appengine.NewContext(r)
	reportName := ps.ByName("ID")
	if len(reportName) == 0 {
		serializeErr(ctx, w, errors.New("missing require parameter 'ID'"))
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
