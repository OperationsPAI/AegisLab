package harbor

import (
	"context"
	"fmt"
	"sort"
	"time"

	"aegis/config"
	"aegis/consts"

	"github.com/goharbor/go-client/pkg/harbor"
	"github.com/goharbor/go-client/pkg/sdk/v2.0/client/artifact"
	"github.com/goharbor/go-client/pkg/sdk/v2.0/models"
)

type Gateway struct {
	namespace string
	clientSet *harbor.ClientSet
}

func NewGateway() *Gateway {
	namespace := config.GetString("harbor.namespace")
	return &Gateway{
		namespace: namespace,
		clientSet: newClientSet(),
	}
}

func (g *Gateway) GetLatestTag(image string) (string, error) {
	if g.clientSet == nil {
		return "", fmt.Errorf("harbor client is not initialized")
	}

	ctx, cancel := context.WithTimeout(context.Background(), consts.HarborTimeout*consts.HarborTimeUnit)
	defer cancel()

	response, err := g.clientSet.V2().Artifact.ListArtifacts(ctx, &artifact.ListArtifactsParams{
		ProjectName:    g.namespace,
		RepositoryName: image,
		Context:        ctx,
	})
	if err != nil {
		return "", fmt.Errorf("failed to list artifacts: %v", err)
	}
	if len(response.Payload) == 0 {
		return "", fmt.Errorf("no artifacts found for image %s", image)
	}

	var allTags []*models.Tag
	for _, item := range response.Payload {
		if item.Tags != nil {
			allTags = append(allTags, item.Tags...)
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

func (g *Gateway) CheckImageExists(repository, tag string) (bool, error) {
	if g.clientSet == nil {
		return false, fmt.Errorf("harbor client is not initialized")
	}

	ctx, cancel := context.WithTimeout(context.Background(), consts.HarborTimeout*consts.HarborTimeUnit)
	defer cancel()

	response, err := g.clientSet.V2().Artifact.ListArtifacts(ctx, &artifact.ListArtifactsParams{
		ProjectName:    g.namespace,
		RepositoryName: repository,
		Context:        ctx,
	})
	if err != nil || len(response.Payload) == 0 {
		return false, nil
	}
	if tag == "" || tag == consts.DefaultContainerTag {
		return true, nil
	}

	for _, item := range response.Payload {
		if item.Tags == nil {
			continue
		}
		for _, currentTag := range item.Tags {
			if currentTag.Name == tag {
				return true, nil
			}
		}
	}

	return false, nil
}

func newClientSet() *harbor.ClientSet {
	registry := config.GetString("harbor.registry")
	username := config.GetString("harbor.username")
	password := config.GetString("harbor.password")
	harborURL := fmt.Sprintf("http://%s", registry)

	clientSet, err := harbor.NewClientSet(&harbor.ClientSetConfig{
		URL:      harborURL,
		Username: username,
		Password: password,
		Insecure: true,
	})
	if err != nil {
		return nil
	}

	return clientSet
}
