package versiontransformer

import (
	"github.com/hostfactor/api/go/providerconfig"
	"github.com/hostfactor/diazo/pkg/version/remoteversion"
	"github.com/stretchr/testify/suite"
	"testing"
)

type SubstringReplaceTestSuite struct {
	suite.Suite
}

func (s *SubstringReplaceTestSuite) TestTransformRemoteVersion() {
	// -- Given
	//
	substr := &providerconfig.SubstringReplace{
		Matches:     "derp",
		Replacement: "dorp",
	}

	service, err := NewSubstringReplace(substr)
	if !s.NoError(err) {
		s.FailNow(err.Error())
	}

	given := &remoteversion.RemoteVersion{Name: "derp1"}
	expected := &remoteversion.RemoteVersion{Name: "dorp1"}

	// -- When
	//
	actual := service.TransformRemoteVersion(given)

	// -- Then
	//
	s.Equal(expected, actual)
}

func TestSubstringReplaceTestSuite(t *testing.T) {
	suite.Run(t, new(SubstringReplaceTestSuite))
}
