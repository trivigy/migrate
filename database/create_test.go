package database

import (
	"bytes"
	"os"
	"testing"

	"github.com/Pallinder/go-randomdata"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/trivigy/migrate/v2/config"
	"github.com/trivigy/migrate/v2/driver/docker"
	"github.com/trivigy/migrate/v2/driver/provider"
)

type CreateSuite struct {
	suite.Suite
	name string
}

func (r *CreateSuite) SetupTest() {
	r.name = "create"
}

func (r *CreateSuite) TestCreateCommand() {
	cfg := map[string]config.Database{}
	if os.Getenv("CI") != "true" {
		cfg = map[string]config.Database{
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
		cfg = map[string]config.Database{
			"default": {
				Driver: provider.SQL{
					Dialect:    "postgres",
					DataSource: "host=localhost user=postgres dbname=unittest sslmode=disable",
				},
			},
		}
	}

	buffer := bytes.NewBuffer(nil)
	assert.Nil(r.T(), NewCreate(cfg).Execute(r.name, buffer, []string{}))
	assert.Nil(r.T(), NewDestroy(cfg).Execute(r.name, buffer, []string{}))
}

func TestCreateSuite(t *testing.T) {
	suite.Run(t, new(CreateSuite))
}
