package database

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/trivigy/migrate/internal/testutils"
)

type GenerateSuite struct {
	suite.Suite
	name string
}

func (r *GenerateSuite) SetupTest() {
	r.name = "generate"
}

func (r *GenerateSuite) TestGenerate() {
	defaultConfig := map[string]Config{"default": {}}

	dir, err := ioutil.TempDir(os.TempDir(), "migrate-")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(dir)

	testCases := []struct {
		shouldFail bool
		onFail     string
		config     map[string]Config
		buffer     *bytes.Buffer
		args       []string
	}{
		{
			false, "",
			defaultConfig,
			bytes.NewBuffer(nil),
			[]string{"example", "-d", dir},
		},
		{
			true, "Error: missing \"default\" environment configuration\n",
			map[string]Config{},
			bytes.NewBuffer(nil),
			[]string{"example"},
		},
		{
			true, "Error: directory \"./not-found\" not found\n",
			defaultConfig,
			bytes.NewBuffer(nil),
			[]string{"example", "-d", "./not-found"},
		},
		{
			true,
			"Error: accepts 1 arg(s), received 0 for \"generate\"\n" +
				"\n" +
				"Usage:  generate NAME[:TAG] [flags]\n",
			defaultConfig,
			bytes.NewBuffer(nil),
			[]string{},
		},
		{
			true,
			"Error: invalid argument \"name:wrong\" for \"generate\"\n" +
				"\n" +
				"Usage:  generate NAME[:TAG] [flags]\n",
			defaultConfig,
			bytes.NewBuffer(nil),
			[]string{"name:wrong"},
		},
		{
			true,
			"Error: invalid argument \"name:0.0.0-alpha.2+001\" for \"generate\"\n" +
				"\n" +
				"Usage:  generate NAME[:TAG] [flags]\n",
			defaultConfig,
			bytes.NewBuffer(nil),
			[]string{"name:0.0.0-alpha.2+001"},
		},
	}

	for i, testCase := range testCases {
		failMsg := fmt.Sprintf("testCase: %d %v", i, testCase)
		runner := func() {
			command := NewGenerate(testCase.config)
			err := command.Execute(r.name, testCase.buffer, testCase.args)
			if err != nil {
				panic(testCase.buffer.String())
			}

			empty, err := testutils.IsDirEmpty(dir)
			if err != nil {
				panic(err)
			}

			if empty {
				panic("not empty")
			}
		}

		if testCase.shouldFail {
			assert.PanicsWithValue(r.T(), testCase.onFail, runner, failMsg)
		} else {
			assert.NotPanics(r.T(), runner, failMsg)
		}
	}
}

func TestGenerateSuite(t *testing.T) {
	suite.Run(t, new(GenerateSuite))
}
