package client

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/goharbor/go-client/pkg/harbor"
	"github.com/goharbor/go-client/pkg/sdk/v2.0/client/artifact"
	"github.com/goharbor/go-client/pkg/sdk/v2.0/models"

	"rcabench/config"
	"rcabench/consts"
)

// Singleton pattern Harbor client
var (
	harborClient *HarborClient
	harborOnce   sync.Once
)

type HarborClient struct {
	host      string
	project   string
	username  string
	password  string
	clientSet *harbor.ClientSet
}

func GetHarborClient() *HarborClient {
	harborOnce.Do(func() {
		host := config.GetString("harbor.host")
		project := config.GetString("harbor.project")
		username := config.GetString("harbor.username")
		password := config.GetString("harbor.password")

		// Build complete Harbor URL
		harborURL := fmt.Sprintf("http://%s", host)

		clientSet, err := harbor.NewClientSet(&harbor.ClientSetConfig{
			URL:      harborURL,
			Username: username,
			Password: password,
			Insecure: true, // Adjust as needed
		})
		if err != nil {
			// If client creation fails, log error but continue using nil client
			// Will return error in actual methods
			harborClient = &HarborClient{
				host:      host,
				project:   project,
				username:  username,
				password:  password,
				clientSet: nil,
			}
			return
		}

		harborClient = &HarborClient{
			host:      host,
			project:   project,
			username:  username,
			password:  password,
			clientSet: clientSet,
		}
	})
	return harborClient
}

func (h *HarborClient) GetLatestTag(image string) (string, error) {
	if h.clientSet == nil {
		return "", fmt.Errorf("harbor client is not initialized")
	}

	ctx, cancel := context.WithTimeout(context.Background(), consts.HarborTimeout*consts.HarborTimeUnit)
	defer cancel()

	params := &artifact.ListArtifactsParams{
		ProjectName:    h.project,
		RepositoryName: image,
		Context:        ctx,
	}

	response, err := h.clientSet.V2().Artifact.ListArtifacts(ctx, params)
	if err != nil {
		return "", fmt.Errorf("failed to list artifacts: %v", err)
	}

	if len(response.Payload) == 0 {
		return "", fmt.Errorf("no artifacts found for image %s", image)
	}

	var allTags []*models.Tag
	for _, artifact := range response.Payload {
		if artifact.Tags != nil {
			allTags = append(allTags, artifact.Tags...)
		}
	}

	if len(allTags) == 0 {
		return "", fmt.Errorf("no tags found for image %s", image)
	}

	sort.Slice(allTags, func(i, j int) bool {
		return time.Time(allTags[i].PushTime).After(time.Time(allTags[j].PushTime))
	})

	return allTags[0].Name, nil
}

func (h *HarborClient) CheckImageExists(image, tag string) (bool, error) {
	if h.clientSet == nil {
		return false, fmt.Errorf("harbor client is not initialized")
	}

	ctx, cancel := context.WithTimeout(context.Background(), consts.HarborTimeout*consts.HarborTimeUnit)
	defer cancel()

	params := &artifact.ListArtifactsParams{
		ProjectName:    h.project,
		RepositoryName: image,
		Context:        ctx,
	}

	response, err := h.clientSet.V2().Artifact.ListArtifacts(ctx, params)
	if err != nil {
		return false, nil
	}

	if len(response.Payload) == 0 {
		return false, nil
	}

	if tag == "" || tag == "latest" {
		return true, nil
	}

	for _, artifact := range response.Payload {
		if artifact.Tags != nil {
			for _, t := range artifact.Tags {
				if t.Name == tag {
					return true, nil
				}
			}
		}
	}

	return false, nil
}
