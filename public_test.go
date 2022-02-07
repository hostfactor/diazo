package diazo

import (
	"github.com/hostfactor/api/go/providerconfig"
	"github.com/hostfactor/diazo/pkg/containers"
	"github.com/hostfactor/diazo/pkg/version/remoteversion"
	"github.com/stretchr/testify/suite"
	"testing"
)

type PublicTestSuite struct {
	suite.Suite

	Service         *client
	ContainerClient *containers.MockClient
	ImageUrl        containers.DockerImageUrl
}

func (p *PublicTestSuite) BeforeTest(_, _ string) {
	p.ContainerClient = new(containers.MockClient)
	p.ImageUrl = "ghcr.io/hostfactor/minecraft-server"
	p.Service = &client{
		ContainerClient: p.ContainerClient,
		Config: &providerconfig.ProviderConfig{
			Image: string(p.ImageUrl),
		},
	}
}

func (p *PublicTestSuite) TestSyncVersions() {
	// -- Given
	//
	syncer := new(MockVersionSyncer)
	given := &providerconfig.VersionSyncSpec{
		Source: &providerconfig.RemoteVersionSource{
			ContainerRegistry: &providerconfig.ContainerRegistryVersionSource{
				MatchesTag: "1\\.*",
			},
		},
		NameTransformer: &providerconfig.NameTransformer{
			SubstringReplace: &providerconfig.SubstringReplace{
				Matches:     "1\\.17([\\.1-9]*)(-?.*)",
				Replacement: "1.16$1$2",
			},
		},
	}

	givenTags := []string{
		"1.19",
		"1.17.1-java50",
		"1.17.1",
		"1.17",
		"2.0-java50",
	}
	outputTags := []string{
		"1.19",
		"1.16.1-java50",
		"1.16.1",
		"1.16",
	}
	p.ContainerClient.On("GetRepositoryTags", p.ImageUrl).Return(givenTags, nil)
	for _, tag := range outputTags {
		ver := &remoteversion.RemoteVersion{Name: tag}
		syncer.On("SyncVersions", ver).Return(nil)
	}

	// -- When
	//
	err := p.Service.SyncVersions(syncer, given)

	// -- Then
	//
	if p.NoError(err) {
		syncer.AssertExpectations(p.T())
	}
}

func TestPublicTestSuite(t *testing.T) {
	suite.Run(t, new(PublicTestSuite))
}
