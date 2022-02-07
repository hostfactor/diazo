package remoteversion

import (
	"github.com/hostfactor/api/go/providerconfig"
	"github.com/hostfactor/diazo/pkg/containers"
	"github.com/stretchr/testify/suite"
	"testing"
)

type ContainerRegistryTestSuite struct {
	suite.Suite

	Service *containerRegistry

	ContainerClient *containers.MockClient
	ImageUrl        containers.ImageURL
}

func (c *ContainerRegistryTestSuite) BeforeTest(_, _ string) {
	c.ContainerClient = new(containers.MockClient)
	c.ImageUrl = containers.DockerImageUrl("ghcr.io/hostfactor/minecraft-server")

	c.Service = &containerRegistry{
		ContainerClient: c.ContainerClient,
		ImageUrl:        c.ImageUrl,
	}
}

func (c *ContainerRegistryTestSuite) TestFetchVersions() {
	// -- Given
	//
	given := &providerconfig.RemoteVersionSource{
		ContainerRegistry: &providerconfig.ContainerRegistryVersionSource{
			MatchesTag: "1.+",
		},
	}

	givenTags := []string{
		"1.19",
		"1.17.1-java50",
		"1.17.1",
		"1.17",
		"2.0-java50",
	}
	expected := []*RemoteVersion{
		{Name: "1.19"},
		{Name: "1.17.1-java50"},
		{Name: "1.17.1"},
		{Name: "1.17"},
	}
	c.ContainerClient.On("GetRepositoryTags", c.ImageUrl).Return(givenTags, nil)

	// -- When
	//
	vers, err := c.Service.FetchVersions(given)

	// -- Then
	//
	if c.NoError(err) {
		c.ContainerClient.AssertExpectations(c.T())
		c.Equal(expected, vers)
	}
}

func (c *ContainerRegistryTestSuite) TestFetchVersionsNoSource() {
	// -- Given
	//
	given := &providerconfig.RemoteVersionSource{}

	// -- When
	//
	vers, err := c.Service.FetchVersions(given)

	// -- Then
	//
	if c.NoError(err) {
		c.ContainerClient.AssertExpectations(c.T())
		c.Equal([]*RemoteVersion{}, vers)
	}
}

func TestContainerRegistryTestSuite(t *testing.T) {
	suite.Run(t, new(ContainerRegistryTestSuite))
}
