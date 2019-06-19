package migrate

import (
	"database/sql"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type MigrateSuite struct {
	suite.Suite
	db *sql.DB
}

func (r *MigrateSuite) SetupTest() {
	config := DatabaseConfig{
		Driver: "sqlite3",
		Source: ":memory:",
	}
	SetConfigs(map[string]DatabaseConfig{
		"testing": config,
	})

	var err error
	r.db, err = sql.Open(config.Driver, config.Source)
	assert.Nil(r.T(), err)

	err = SetDB(r.db)
	assert.Nil(r.T(), err)

	err = EnsureConfigured()
	assert.Nil(r.T(), err)
}

func (r *MigrateSuite) TearDownTest() {
	Restart()
	err := r.db.Close()
	assert.Nil(r.T(), err)
}

func (r *MigrateSuite) TestExecuteWithArgs() {
	Append(Migration{
		Tag: "0.0.1",
		Up: []Operation{
			{Query: `CREATE TABLE unittests (value text)`},
		},
		Down: []Operation{
			{Query: `DROP TABLE unittests`},
		},
	})

	output, err := ExecuteWithArgs("up", "--env", "testing")
	assert.Nil(r.T(), err)
	assert.Equal(r.T(), "migration \"0.0.1\" successfully applied (up)\n", output)

	output, err = ExecuteWithArgs("down", "--env", "testing")
	assert.Nil(r.T(), err)
	assert.Equal(r.T(), "migration \"0.0.1\" successfully removed (down)\n", output)
}

func TestMigrateSuite(t *testing.T) {
	suite.Run(t, new(MigrateSuite))
}
