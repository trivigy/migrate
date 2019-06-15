package cmd

import (
	"testing"

	"github.com/blang/semver"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gopkg.in/yaml.v2"

	"github.com/trivigy/migrate/internal/dto"
	"github.com/trivigy/migrate/internal/store"
)

type UpCommandSuite struct {
	suite.Suite
}

func (r *UpCommandSuite) SetupTest() {
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

func (r *UpCommandSuite) TearDownTest() {
	ReInitialize()
	err := db.Close()
	assert.Nil(r.T(), err)
}

func (r *UpCommandSuite) TestSingleMigrationAppliedWithDryRun() {
	err := Append(dto.Migration{
		Tag: semver.Version{Major: 0, Minor: 0, Patch: 1},
		Up: []dto.Operation{
			{Query: `CREATE TABLE unittest (value text)`},
		},
		Down: []dto.Operation{
			{Query: `DROP TABLE unittest`},
		},
	})
	assert.Nil(r.T(), err)

	output, err := executeCommand(&root.Command, "up", "--env", "testing", "--dry-run")
	assert.Nil(r.T(), err)

	// @formatter:off
	expected :=
		"==> migration \"0.0.1\" (up)\n" +
		"CREATE TABLE unittest (value text);\n"
	// @formatter:on
	assert.Equal(r.T(), expected, output)

}

func (r *UpCommandSuite) TestSingleMigrationApplied() {
	err := Append(dto.Migration{
		Tag: semver.Version{Major: 0, Minor: 0, Patch: 1},
		Up: []dto.Operation{
			{Query: `CREATE TABLE unittests (value text)`},
		},
		Down: []dto.Operation{
			{Query: `DROP TABLE unittests`},
		},
	})
	assert.Nil(r.T(), err)

	output, err := executeCommand(&root.Command, "up", "--env", "testing")
	assert.Nil(r.T(), err)
	assert.Equal(r.T(), "migration \"0.0.1\" successfully applied (up)\n", output)

	_, err = db.Unittests.GetUnittests()
	assert.Nil(r.T(), err)
}

func (r *UpCommandSuite) TestSingleMigrationAppliedWithMultipleQueries() {
	err := Append(dto.Migration{
		Tag: semver.Version{Major: 0, Minor: 0, Patch: 1},
		Up: []dto.Operation{
			{Query: `CREATE TABLE unittests (value text)`},
			{Query: `INSERT INTO unittests(value) VALUES ("hello"), ("world")`},
		},
		Down: []dto.Operation{
			{Query: `DELETE FROM unittests WHERE value in ("hello", "world")`},
			{Query: `DROP TABLE unittests`},
		},
	})
	assert.Nil(r.T(), err)

	output, err := executeCommand(&root.Command, "up", "--env", "testing")
	assert.Nil(r.T(), err)
	assert.Equal(r.T(), "migration \"0.0.1\" successfully applied (up)\n", output)

	unittests, err := db.Unittests.GetUnittests()
	assert.Nil(r.T(), err)
	assert.Equal(r.T(), 2, len(unittests))
}

func (r *UpCommandSuite) TestMultipleMigrationApplied() {
	err := Append(dto.Migration{
		Tag: semver.Version{Major: 0, Minor: 0, Patch: 1},
		Up: []dto.Operation{
			{Query: `CREATE TABLE unittests (value text)`},
		},
		Down: []dto.Operation{
			{Query: `DROP TABLE unittests`},
		},
	})
	assert.Nil(r.T(), err)

	err = Append(dto.Migration{
		Tag: semver.Version{Major: 0, Minor: 0, Patch: 2},
		Up: []dto.Operation{
			{Query: `INSERT INTO unittests(value) VALUES ("hello"), ("world")`},
		},
		Down: []dto.Operation{
			{Query: `DELETE FROM unittests WHERE value in ("hello", "world")`},
		},
	})
	assert.Nil(r.T(), err)

	output, err := executeCommand(&root.Command, "up", "--env", "testing")
	assert.Nil(r.T(), err)

	// @formatter:off
	expected :=
		"migration \"0.0.1\" successfully applied (up)\n" +
		"migration \"0.0.2\" successfully applied (up)\n"
	// @formatter:on
	assert.Equal(r.T(), expected, output)

	unittests, err := db.Unittests.GetUnittests()
	assert.Nil(r.T(), err)
	assert.Equal(r.T(), 2, len(unittests))
}

func (r *UpCommandSuite) TestMultipleMigrationAppliedWithSingleStep() {
	err := Append(dto.Migration{
		Tag: semver.Version{Major: 0, Minor: 0, Patch: 1},
		Up: []dto.Operation{
			{Query: `CREATE TABLE unittests (value text)`},
		},
		Down: []dto.Operation{
			{Query: `DROP TABLE unittests`},
		},
	})
	assert.Nil(r.T(), err)

	err = Append(dto.Migration{
		Tag: semver.Version{Major: 0, Minor: 0, Patch: 2},
		Up: []dto.Operation{
			{Query: `INSERT INTO unittests(value) VALUES ("hello"), ("world")`},
		},
		Down: []dto.Operation{
			{Query: `DELETE FROM unittests WHERE value in ("hello", "world")`},
		},
	})
	assert.Nil(r.T(), err)

	output, err := executeCommand(&root.Command, "up", "--env", "testing", "-n", "1")
	assert.Nil(r.T(), err)
	assert.Equal(r.T(), "migration \"0.0.1\" successfully applied (up)\n", output)

	unittests, err := db.Unittests.GetUnittests()
	assert.Nil(r.T(), err)
	assert.Equal(r.T(), 0, len(unittests))
}

func TestUpCommandSuite(t *testing.T) {
	suite.Run(t, new(UpCommandSuite))
}
