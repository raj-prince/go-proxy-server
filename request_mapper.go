package main

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/raj-prince/go-proxy-server/util"
)

type RequestType string

const (
	XmlRead     RequestType = "XmlRead"
	JsonStat    RequestType = "JsonStat"
	JsonDelete  RequestType = "JsonDelete"
	JsonUpdate  RequestType = "JsonUpdate"
	JsonCreate  RequestType = "JsonCreate"
	JsonCopy    RequestType = "JsonCopy"
	JsonList    RequestType = "JsonList"
	JsonCompose RequestType = "JsonCompose"
	Unknown     RequestType = "Unknown"
)

func deduceRequestType(r *http.Request) RequestType {
	path := r.URL.Path
	method := r.Method

	// Generic JSON API format:
	// https://storage.googleapis.com/storage/v1/b/)<bucket-name>/o/<object-name>
	if strings.Contains(path, "/storage/v1") {
		switch {
		case method == http.MethodGet:
			return JsonStat
		case method == http.MethodPost:
			return JsonCreate
		case method == http.MethodPut:
			return JsonUpdate
		default:
			return Unknown
		}
	} else { // Assuming XML to start.
		switch {
		case method == http.MethodGet:
			return XmlRead
		default:
			return Unknown
		}
	}
}

func handleXMLRead(r *http.Request) error {
	plantOp := gOpManager.retrieveOperation(XmlRead)
	if plantOp != "" {
		testID := util.CreateRetryTest(gConfig.TargetHost, map[string][]string{"storage.objects.get": {plantOp}})
		r.Header.Set("x-retry-test-id", testID)
	}
	return nil
}

func handleRequest(requestType RequestType, r *http.Request) error {
	switch requestType {
	case XmlRead:
		return handleXMLRead(r)
	case JsonStat:
		fmt.Println("No handling for...json stat")
		return nil
	default:
		fmt.Println("No handling for unknown operation")
		return nil
	}
}
