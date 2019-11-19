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

type DestroySuite struct {
	suite.Suite
}

func (r *DestroySuite) TestDestroyCommand() {
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
			Destroy{Driver: &testutils.Driver{}},
			bytes.NewBuffer(nil),
			[]string{},
			"",
		},
		{
			false, "",
			Destroy{Driver: &testutils.Driver{}},
			bytes.NewBuffer(nil),
			[]string{"--help"},
			"Stops and removes running instance of this resource\n" +
				"\n" +
				"Usage:\n  destroy [flags]\n" +
				"\n" +
				"Flags:\n" +
				"  -t, --try    Simulates and prints resource execution parameters.\n" +
				"      --help   Show help information.\n",
		},
	}

	for i, testCase := range testCases {
		failMsg := fmt.Sprintf("testCase: %d %v", i, testCase)
		runner := func() {
			err := testCase.cmd.Execute("destroy", testCase.buffer, testCase.args)
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

func TestDestroySuite(t *testing.T) {
	suite.Run(t, new(DestroySuite))
}
