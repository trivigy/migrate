package release

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type ListSuite struct {
	suite.Suite
	name string
}

func (r *ListSuite) SetupTest() {
	r.name = "list"
}

func (r *ListSuite) TestListCommand() {
}

func TestListSuite(t *testing.T) {
	suite.Run(t, new(ListSuite))
}
