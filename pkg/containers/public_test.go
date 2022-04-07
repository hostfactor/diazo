package containers

import (
	"github.com/stretchr/testify/suite"
	"regexp"
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
	given, _ := ParseImageURL("lloesche/valheim-server")
	opts := GetRepositoryTagsOpts{}

	// -- When
	//
	tags, err := p.Service.GetRepositoryTags(given, opts)

	// -- Then
	//
	if p.NoError(err) {
		p.Greater(len(tags), 0)
	}
}

func (p *PublicTestSuite) TestGetRepositoryTagsSingularName() {
	// -- Given
	//
	given, _ := ParseImageURL("nginx")
	opts := GetRepositoryTagsOpts{Regexp: regexp.MustCompile("^1.21.6$")}

	// -- When
	//
	tags, err := p.Service.GetRepositoryTags(given, opts)

	// -- Then
	//
	if p.NoError(err) {
		p.Len(tags, 1)
	}
}

func (p *PublicTestSuite) TestGetRepositoryTagsDoesntExist() {
	// -- Given
	//
	given, _ := ParseImageURL("qwerasdzxc")
	opts := GetRepositoryTagsOpts{}

	// -- When
	//
	tags, err := p.Service.GetRepositoryTags(given, opts)

	// -- Then
	//
	p.Error(err)
	p.Empty(tags)
}

func TestPublicTestSuite(t *testing.T) {
	suite.Run(t, new(PublicTestSuite))
}
