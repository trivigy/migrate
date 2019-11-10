package resource

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/trivigy/migrate/v2/types"
)

type ClusterSuite struct {
	suite.Suite
}

func (r *ClusterSuite) TestClusterCommand() {
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
			"accepts 1 arg(s), received 0 for \"kubernetes\"\n" +
				"\n" +
				"Usage:\n" +
				"  kubernetes COMMAND [flags]\n" +
				"\n" +
				"Available Commands:\n" +
				"  create      Constructs and starts a new instance of this resource.\n" +
				"  destroy     Stops and removes running instance of this resource.\n" +
				"  releases    Manages the lifecycle of a kubernetes release.\n" +
				"  source      Prints the data source name as a connection string.\n" +
				"\n" +
				"Flags:\n      --help   Show help information.\n",
			Kubernetes{},
			bytes.NewBuffer(nil),
			[]string{},
			"",
		},
		{
			false, "",
			Kubernetes{},
			bytes.NewBuffer(nil),
			[]string{"--help"},
			"Kubernetes cluster release and deployment controller\n" +
				"\n" +
				"Usage:\n" +
				"  kubernetes COMMAND [flags]\n" +
				"\n" +
				"Available Commands:\n" +
				"  create      Constructs and starts a new instance of this resource.\n" +
				"  destroy     Stops and removes running instance of this resource.\n" +
				"  releases    Manages the lifecycle of a kubernetes release.\n" +
				"  source      Prints the data source name as a connection string.\n" +
				"\n" +
				"Flags:\n" +
				"      --help   Show help information.\n",
		},
	}

	for i, testCase := range testCases {
		failMsg := fmt.Sprintf("testCase: %d %v", i, testCase)
		runner := func() {
			err := testCase.cmd.Execute("kubernetes", testCase.buffer, testCase.args)
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

func TestClusterSuite(t *testing.T) {
	suite.Run(t, new(ClusterSuite))
}
