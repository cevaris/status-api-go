package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"
)

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
func serializeData(ctx context.Context, w http.ResponseWriter, data interface{}, isPrettyJSON bool) {
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

// SerializeErr writes exceptional JSON responses
func serializeErr(ctx context.Context, w http.ResponseWriter, err error) {
	response := Response{Status: "error", Message: err.Error()}
	b, err := marshal(response, true)
	if err != nil {
		logger.Error(ctx, "failed to serialize json", err, "for", response)
		http.Error(w, serverErrorJSON, 500)
		return
	}
	http.Error(w, string(b), 400)
}

// GetString parses required string
func GetString(ctx context.Context, r *http.Request, paramName string) (string, error) {
	str := r.URL.Query().Get(paramName)
	if len(str) == 0 {
		msg := fmt.Sprintf("missing require parameter '%s'", paramName)
		logger.Error(ctx, msg)
		return "", errors.New(msg)
	}
	return str, nil
}

// GetTime parses required string
func GetTime(ctx context.Context, r *http.Request, paramName string) (time.Time, error) {
	timeStr := r.URL.Query().Get(paramName)
	if len(timeStr) == 0 {
		msg := fmt.Sprintf("missing require parameter '%s'", paramName)
		logger.Error(ctx, msg)
		return time.Time{}, errors.New(msg)
	}
	parsedTime, err := time.Parse(time.RFC3339, timeStr)
	if err != nil {
		msg := fmt.Sprintf("'%s' is not in RFC3339 format", timeStr)
		logger.Error(ctx, msg)
		return time.Time{}, err
	}
	return parsedTime, nil
}
