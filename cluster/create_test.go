package cluster

import (
	"bytes"
	"testing"

	"github.com/Pallinder/go-randomdata"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/trivigy/migrate/driver/provider"
)

type CreateSuite struct {
	suite.Suite
	name string
}

func (r *CreateSuite) SetupTest() {
	r.name = "create"
}

func (r *CreateSuite) TestCreate() {
	config := map[string]Config{
		"default": {
			Driver: provider.Kind{
				Name: randomdata.SillyName(),
			},
		},
	}

	buffer := bytes.NewBuffer(nil)
	assert.Nil(r.T(), NewCreate(config).Execute(r.name, buffer, []string{}))
	assert.Nil(r.T(), NewDestroy(config).Execute(r.name, buffer, []string{}))
	assert.NotEqual(r.T(), 0, buffer.Len())
}

func TestCreateSuite(t *testing.T) {
	suite.Run(t, new(CreateSuite))
}
