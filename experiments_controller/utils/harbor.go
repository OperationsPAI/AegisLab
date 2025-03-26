package utils

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"time"
)

const HarborURL = "http://%s/api/v2.0/projects/%s/repositories/%s/artifacts?page_size=100"

type HarborConfig struct {
	Host     string
	Project  string
	Repo     string
	Username string
	Password string
}

type HarborArtifact struct {
	Tags []HarborTag `json:"tags"`
}

type HarborTag struct {
	Name     string    `json:"name"`
	PushTime time.Time `json:"push_time"`
}

func GetLatestTag(config HarborConfig) (string, error) {
	url := fmt.Sprintf(HarborURL, config.Host, config.Project, config.Repo)

	client := &http.Client{}
	resp, err := client.Get(url)
	if err != nil {
		return "", fmt.Errorf("Failed to request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API returns an error status code: %s", resp.Status)
	}

	var artifacts []HarborArtifact
	if err := json.NewDecoder(resp.Body).Decode(&artifacts); err != nil {
		return "", fmt.Errorf("Failed to parse JSON: %v", err)
	}

	var tags []HarborTag
	for _, artifact := range artifacts {
		tags = append(tags, artifact.Tags...)
	}

	if len(tags) == 0 {
		return "", fmt.Errorf("Failed to find valid tags")
	}

	sort.Slice(tags, func(i, j int) bool {
		return tags[i].PushTime.After(tags[j].PushTime)
	})

	return tags[0].Name, nil
}
