package release

import (
	"bytes"
	"strings"
	"testing"

	"github.com/Pallinder/go-randomdata"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/trivigy/migrate/v2/driver"
	"github.com/trivigy/migrate/v2/driver/provider"
)

type ReleaseSuite struct {
	suite.Suite
	driver driver.Cluster
}

func (r *ReleaseSuite) SetupSuite() {
	buffer := bytes.NewBuffer(nil)
	r.driver = provider.Kind{Name: strings.ToLower(randomdata.SillyName())}
	assert.Nil(r.T(), r.driver.Setup(buffer))
}

func (r *ReleaseSuite) TearDownSuite() {
	buffer := bytes.NewBuffer(nil)
	assert.Nil(r.T(), r.driver.TearDown(buffer))
}

func (r *ReleaseSuite) TestReleaseCommand() {
	// command := NewCluster(map[string]Config{"default": {}})
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

func TestReleaseSuite(t *testing.T) {
	suite.Run(t, new(ReleaseSuite))
}
