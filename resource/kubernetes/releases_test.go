package kubernetes

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/trivigy/migrate/v2/resource/kubernetes/releases"
	"github.com/trivigy/migrate/v2/types"
)

type ReleaseSuite struct {
	suite.Suite
}

func (r *ReleaseSuite) TestReleaseCommand() {
	testCases := []struct {
		shouldFail bool
		onFail     string
		cmd        types.Command
		buffer     *bytes.Buffer
		args       []string
		output     string
	}{
		{
			true,
			"accepts 1 arg(s), received 0 for \"releases\"\n" +
				"\n" +
				"Usage:\n" +
				"  releases COMMAND [flags]\n" +
				"\n" +
				"Available Commands:\n" +
				"  describe    Describe registered releases with states information.\n" +
				"  generate    Adds a new release template.\n" +
				"  install     Deploys release resources on running cluster.\n" +
				"  uninstall   Stops a running release and removes the resources.\n" +
				"  update      Redeploy a modified release and track revision version.\n" +
				"\n" +
				"Flags:\n" +
				"      --help   Show help information.\n",
			Releases{
				"generate":  releases.Generate{},
				"install":   releases.Install{},
				"uninstall": releases.Uninstall{},
				"update":    releases.Update{},
				"describe":  releases.Describe{},
			},
			bytes.NewBuffer(nil),
			[]string{},
			"",
		},
		{
			false, "",
			Releases{
				"generate":  releases.Generate{},
				"install":   releases.Install{},
				"uninstall": releases.Uninstall{},
				"update":    releases.Update{},
				"describe":  releases.Describe{},
			},
			bytes.NewBuffer(nil),
			[]string{"--help"},
			"Manages the lifecycle of a kubernetes release\n" +
				"\n" +
				"Usage:\n" +
				"  releases COMMAND [flags]\n" +
				"\n" +
				"Available Commands:\n" +
				"  describe    Describe registered releases with states information.\n" +
				"  generate    Adds a new release template.\n" +
				"  install     Deploys release resources on running cluster.\n" +
				"  uninstall   Stops a running release and removes the resources.\n" +
				"  update      Redeploy a modified release and track revision version.\n" +
				"\n" +
				"Flags:\n" +
				"      --help   Show help information.\n",
		},
	}

	for i, testCase := range testCases {
		failMsg := fmt.Sprintf("testCase: %d %v", i, testCase)
		runner := func() {
			err := testCase.cmd.Execute("releases", testCase.buffer, testCase.args)
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

func TestReleaseSuite(t *testing.T) {
	suite.Run(t, new(ReleaseSuite))
}
