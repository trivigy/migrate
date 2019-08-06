package release

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type UpgradeSuite struct {
	suite.Suite
	name string
}

func (r *UpgradeSuite) SetupTest() {
	r.name = "upgrade"
}

func (r *UpgradeSuite) TestUpgradeCommand() {
}

func TestUpgradeSuite(t *testing.T) {
	suite.Run(t, new(UpgradeSuite))
}
