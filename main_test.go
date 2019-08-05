package migrate

import (
	"bytes"
	"fmt"
	"os"
	"testing"

	"github.com/Pallinder/go-randomdata"
	"github.com/stretchr/testify/suite"

	"github.com/trivigy/migrate/cluster"
	"github.com/trivigy/migrate/database"
	"github.com/trivigy/migrate/driver/docker"
)

type MigrateSuite struct {
	suite.Suite
	name string
}

func (r *MigrateSuite) SetupTest() {
	r.name = "migrate"
}

func (r *MigrateSuite) TestMigrate_ExecuteWithArgs() {
	command := NewMigrate(map[string]Config{
		"default": {
			Cluster: cluster.Config{
				// ProjectID:   "digicontract-248304",
				// Location:    "us-east4-b",
				// MachineType: "n1-standard-1",
				// ImageType:   "ubuntu",
			},
			Database: database.Config{
				Driver: docker.Postgres{
					Tag:    "9.6",
					Name:   randomdata.SillyName(),
					DBName: "unittest",
					User:   "postgres",
				},
			},
		},
	})

	out := bytes.NewBuffer(nil)
	if err := command.Execute(r.name, out, []string{"--help"}); err != nil {
		os.Exit(1)
	}
	fmt.Fprintf(out, "-----\n")

	if err := command.Execute(r.name, out, []string{"database", "--help"}); err != nil {
		os.Exit(1)
	}
	fmt.Fprintf(out, "-----\n")

	if err := command.Execute(r.name, out, []string{"database", "create", "--help"}); err != nil {
		os.Exit(1)
	}

	// Append(Migration{
	// 	Tag: "0.0.1",
	// 	Up: []Operation{
	// 		{Query: `CREATE TABLE unittests (value text)`},
	// 	},
	// 	Down: []Operation{
	// 		{Query: `DROP TABLE unittests`},
	// 	},
	// })
	//
	// output, err := ExecuteWithArgs("up", "--env", "testing")
	// assert.Nil(r.T(), err)
	// assert.Equal(r.T(), "migration \"0.0.1\" successfully applied (up)\n", output)
	//
	// output, err = ExecuteWithArgs("down", "--env", "testing")
	// assert.Nil(r.T(), err)
	// assert.Equal(r.T(), "migration \"0.0.1\" successfully removed (down)\n", output)
}

func TestMigrateSuite(t *testing.T) {
	suite.Run(t, new(MigrateSuite))
}
