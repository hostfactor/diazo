package remoteversion

import (
	"fmt"
	"github.com/hostfactor/api/go/providerconfig"
	"github.com/hostfactor/diazo/pkg/containers"
	"regexp"
)

func NewContainerRegistrySource(registry string, client containers.Client) VersionFetcher {
	return &containerRegistry{
		ContainerClient: client,
		ImageUrl:        containers.DockerImageUrl(registry),
	}
}

type containerRegistry struct {
	ContainerClient containers.Client
	ImageUrl        containers.ImageURL
}

func (c *containerRegistry) FetchVersions(cr *providerconfig.RemoteVersionSource) ([]*RemoteVersion, error) {
	if cr.GetContainerRegistry() == nil {
		return []*RemoteVersion{}, nil
	}
	tags, err := c.ContainerClient.GetRepositoryTags(c.ImageUrl)
	if err != nil {
		return nil, err
	}

	if cr.ContainerRegistry.MatchesTag != "" {
		tagMatcher, err := regexp.Compile(cr.ContainerRegistry.MatchesTag)
		if err != nil {
			return nil, fmt.Errorf("invalid tag matcher for %s: %s", c.ImageUrl.String(), err.Error())
		}
		filtered := make([]string, 0, len(tags))
		for _, v := range tags {
			if tagMatcher.MatchString(v) {
				filtered = append(filtered, v)
			}
		}
		tags = filtered
	}

	out := make([]*RemoteVersion, 0, len(tags))
	for _, v := range tags {
		out = append(out, c.tagToRemoteVersion(v))
	}

	return out, nil
}

func (c *containerRegistry) tagToRemoteVersion(tag string) *RemoteVersion {
	return &RemoteVersion{Name: tag}
}
