package cluster

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type ClusterSuite struct {
	suite.Suite
	name string
}

func (r *ClusterSuite) SetupTest() {
	r.name = "cluster"
}

func (r *ClusterSuite) TestClusterCommand() {
	// command := NewCluster(map[string]config.Cluster{"default": {}})
	//
	// buffer := bytes.NewBuffer(nil)
	// if err := command.Execute(r.name, buffer, []string{"--help"}); err != nil {
	// 	os.Exit(1)
	// }
	// if err := command.Execute(r.name, buffer, []string{"create", "--help"}); err != nil {
	// 	os.Exit(1)
	// }

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

func TestClusterSuite(t *testing.T) {
	suite.Run(t, new(ClusterSuite))
}