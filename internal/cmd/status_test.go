package cmd

import (
	"strings"
	"testing"

	"github.com/blang/semver"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gopkg.in/yaml.v2"

	"github.com/trivigy/migrate/internal/dto"
)

type StatusCommandSuite struct {
	suite.Suite
}

func (r *StatusCommandSuite) SetupTest() {
	rbytes, err := yaml.Marshal(map[string]interface{}{
		"testing": map[string]interface{}{
			"driver": "sqlite3",
			"source": ":memory:",
		},
	})
	assert.Nil(r.T(), err)
	assert.Nil(r.T(), SetConfigs(rbytes))
}

func (r *StatusCommandSuite) TearDownTest() {
	_ = Clear()
	assert.Nil(r.T(), Close())
}

func (r *StatusCommandSuite) TestSinglePendingMigration() {
	err := Append(dto.Migration{
		Tag: semver.Version{Major: 0, Minor: 0, Patch: 1},
		Up: []dto.Operation{
			{Query: `CREATE TABLE unittest1 (id int)`},
		},
		Down: []dto.Operation{
			{Query: `DROP TABLE unittest1`},
		},
	})
	assert.Nil(r.T(), err)

	// @formatter:off
	expected :=
		"+-------+---------+\n" +
			"|  TAG  | APPLIED |\n" +
			"+-------+---------+\n" +
			"| 0.0.1 | pending |\n" +
			"+-------+---------+\n"
	// @formatter:on

	output, err := ExecuteWithArgs("status", "--env", "testing")
	assert.Equal(r.T(), expected, output)
	assert.Nil(r.T(), err)
}

func (r *StatusCommandSuite) TestMultiplePendingMigration() {
	err := Append(dto.Migration{
		Tag: semver.Version{Major: 0, Minor: 0, Patch: 1},
		Up: []dto.Operation{
			{Query: `CREATE TABLE unittest1 (id int)`},
		},
		Down: []dto.Operation{
			{Query: `DROP TABLE unittest1`},
		},
	})
	assert.Nil(r.T(), err)

	err = Append(dto.Migration{
		Tag: semver.Version{Major: 0, Minor: 0, Patch: 2},
		Up: []dto.Operation{
			{Query: `CREATE TABLE unittest2 (id int)`},
		},
		Down: []dto.Operation{
			{Query: `DROP TABLE unittest2`},
		},
	})
	assert.Nil(r.T(), err)

	// @formatter:off
	expected :=
		"+-------+---------+\n" +
			"|  TAG  | APPLIED |\n" +
			"+-------+---------+\n" +
			"| 0.0.1 | pending |\n" +
			"| 0.0.2 | pending |\n" +
			"+-------+---------+\n"
	// @formatter:on

	output, err := ExecuteWithArgs("status", "--env", "testing")
	assert.Equal(r.T(), expected, output)
	assert.Nil(r.T(), err)
}

func (r *StatusCommandSuite) TestTwoPendingOneAppliedMigration() {
	err := Append(dto.Migration{
		Tag: semver.Version{Major: 0, Minor: 0, Patch: 1},
		Up: []dto.Operation{
			{Query: `CREATE TABLE unittest1 (id int)`},
		},
		Down: []dto.Operation{
			{Query: `DROP TABLE unittest1`},
		},
	})
	assert.Nil(r.T(), err)

	err = Append(dto.Migration{
		Tag: semver.Version{Major: 0, Minor: 0, Patch: 2},
		Up: []dto.Operation{
			{Query: `CREATE TABLE unittest2 (id int)`},
		},
		Down: []dto.Operation{
			{Query: `DROP TABLE unittest2`},
		},
	})
	assert.Nil(r.T(), err)

	output, err := ExecuteWithArgs("up", "--env", "testing", "-n", "1")
	assert.Nil(r.T(), err)
	assert.Equal(r.T(), "migration \"0.0.1\" successfully applied (up)\n", output)

	output, err = ExecuteWithArgs("status", "--env", "testing")
	assert.Equal(r.T(), 1, strings.Count(output, "pending"))
	assert.Nil(r.T(), err)
}

func TestStatusCommandSuite(t *testing.T) {
	suite.Run(t, new(StatusCommandSuite))
}
