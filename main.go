package main

import (
        "fmt"
        "io"
        "net/http"
        "os"
)

func handler(w http.ResponseWriter, r *http.Request) {
        // Create a new request to the target server
        targetURL := "localhost" // Replace with your target server URL
        req, err := http.NewRequest(r.Method, targetURL, r.Body)
        if err != nil {
                http.Error(w, err.Error(), http.StatusInternalServerError)
                return
        }

        // Copy 1  headers from the original request
        for name, values := range r.Header {
                for _, value := range values {
                        req.Header.Add(name, value)
                }
        }

        // Add a custom header to the request
        req.Header.Set("X-My-Custom-Header", "This request has been proxied!")

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