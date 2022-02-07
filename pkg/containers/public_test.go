package containers

import (
	"github.com/stretchr/testify/suite"
	"testing"
)

type PublicTestSuite struct {
	suite.Suite

	Service *client
}

func (p *PublicTestSuite) BeforeTest(_, _ string) {
	cl, err := NewClient()
	if err != nil {
		p.FailNow(err.Error())
	}

	p.Service = cl.(*client)
}

func (p *PublicTestSuite) TestGetRepositoryTags() {
	// -- Given
	//
	given := DockerImageUrl("lloesche/valheim-server")

	// -- When
	//
	tags, err := p.Service.GetRepositoryTags(given)

	// -- Then
	//
	if p.NoError(err) {
		p.Greater(len(tags), 0)
	}
}

func (p *PublicTestSuite) TestGetRepositoryTagsDoesntExist() {
	// -- Given
	//
	given := DockerImageUrl("qwerasdzxc")

	// -- When
	//
	tags, err := p.Service.GetRepositoryTags(given)

	// -- Then
	//
	p.Error(err)
	p.Empty(tags)
}

func TestPublicTestSuite(t *testing.T) {
	suite.Run(t, new(PublicTestSuite))
}
