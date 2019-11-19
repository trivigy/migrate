package docker

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/Pallinder/go-randomdata"
	"github.com/davecgh/go-spew/spew"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/trivigy/migrate/v2/driver"
)

type PostgresSuite struct {
	suite.Suite
	Driver interface {
		driver.WithCreate
		driver.WithDestroy
		driver.WithSource
	}
}

func (r *PostgresSuite) SetupSuite() {
	r.Driver = &Postgres{
		Name:     strings.ToLower(randomdata.SillyName()),
		Version:  "9.6",
		User:     "postgres",
		Password: "postgres",
		DBName:   "default",
	}
}

func (r *PostgresSuite) TestPostgres_SetupTearDownCycle() {
	testCases := []struct {
		shouldFail bool
		onFail     string
		buffer     *bytes.Buffer
		output     string
	}{
		{
			false, "",
			bytes.NewBuffer(nil),
			"",
		},
	}

	for i, tc := range testCases {
		failMsg := fmt.Sprintf("test: %d %v", i, spew.Sprint(tc))
		runner := func() {
			err := r.Driver.Create(context.Background(), tc.buffer)
			if err != nil {
				panic(err.Error())
			}

			err = r.Driver.Destroy(context.Background(), tc.buffer)
			if err != nil {
				panic(err.Error())
			}
		}

		if tc.shouldFail {
			assert.PanicsWithValue(r.T(), tc.onFail, runner, failMsg)
		} else {
			assert.NotPanics(r.T(), runner, failMsg)
		}
	}
}

func (r *PostgresSuite) TestPostgres_MarshalUnmarshal() {
	testCases := []struct {
		shouldFail bool
		onFail     string
		driver     interface {
			driver.WithCreate
			driver.WithDestroy
			driver.WithSource
		}
	}{
		{
			false, "",
			&Postgres{
				Name:         "unittest",
				Version:      "unittest",
				User:         "unittest",
				Password:     "unittest",
				DBName:       "unittest",
				InitDBArgs:   "unittest",
				InitDBWalDir: "unittest",
				PGData:       "unittest",
			},
		},
	}

	for i, tc := range testCases {
		failMsg := fmt.Sprintf("test: %d %v", i, spew.Sprint(tc))
		runner := func() {
			rbytes, err := json.Marshal(tc.driver)
			if err != nil {
				panic(err.Error())
			}

			actual := &Postgres{}
			err = json.Unmarshal(rbytes, actual)
			if err != nil {
				panic(err.Error())
			}

			assert.EqualValues(r.T(), tc.driver, actual)
		}

		if tc.shouldFail {
			assert.PanicsWithValue(r.T(), tc.onFail, runner, failMsg)
		} else {
			assert.NotPanics(r.T(), runner, failMsg)
		}
	}
}

func TestPostgresSuite(t *testing.T) {
	suite.Run(t, new(PostgresSuite))
}
