package release

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type HistorySuite struct {
	suite.Suite
	name string
}

func (r *HistorySuite) SetupTest() {
	r.name = "history"
}

func (r *HistorySuite) TestHistoryCommand() {
}

func TestHistorySuite(t *testing.T) {
	suite.Run(t, new(HistorySuite))
}
