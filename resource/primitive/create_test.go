package primitive

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/trivigy/migrate/v2/internal/testutils"
	"github.com/trivigy/migrate/v2/types"
)

type CreateSuite struct {
	suite.Suite
}

func (r *CreateSuite) TestCreateCommand() {
	testCases := []struct {
		shouldFail bool
		onFail     string
		cmd        types.Command
		buffer     *bytes.Buffer
		args       []string
		output     string
	}{
		{
			true, "implement me",
			Create{Driver: &testutils.Driver{}},
			bytes.NewBuffer(nil),
			[]string{},
			"",
		},
		{
			false, "",
			Create{Driver: &testutils.Driver{}},
			bytes.NewBuffer(nil),
			[]string{"--help"},
			"Constructs and starts a new instance of this resource\n" +
				"\n" +
				"Usage:\n" +
				"  create [flags]\n" +
				"\n" +
				"Flags:\n" +
				"  -m, --merge PATH   Merges specified json PATH with configured parameters.\n" +
				"      --dry-run      Simulate parameter merging without resource execution.\n" +
				"      --help         Show help information.\n",
		},
	}

	for i, testCase := range testCases {
		failMsg := fmt.Sprintf("testCase: %d %v", i, testCase)
		runner := func() {
			err := testCase.cmd.Execute("create", testCase.buffer, testCase.args)
			if err != nil {
				panic(err.Error())
			}

			if testCase.output != testCase.buffer.String() {
				fmt.Printf("%q\n", testCase.buffer.String())
				panic(testCase.buffer.String())
			}
		}

		if testCase.shouldFail {
			assert.PanicsWithValue(r.T(), testCase.onFail, runner, failMsg)
		} else {
			assert.NotPanics(r.T(), runner, failMsg)
		}
	}
}

func TestCreateSuite(t *testing.T) {
	suite.Run(t, new(CreateSuite))
}
