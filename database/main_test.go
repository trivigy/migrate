package database

import (
	"bytes"
	"os"
	"testing"

	"github.com/Pallinder/go-randomdata"
	"github.com/stretchr/testify/suite"

	"github.com/trivigy/migrate/v2/config"
	"github.com/trivigy/migrate/v2/driver/docker"
)

type DatabaseSuite struct {
	suite.Suite
	name string
}

func (r *DatabaseSuite) SetupTest() {
	r.name = "database"
}

func (r *DatabaseSuite) TestDatabaseCommand() {
	command := NewDatabase(map[string]config.Database{
		"default": {
			Driver: docker.Postgres{
				RefName: randomdata.SillyName(),
				Version: "9.6",
				DBName:  "unittest",
				User:    "postgres",
			},
		},
	})

	buffer := bytes.NewBuffer(nil)
	if err := command.Execute(r.name, buffer, []string{"--help"}); err != nil {
		os.Exit(1)
	}
	if err := command.Execute(r.name, buffer, []string{"create", "--help"}); err != nil {
		os.Exit(1)
	}

	// err := Append(dto.Migration{
	// 	Tag: semver.Version{Major: 0, Minor: 0, Patch: 1},
	// 	Up: []dto.Operation{
	// 		{Query: `CREATE TABLE unittest1 (id int)`},
	// 	},
	// 	Down: []dto.Operation{
	// 		{Query: `DROP TABLE unittest1`},
	// 	},
	// })
	// assert.Nil(r.T(), err)
	//
	// // @formatter:off
	// expected :=
	// 	"+-------+---------+\n" +
	// 		"|  TAG  | APPLIED |\n" +
	// 		"+-------+---------+\n" +
	// 		"| 0.0.1 | pending |\n" +
	// 		"+-------+---------+\n"
	// // @formatter:on
	//
	// output, err := ExecuteWithArgs("status", "--env", "testing")
	// assert.Equal(r.T(), expected, output)
	// assert.Nil(r.T(), err)
}

func TestDatabaseSuite(t *testing.T) {
	suite.Run(t, new(DatabaseSuite))
}
