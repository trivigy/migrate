package release

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type InstallSuite struct {
	suite.Suite
	name string
}

func (r *InstallSuite) SetupTest() {
	r.name = "install"
}

func (r *InstallSuite) TestInstallCommand() {
}

func TestInstallSuite(t *testing.T) {
	suite.Run(t, new(InstallSuite))
}
