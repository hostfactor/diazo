package except

import (
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"testing"
)

type ExceptTestSuite struct {
	suite.Suite
}

func (e *ExceptTestSuite) TestToGRPC() {
	err := ToGRPC(ErrInvalid)
	st, _ := status.FromError(err)
	e.Equal(codes.InvalidArgument, st.Code())
}

func TestExceptTestSuite(t *testing.T) {
	suite.Run(t, new(ExceptTestSuite))
}
