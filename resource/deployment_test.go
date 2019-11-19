package resource

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/trivigy/migrate/v2/internal/testutils"
	"github.com/trivigy/migrate/v2/resource/primitive"
	"github.com/trivigy/migrate/v2/types"
)

type DeploymentSuite struct {
	suite.Suite
}

func (r *DeploymentSuite) TestDeploymentCommand() {
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
			Deployment{
				"create":  primitive.Create{Driver: &testutils.Driver{}},
				"destroy": primitive.Destroy{Driver: &testutils.Driver{}},
			},
			bytes.NewBuffer(nil),
			[]string{"create"},
			"",
		},
		{
			true,
			"accepts 1 arg(s), received 0 for \"deployment\"\n" +
				"\n" +
				"Usage:\n" +
				"  deployment COMMAND [flags]\n" +
				"\n" +
				"Available Commands:\n" +
				"  create      Constructs and starts a new instance of this resource.\n" +
				"  destroy     Stops and removes running instance of this resource.\n" +
				"\n" +
				"Flags:\n" +
				"      --help   Show help information.\n",
			Deployment{
				"create":  primitive.Create{Driver: &testutils.Driver{}},
				"destroy": primitive.Destroy{Driver: &testutils.Driver{}},
			},
			bytes.NewBuffer(nil),
			[]string{},
			"",
		},
	}

	for i, testCase := range testCases {
		failMsg := fmt.Sprintf("testCase: %d %v", i, testCase)
		runner := func() {
			err := testCase.cmd.Execute("deployment", testCase.buffer, testCase.args)
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

func TestDeploymentSuite(t *testing.T) {
	suite.Run(t, new(DeploymentSuite))
}
