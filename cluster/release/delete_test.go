package release

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type DeleteSuite struct {
	suite.Suite
	name string
}

func (r *DeleteSuite) SetupTest() {
	r.name = "delete"
}

func (r *DeleteSuite) TestDeleteCommand() {
}

func TestDeleteSuite(t *testing.T) {
	suite.Run(t, new(DeleteSuite))
}
