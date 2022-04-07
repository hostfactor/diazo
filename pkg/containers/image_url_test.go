package containers

import (
	"github.com/stretchr/testify/suite"
	"testing"
)

type ImageUrlTestSuite struct {
	suite.Suite
}

func (t *ImageUrlTestSuite) TestParseImage() {
	// -- Given
	//
	type test struct {
		Given          string
		ExpectedString string
		ExpectedRepo   string
		ExpectedName   string
		ExpectedPath   string
		ExpectedTag    string
		ExpectedProto  string
		ExpectedHost   string
		ExpectedError  string
	}

	tests := []test{
		{
			Given:          "nginx",
			ExpectedName:   "nginx",
			ExpectedPath:   "nginx:latest",
			ExpectedTag:    "latest",
			ExpectedProto:  "docker",
			ExpectedString: "docker://docker.io/nginx:latest",
			ExpectedHost:   "docker.io",
		},
		{
			Given:          "factoriotools/factorio",
			ExpectedName:   "factorio",
			ExpectedRepo:   "factoriotools",
			ExpectedPath:   "factoriotools/factorio:latest",
			ExpectedTag:    "latest",
			ExpectedProto:  "docker",
			ExpectedString: "docker://docker.io/factoriotools/factorio:latest",
			ExpectedHost:   "docker.io",
		},
		{
			Given:          "gcr.io/hostfactor/factorio:1.1",
			ExpectedName:   "factorio",
			ExpectedRepo:   "hostfactor",
			ExpectedPath:   "hostfactor/factorio:1.1",
			ExpectedTag:    "1.1",
			ExpectedProto:  "docker",
			ExpectedString: "docker://gcr.io/hostfactor/factorio:1.1",
			ExpectedHost:   "gcr.io",
		},
		{
			Given:          "gcr.io/factorio:1.1",
			ExpectedName:   "factorio",
			ExpectedPath:   "factorio:1.1",
			ExpectedTag:    "1.1",
			ExpectedProto:  "docker",
			ExpectedString: "docker://gcr.io/factorio:1.1",
			ExpectedHost:   "gcr.io",
		},
		{
			ExpectedPath:   ":latest",
			ExpectedTag:    "latest",
			ExpectedProto:  "docker",
			ExpectedString: "docker://docker.io/:latest",
			ExpectedHost:   "docker.io",
		},
		{
			Given:          "docker://docker.io/nginx",
			ExpectedPath:   "nginx:latest",
			ExpectedName:   "nginx",
			ExpectedTag:    "latest",
			ExpectedProto:  "docker",
			ExpectedString: "docker://docker.io/nginx:latest",
			ExpectedHost:   "docker.io",
		},
	}

	for i, test := range tests {
		actual, err := ParseImageURL(test.Given)
		if test.ExpectedError != "" {
			t.EqualError(err, test.ExpectedError, "test %d", i)
		}
		t.Equal(test.ExpectedRepo, actual.Repo, "test %d", i)
		t.Equal(test.ExpectedName, actual.Name, "test %d", i)
		t.Equal(test.ExpectedPath, actual.Path(), "test %d", i)
		t.Equal(test.ExpectedTag, actual.Tag, "test %d", i)
		t.Equal(test.ExpectedProto, actual.Proto, "test %d", i)
		t.Equal(test.ExpectedHost, actual.Host, "test %d", i)
		t.Equal(test.ExpectedString, actual.String(), "test %d", i)
	}
}

func TestImageUrlTestSuite(t *testing.T) {
	suite.Run(t, new(ImageUrlTestSuite))
}
