package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var (
	// Flags to accept config-file path.
	fConfigPath = flag.String("config-path", "./configs/config.yaml", "Path to the file")

	// Initialized before the server gets started.
	gConfig *Config

	// Initialized before the server gets started.
	gOpManager *OperationManager
)

type ProxyHandler struct {
	http.Handler
}

func (ph ProxyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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

// ProxyServer represents a simple proxy server
type ProxyServer struct {
	port     string
	server   *http.Server
	shutdown chan os.Signal
}

// NewProxyServer creates a new ProxyServer instance
func NewProxyServer(port string) *ProxyServer {
	return &ProxyServer{
		port:     port,
		shutdown: make(chan os.Signal, 1),
	}
}

// Start starts the proxy server.
func (ps *ProxyServer) Start() {
	ps.server = &http.Server{
		Addr:    ":" + ps.port,
		Handler: ProxyHandler{},
	}

	// Start the server in a new goroutine
	go func() {
		log.Printf("Proxy server started on port %s\n", ps.port)
		if err := ps.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// Handle graceful shutdown
	signal.Notify(ps.shutdown, syscall.SIGINT, syscall.SIGTERM)
	<-ps.shutdown
	log.Println("Shutting down proxy server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := ps.server.Shutdown(ctx); err != nil {
		log.Fatalf("Proxy server forced to shutdown: %v", err)
	} else {
		log.Println("Proxy server exiting")
	}
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
	ps := NewProxyServer("8080")
	ps.Start()
}
