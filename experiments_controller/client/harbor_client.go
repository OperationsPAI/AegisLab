package client

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"sync"
	"time"

	"github.com/LGU-SE-Internal/rcabench/config"
	"github.com/LGU-SE-Internal/rcabench/consts"
)

// 单例模式的 Harbor 客户端
var (
	harborClient *HarborClient
	harborOnce   sync.Once
)

type HarborClient struct {
	host     string
	project  string
	username string
	password string
	client   *http.Client
}

type HarborArtifact struct {
	Tags []HarborTag `json:"tags"`
}

type HarborTag struct {
	Name     string    `json:"name"`
	PushTime time.Time `json:"push_time"`
}

func GetHarborClient() *HarborClient {
	harborOnce.Do(func() {
		harborClient = &HarborClient{
			host:     config.GetString("harbor.host"),
			project:  config.GetString("harbor.project"),
			username: config.GetString("harbor.username"),
			password: config.GetString("harbor.password"),
			client: &http.Client{
				Timeout: consts.HarborTimeout * consts.HarborTimeUnit,
			},
		}
	})
	return harborClient
}

func (h *HarborClient) GetLatestTag(image string) (string, error) {
	url := fmt.Sprintf(consts.HarborURL, h.host, h.project, image)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %v", err)
	}

	if h.username != "" && h.password != "" {
		req.SetBasicAuth(h.username, h.password)
	}

	resp, err := h.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API returns an error status code: %s", resp.Status)
	}

	var artifacts []HarborArtifact
	if err := json.NewDecoder(resp.Body).Decode(&artifacts); err != nil {
		return "", fmt.Errorf("failed to parse JSON: %v", err)
	}

	var tags []HarborTag
	for _, artifact := range artifacts {
		tags = append(tags, artifact.Tags...)
	}

	if len(tags) == 0 {
		return "", fmt.Errorf("failed to find valid tags")
	}

	sort.Slice(tags, func(i, j int) bool {
		return tags[i].PushTime.After(tags[j].PushTime)
	})

	return tags[0].Name, nil
}
