package variable

import (
	"bytes"
	jsoniter "github.com/json-iterator/go"
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

func (s *StoreTestSuite) TestStringInterface() {
	// -- Given
	//
	vals := `{"hello": {"nested": 1}}`
	out := map[string]interface{}{}
	dec := jsoniter.NewDecoder(bytes.NewBufferString(vals))
	dec.UseNumber()
	_ = dec.Decode(&out)
	given := NewStore()
	given.AddEntries(&Entry{
		Key: "derp",
		Val: out,
	})
	expected := "hi 1"

	// -- When
	//
	actual := RenderString("hi {{derp.hello.nested}}", given)

	// -- Then
	//
	s.Equal(expected, actual)
}

func TestStoreTestSuite(t *testing.T) {
	suite.Run(t, new(StoreTestSuite))
}
