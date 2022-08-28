package variable

import (
	"github.com/stretchr/testify/suite"
	"testing"
)

type StoreTestSuite struct {
	suite.Suite
}

func (s *StoreTestSuite) TestStructValue() {
	// -- Given
	//
	given := NewStore()
	given.AddEntries(&Entry{
		Key: "derp",
		Val: map[string]string{"dorp": "there"},
	})
	expected := "hi there"

	// -- When
	//
	actual := RenderString("hi {{derp.dorp}}", given)

	// -- Then
	//
	s.Equal(expected, actual)
}

func TestStoreTestSuite(t *testing.T) {
	suite.Run(t, new(StoreTestSuite))
}
