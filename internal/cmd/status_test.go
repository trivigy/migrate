package cmd

import (
	"strings"
	"testing"
	"time"

	"github.com/blang/semver"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gopkg.in/yaml.v2"

	"github.com/trivigy/migrate/internal/dao"
	"github.com/trivigy/migrate/internal/dto"
	"github.com/trivigy/migrate/internal/store"
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

	err = SetConfigs(rbytes)
	assert.Nil(r.T(), err)

	db, err = store.Open("sqlite3", ":memory:")
	assert.Nil(r.T(), err)
	err = db.Migrations.CreateTableIfNotExists()
	assert.Nil(r.T(), err)
}

func (r *StatusCommandSuite) TearDownTest() {
	Restart()
	err := db.Close()
	assert.Nil(r.T(), err)
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

	output, err := executeCommand(&root.Command, "status", "--env", "testing")
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

	output, err := executeCommand(&root.Command, "status", "--env", "testing")
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

	err = db.Migrations.Insert(&dao.Migration{Tag: "0.0.1", Timestamp: time.Now()})
	assert.Nil(r.T(), err)

	output, err := executeCommand(&root.Command, "status", "--env", "testing")
	assert.Equal(r.T(), 1, strings.Count(output, "pending"))
	assert.Nil(r.T(), err)
}

func TestStatusCommandSuite(t *testing.T) {
	suite.Run(t, new(StatusCommandSuite))
}
