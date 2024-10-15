package main

import (
	"bytes"
	"context"
	"crypto/rand"
	"fmt"
	"io"
	"os"
	"testing"
	"time"

	"cloud.google.com/go/storage"
	"github.com/googleapis/gax-go/v2"
	"github.com/raj-prince/go-proxy-server/util"
	"golang.org/x/sync/errgroup"
)

var (
	maxRetryDuration = 30 * time.Second

	retryMultiplier = 2.0
)

var client *storage.Client
var bucketHandle *storage.BucketHandle

// Generates size random bytes.
func generateRandomBytes(n int) []byte {
	b := make([]byte, n)
	_, _ = rand.Read(b)
	return b
}

func TestMain(m *testing.M) {
	ctx := context.Background()

	// Create client
	var err error
	client, err = util.CreateHTTPClient(ctx, false)

	if err != nil {
		fmt.Printf("while creating the client: %v", err)
		os.Exit(1)
	}
	client.SetRetry(
		storage.WithBackoff(gax.Backoff{
			Max:        maxRetryDuration,
			Multiplier: retryMultiplier,
		}),
		storage.WithPolicy(storage.RetryAlways))

	// Run the tests
	retCode := m.Run()

	// Teardown code: Run after all tests (e.g., close database connections)

	os.Exit(retCode)
}

func TestSomething(t *testing.T) {
	ctx := context.Background()

	// Setup bucket.
	project := "fake-project"
	bucket := fmt.Sprintf("http-bucket-%d", time.Now().Nanosecond())
	t.Logf("Bucket name: %s", bucket)
	if err := client.Bucket(bucket).Create(ctx, project, nil); err != nil {
		fmt.Printf("Error while creating bucket: %v", err)
		os.Exit(0)
	}

	// Setup object.
	prefix := time.Now().Nanosecond()
	objName := fmt.Sprintf("%d-object", prefix)
	t.Logf("Object name: %s", objName)
	w := client.Bucket(bucket).Object(objName).NewWriter(ctx)
	objBytes := generateRandomBytes(8 * 1024)
	if _, err := w.Write(objBytes); err != nil {
		t.Errorf("Failed while writing: %v", err)
	}
	if err := w.Close(); err != nil {
		t.Errorf("Failed while closing: %v", err)
	}
	_, err := client.Bucket(bucket).Object(objName).Attrs(ctx)
	if err != nil {
		t.Errorf("Failed while fetching object attrs")
	}

	workerCount := 10
	iterationsPerWorker := 6
	var g errgroup.Group

	// Launch worker goroutines using errgroup
	for i := 0; i < workerCount; i++ {
		g.Go(func() error {
			for j := 0; j < iterationsPerWorker; j++ {
				startTime := time.Now()
				r, err := client.Bucket(bucket).Object(objName).NewReader(ctx)
				fmt.Printf("Request time taken: %v seconds\n", time.Since(startTime).Seconds())
				if err != nil {
					return fmt.Errorf("NewReader: %w", err)
				}
				defer r.Close()

				buf := &bytes.Buffer{}
				if _, err := io.Copy(buf, r); err != nil {
					return fmt.Errorf("io.Copy: %w", err)
				}
				if !bytes.Equal(buf.Bytes(), objBytes) {
					return fmt.Errorf("content does not match, got len %v, want len %v", buf.Len(), len(objBytes))
				}
			}
			return nil
		})
	}

	// Wait for all worker goroutines to finish and check for errors
	if err := g.Wait(); err != nil {
		fmt.Printf("Error in worker: %v\n", err)
	}
}
