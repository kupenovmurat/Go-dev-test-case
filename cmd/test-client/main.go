package main

import (
	"flag"
	"log"
	"os"

	"github.com/kupenovmurat/Go-dev-test-case/pkg/test"
)

var (
	restServer    = flag.String("rest", "http://localhost:8080", "REST server URL")
	storageServer = flag.String("storage", "http://localhost:8081", "Storage server URL (for registration)")
	testDataDir   = flag.String("data", "./test-data", "Directory for test data")
	fileSize      = flag.Int64("size", 10*1024*1024, "Size of the test file in bytes (default: 10MB)")
)

func main() {
	flag.Parse()

	if err := os.MkdirAll(*testDataDir, 0755); err != nil {
		log.Fatalf("Failed to create test data directory: %v", err)
	}

	testModule := test.NewTestModule(*restServer, *storageServer, *testDataDir)

	if err := testModule.RegisterStorageServer(*storageServer); err != nil {
		log.Printf("Warning: Failed to register storage server: %v", err)
	}

	if err := testModule.RunTest(*fileSize); err != nil {
		log.Fatalf("Test failed: %v", err)
	}

	log.Println("Test completed successfully!")
}
