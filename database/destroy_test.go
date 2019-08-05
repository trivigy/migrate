package database

import (
	"bytes"
	"os"
	"testing"

	"github.com/Pallinder/go-randomdata"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/trivigy/migrate/driver/docker"
	"github.com/trivigy/migrate/driver/provider"
)

type DestroySuite struct {
	suite.Suite
	name string
}

func (r *DestroySuite) SetupTest() {
	r.name = "destroy"
}

func (r *DestroySuite) TestDestroyCommand() {
	config := map[string]Config{}
	if os.Getenv("CI") != "true" {
		config = map[string]Config{
			"default": {
				Driver: docker.Postgres{
					RefName: randomdata.SillyName(),
					Version: "9.6",
					DBName:  "unittest",
					User:    "postgres",
				},
			},
		}
	} else {
		config = map[string]Config{
			"default": {
				Driver: provider.SQLDatabase{
					Dialect:    "postgres",
					DataSource: "host=localhost user=postgres dbname=unittest sslmode=disable",
				},
			},
		}
	}

	buffer := bytes.NewBuffer(nil)
	assert.Nil(r.T(), NewCreate(config).Execute(r.name, buffer, []string{}))
	assert.Nil(r.T(), NewDestroy(config).Execute(r.name, buffer, []string{}))
}

func TestDestroySuite(t *testing.T) {
	suite.Run(t, new(DestroySuite))
}
