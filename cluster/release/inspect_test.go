package release

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type InspectSuite struct {
	suite.Suite
	name string
}

func (r *InspectSuite) SetupTest() {
	r.name = "inspect"
}

func (r *InspectSuite) TestInspectCommand() {
}

func TestInspectSuite(t *testing.T) {
	suite.Run(t, new(InspectSuite))
}
