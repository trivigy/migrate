package cmd

import (
	"testing"

	"github.com/blang/semver"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gopkg.in/yaml.v2"

	"github.com/trivigy/migrate/internal/dto"
)

type DownCommandSuite struct {
	suite.Suite
}

func (r *DownCommandSuite) SetupTest() {
	rbytes, err := yaml.Marshal(map[string]interface{}{
		"testing": map[string]interface{}{
			"driver": "sqlite3",
			"source": ":memory:",
		},
	})
	assert.Nil(r.T(), err)
	assert.Nil(r.T(), SetConfigs(rbytes))
}

func (r *DownCommandSuite) TearDownTest() {
	_ = Clear()
	assert.Nil(r.T(), Close())
}

func (r *DownCommandSuite) TestSingleMigrationRemovedWithDryRun() {
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

	output, err := ExecuteWithArgs("up", "--env", "testing")
	assert.Nil(r.T(), err)
	assert.Equal(r.T(), "migration \"0.0.1\" successfully applied (up)\n", output)

	output, err = ExecuteWithArgs("down", "--env", "testing", "--dry-run")
	assert.Nil(r.T(), err)

	// @formatter:off
	expected :=
		"==> migration \"0.0.1\" (down)\n" +
			"DROP TABLE unittests;\n"
	// @formatter:on
	assert.Equal(r.T(), expected, output)

}

func (r *DownCommandSuite) TestSingleMigrationRemoved() {
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

	output, err := ExecuteWithArgs("up", "--env", "testing")
	assert.Nil(r.T(), err)
	assert.Equal(r.T(), "migration \"0.0.1\" successfully applied (up)\n", output)

	output, err = ExecuteWithArgs("down", "--env", "testing")
	assert.Nil(r.T(), err)
	assert.Equal(r.T(), "migration \"0.0.1\" successfully removed (down)\n", output)

	_, err = db.Unittests.GetUnittests()
	assert.NotNil(r.T(), err)
}

func (r *DownCommandSuite) TestSingleMigrationDeletedWithMultipleQueries() {
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

	output, err := ExecuteWithArgs("up", "--env", "testing")
	assert.Nil(r.T(), err)
	assert.Equal(r.T(), "migration \"0.0.1\" successfully applied (up)\n", output)

	output, err = ExecuteWithArgs("down", "--env", "testing")
	assert.Nil(r.T(), err)
	assert.Equal(r.T(), "migration \"0.0.1\" successfully removed (down)\n", output)

	_, err = db.Unittests.GetUnittests()
	assert.NotNil(r.T(), err)
}

func (r *UpCommandSuite) TestMultipleMigrationRemoved() {
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

	output, err := ExecuteWithArgs("up", "--env", "testing")
	assert.Nil(r.T(), err)
	// @formatter:off
	expectedApplied :=
		"migration \"0.0.1\" successfully applied (up)\n" +
			"migration \"0.0.2\" successfully applied (up)\n"
	// @formatter:on
	assert.Equal(r.T(), expectedApplied, output)

	output, err = ExecuteWithArgs("down", "--env", "testing")
	assert.Nil(r.T(), err)
	// @formatter:off
	expectedRemoved :=
		"migration \"0.0.2\" successfully removed (down)\n" +
			"migration \"0.0.1\" successfully removed (down)\n"
	// @formatter:on
	assert.Equal(r.T(), expectedRemoved, output)

	_, err = db.Unittests.GetUnittests()
	assert.NotNil(r.T(), err)
}

func (r *UpCommandSuite) TestMultipleMigrationDeletedWithSingleStep() {
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

	output, err := ExecuteWithArgs("up", "--env", "testing")
	assert.Nil(r.T(), err)
	// @formatter:off
	expectedApplied :=
		"migration \"0.0.1\" successfully applied (up)\n" +
			"migration \"0.0.2\" successfully applied (up)\n"
	// @formatter:on
	assert.Equal(r.T(), expectedApplied, output)

	output, err = ExecuteWithArgs("down", "--env", "testing", "-n", "1")
	assert.Nil(r.T(), err)
	assert.Equal(r.T(), "migration \"0.0.2\" successfully removed (down)\n", output)

	unittests, err := db.Unittests.GetUnittests()
	assert.Nil(r.T(), err)
	assert.Equal(r.T(), 0, len(unittests))

	output, err = ExecuteWithArgs("down", "--env", "testing", "-n", "1")
	assert.Nil(r.T(), err)
	assert.Equal(r.T(), "migration \"0.0.1\" successfully removed (down)\n", output)

	_, err = db.Unittests.GetUnittests()
	assert.NotNil(r.T(), err)
}

func TestDownCommandSuite(t *testing.T) {
	suite.Run(t, new(DownCommandSuite))
}
