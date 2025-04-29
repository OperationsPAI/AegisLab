package client

import (
	"log"
	"os"
	"testing"
)

func TestHelm(t *testing.T) {
	// Create a new Helm client with a specific namespace
	namespace := os.Getenv("TS_NS")
	if namespace == "" {
		namespace = "train-ticket"
	}

	client, err := NewHelmClient(namespace)
	if err != nil {
		log.Fatalf("Error creating Helm client: %v", err)
	}

	// Add Train Ticket repository
	if err := client.AddRepo("train-ticket", "https://cuhk-se-group.github.io/train-ticket"); err != nil {
		log.Fatalf("Error adding repository: %v", err)
	}

	// Update repositories
	if err := client.UpdateRepo(); err != nil {
		log.Fatalf("Error updating repositories: %v", err)
	}

	// Install the Train Ticket chart
	port := os.Getenv("PORT")
	if port == "" {
		port = "30080"
	}

	if err := client.InstallTrainTicket(namespace, "637600ea", port); err != nil {
		log.Fatalf("Error installing Train Ticket: %v", err)
	}

	log.Printf("Train Ticket installed successfully")
}
