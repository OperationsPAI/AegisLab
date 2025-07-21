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

	"github.com/LGU-SE-Internal/rcabench/config"
	"github.com/LGU-SE-Internal/rcabench/consts"
)

// 单例模式的 Harbor 客户端
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

		// 构建完整的 Harbor URL
		harborURL := fmt.Sprintf("http://%s", host)

		clientSet, err := harbor.NewClientSet(&harbor.ClientSetConfig{
			URL:      harborURL,
			Username: username,
			Password: password,
			Insecure: true, // 根据需要调整
		})
		if err != nil {
			// 如果客户端创建失败，记录错误但继续使用 nil 客户端
			// 在实际方法中会返回错误
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
