package cluster

import (
	"bytes"
	"strings"
	"testing"

	"github.com/Pallinder/go-randomdata"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/trivigy/migrate/v2/config"
	"github.com/trivigy/migrate/v2/driver/provider"
)

type DestroySuite struct {
	suite.Suite
	name string
}

func (r *DestroySuite) SetupTest() {
	r.name = "destroy"
}

func (r *DestroySuite) TestDestroyCommand() {
	cfg := map[string]config.Cluster{
		"default": {
			Driver: provider.Kind{
				Name: strings.ToLower(randomdata.SillyName()),
			},
		},
	}

	buffer := bytes.NewBuffer(nil)
	assert.Nil(r.T(), NewCreate(cfg).Execute(r.name, buffer, []string{}))
	assert.Nil(r.T(), NewDestroy(cfg).Execute(r.name, buffer, []string{}))
	assert.NotEqual(r.T(), 0, buffer.Len())
}

func TestDestroySuite(t *testing.T) {
	suite.Run(t, new(DestroySuite))
}
