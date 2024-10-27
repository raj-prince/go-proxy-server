package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/raj-prince/go-proxy-server/util"
)

// RequestCategory represents the category of a request.
type RequestCategory int

var fastLatencyCount = int64(40)
var mu sync.Mutex // Global mutex

// Enum values for request categories.
const (
	StorageAPIGet RequestCategory = iota
	StorageAPIPost
	StorageAPIOther
	DirectObjectAccess
)

func getInstructions() map[string][]string {
	mu.Lock()
	defer mu.Unlock()

	if fastLatencyCount <= 0 {
		return map[string][]string{"storage.objects.get": {"stall-for-3s-after-0K"}}
	} else {
		fastLatencyCount--
		return map[string][]string{}
		//return map[string][]string{"storage.objects.get": {"stall-for-1s-after-0K"}}
	}
}

// DifferentiateRequest categorizes the request based on URI and method.
func DifferentiateRequest(r *http.Request) RequestCategory {
	uri := r.URL.Path
	method := r.Method

	if strings.Contains(uri, "/storage/v1") {
		// Requests 1, 2, and 3: API requests
		if method == http.MethodGet {
			return StorageAPIGet
		} else if method == http.MethodPost {
			return StorageAPIPost
		}
		return StorageAPIOther
	} else {
		// Request 4: Direct resource access
		return DirectObjectAccess
	}
}

func printReq(r *http.Request) {
	fmt.Printf("Request catagory:%d\n", DifferentiateRequest(r))
	fmt.Println("RequestURI: ", r.RequestURI)
	fmt.Println("Method: ", r.Method)
	fmt.Println("URL: ", r.URL.String())
	fmt.Println("Host:", r.Host)
}
func handler(w http.ResponseWriter, r *http.Request) {
	//printReq(r)
	// Create a new request to the target server
	targetHost := "localhost:9000" // Replace with your target server URL
	targetURL := fmt.Sprintf("http://%s%s", targetHost, r.RequestURI)

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

	if DifferentiateRequest(r) == DirectObjectAccess {
		testID := util.CreateRetryTest(getInstructions())
		req.Header.Set("x-retry-test-id", testID)
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
	port := "8080" // You can change the port if needed
	fmt.Printf("Starting proxy server on :%s...\n", port)
	http.HandleFunc("/", handler)
	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		fmt.Println("Error starting proxy server:", err)
		os.Exit(1)
	}
}
