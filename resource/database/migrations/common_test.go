package migrations

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/Pallinder/go-randomdata"
	"github.com/blang/semver"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/trivigy/migrate/v2/drivers/docker"
	"github.com/trivigy/migrate/v2/types"
)

type MigrationsSuite struct {
	suite.Suite
	Migrations *types.Migrations `json:"migrations" yaml:"migrations"`
	Driver     interface {
		types.Creator
		types.Destroyer
		types.Sourcer
	} `json:"driver" yaml:"driver"`
}

func (r *MigrationsSuite) SetupSuite() {
	r.Migrations = &types.Migrations{
		{
			Name: "create-unittest-table",
			Tag:  semver.Version{Major: 0, Minor: 0, Patch: 1},
			Up: []types.Operation{
				{Query: `CREATE TABLE unittests (value text)`},
			},
			Down: []types.Operation{
				{Query: `DROP TABLE unittests`},
			},
		},
		{
			Name: "seed-dummy-data",
			Tag:  semver.Version{Major: 0, Minor: 0, Patch: 2},
			Up: []types.Operation{
				{Query: `INSERT INTO unittests(value) VALUES ('hello'), ('world')`},
			},
			Down: []types.Operation{
				{Query: `DELETE FROM unittests WHERE value in ('hello', 'world')`},
			},
		},
		{
			Name: "seed-more-dummy-data",
			Tag:  semver.Version{Major: 0, Minor: 0, Patch: 3},
			Up: []types.Operation{
				{Query: `INSERT INTO unittests(value) VALUES ('here'), ('there')`},
			},
			Down: []types.Operation{
				{Query: `DELETE FROM unittests WHERE value in ('here', 'there')`},
			},
		},
	}

	r.Driver = &docker.Postgres{
		Name:    strings.ToLower(randomdata.SillyName()),
		Version: "9.6",
		DBName:  "unittest",
		User:    "postgres",
	}

	buffer := bytes.NewBuffer(nil)
	assert.Nil(r.T(), r.Driver.Create(context.Background(), buffer))
}

func (r *MigrationsSuite) TearDownSuite() {
	buffer := bytes.NewBuffer(nil)
	assert.Nil(r.T(), r.Driver.Destroy(context.Background(), buffer))
}

func (r *MigrationsSuite) TearDownTest() {
	buffer := bytes.NewBuffer(nil)
	down := Down{Migrations: r.Migrations, Driver: r.Driver}
	assert.Nil(r.T(), down.Execute("down", buffer, []string{"-l", "0"}))
}

func TestMigrationsSuite(t *testing.T) {
	suite.Run(t, new(MigrationsSuite))
}
