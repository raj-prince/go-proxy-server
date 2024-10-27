package main

import (
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
)

var (
	// Flags to accept config-file path.
	fConfigPath = flag.String("config-path", "./configs/config.yaml", "Path to the file")

	// Initialized before the server gets started.
	gConfig *Config

	// Initialized before the server gets started.
	gOpManager *OperationManager
)

func handler(w http.ResponseWriter, r *http.Request) {
	requestType := deduceRequestType(r)

	targetURL := fmt.Sprintf("%s%s", gConfig.TargetHost, r.RequestURI)
	req, err := http.NewRequest(r.Method, targetURL, r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	for name, values := range r.Header {
		for _, value := range values {
			req.Header.Add(name, value)
		}
	}

	err = handleRequest(requestType, req)
	if err != nil {
		fmt.Printf("Error in handling the request: %v", err)
	}

	// Send the request to the target server
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	// Copy headers from the target server's response
	for name, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(name, value)
		}
	}

	// Copy the response body
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}

func main() {
	// Parse the command-line flags
	flag.Parse()

	var err error
	gConfig, err = parseConfigFile(*fConfigPath)
	fmt.Printf("%+v\n", gConfig)
	if err != nil {
		fmt.Printf("Parsing error: %v\n", err)
		os.Exit(1)
	}

	gOpManager = NewOperationManager(*gConfig)

	port := "8080" // You can change the port if needed
	fmt.Printf("Starting proxy server on :%s...\n", port)
	http.HandleFunc("/", handler)
	err = http.ListenAndServe(":"+port, nil)
	if err != nil {
		fmt.Println("Error starting proxy server:", err)
		os.Exit(1)
	}
}
