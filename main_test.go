package migrate

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type MigrateSuite struct {
	suite.Suite
}

func (r *MigrateSuite) SetupTest() {
	SetConfigs(map[string]DataSource{
		"testing": {
			Driver: "sqlite3",
			Source: ":memory:",
		},
	})
}

func (r *MigrateSuite) TearDownTest() {
	_ = Clear()
	assert.Nil(r.T(), Close())
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
