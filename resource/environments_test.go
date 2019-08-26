package resource

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/trivigy/migrate/v2/types"
)

type EnvironmentsSuite struct {
	suite.Suite
}

func (r *EnvironmentsSuite) TestEnvironmentsCommand() {
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
			"accepts 1 arg(s), received 0 for \"environments\"\n" +
				"\n" +
				"Usage:  environments [flags] command",
			Environments{
				"default": Collection{
					"defaultDatabase":   Database{},
					"defaultKubernetes": Kubernetes{},
				},
				"staging": Collection{
					"stagingDatabase":   Database{},
					"stagingKubernetes": Kubernetes{},
				},
			},
			bytes.NewBuffer(nil),
			[]string{},
			"",
		},
		{
			false, "",
			Environments{
				"default": Collection{
					"defaultDatabase":   Database{},
					"defaultKubernetes": Kubernetes{},
				},
				"staging": Collection{
					"stagingDatabase":   Database{},
					"stagingKubernetes": Kubernetes{},
				},
			},
			bytes.NewBuffer(nil),
			[]string{"--help"},
			"Usage:\n" +
				"  environments [flags] command\n" +
				"\n" +
				"Available Commands:\n" +
				"  defaultDatabase   SQL database deployment and migrations management tool.\n" +
				"  defaultKubernetes Kubernetes cluster release and deployment controller.\n" +
				"\n" +
				"Flags:\n" +
				"  -e, --env ENV   Run with env ENV configurations. (default \"default\")\n" +
				"      --help      Show help information.\n",
		},
		{
			true,
			"accepts 1 arg(s), received 0 for \"environments\"\n" +
				"\n" +
				"Usage:  environments [flags] command",
			Environments{
				"default": Collection{
					"defaultDatabase":   Database{},
					"defaultKubernetes": Kubernetes{},
				},
				"staging": Collection{
					"stagingDatabase":   Database{},
					"stagingKubernetes": Kubernetes{},
				},
			},
			bytes.NewBuffer(nil),
			[]string{"-e", "staging"},
			"",
		},
		{
			false, "",
			Environments{
				"default": Collection{
					"defaultDatabase":   Database{},
					"defaultKubernetes": Kubernetes{},
				},
				"staging": Collection{
					"stagingDatabase":   Database{},
					"stagingKubernetes": Kubernetes{},
				},
			},
			bytes.NewBuffer(nil),
			[]string{"-e", "staging", "--help"},
			"Usage:\n" +
				"  environments [flags] command\n" +
				"\n" +
				"Available Commands:\n" +
				"  stagingDatabase   SQL database deployment and migrations management tool.\n" +
				"  stagingKubernetes Kubernetes cluster release and deployment controller.\n" +
				"\n" +
				"Flags:\n" +
				"  -e, --env ENV   Run with env ENV configurations. (default \"default\")\n" +
				"      --help      Show help information.\n",
		},
		{
			false, "",
			Environments{
				"default": Collection{
					"defaultDatabase":   Database{},
					"defaultKubernetes": Kubernetes{},
				},
				"staging": Collection{
					"stagingDatabase":   Database{},
					"stagingKubernetes": Kubernetes{},
				},
			},
			bytes.NewBuffer(nil),
			[]string{"--help", "-e", "staging"},
			"Usage:\n" +
				"  environments [flags] command\n" +
				"\n" +
				"Available Commands:\n" +
				"  stagingDatabase   SQL database deployment and migrations management tool.\n" +
				"  stagingKubernetes Kubernetes cluster release and deployment controller.\n" +
				"\n" +
				"Flags:\n" +
				"  -e, --env ENV   Run with env ENV configurations. (default \"default\")\n" +
				"      --help      Show help information.\n",
		},
		{
			false, "",
			Environments{
				"default": Collection{
					"defaultDatabase":   Database{},
					"defaultKubernetes": Kubernetes{},
				},
				"staging": Collection{
					"stagingDatabase":   Database{},
					"stagingKubernetes": Kubernetes{},
				},
			},
			bytes.NewBuffer(nil),
			[]string{"--help", "defaultDatabase"},
			"Usage:\n" +
				"  environments [flags] command\n" +
				"\n" +
				"Available Commands:\n" +
				"  defaultDatabase   SQL database deployment and migrations management tool.\n" +
				"  defaultKubernetes Kubernetes cluster release and deployment controller.\n" +
				"\n" +
				"Flags:\n" +
				"  -e, --env ENV   Run with env ENV configurations. (default \"default\")\n" +
				"      --help      Show help information.\n",
		},
		{
			false, "",
			Environments{
				"default": Collection{
					"defaultDatabase":   Database{},
					"defaultKubernetes": Kubernetes{},
				},
				"staging": Collection{
					"stagingDatabase":   Database{},
					"stagingKubernetes": Kubernetes{},
				},
			},
			bytes.NewBuffer(nil),
			[]string{"--help", "defaultDatabase", "--help"},
			"Usage:\n" +
				"  environments [flags] command\n" +
				"\n" +
				"Available Commands:\n" +
				"  defaultDatabase   SQL database deployment and migrations management tool.\n" +
				"  defaultKubernetes Kubernetes cluster release and deployment controller.\n" +
				"\n" +
				"Flags:\n" +
				"  -e, --env ENV   Run with env ENV configurations. (default \"default\")\n" +
				"      --help      Show help information.\n",
		},
		{
			false, "",
			Environments{
				"default": Collection{
					"defaultDatabase":   Database{},
					"defaultKubernetes": Kubernetes{},
				},
				"staging": Collection{
					"stagingDatabase":   Database{},
					"stagingKubernetes": Kubernetes{},
				},
			},
			bytes.NewBuffer(nil),
			[]string{"defaultDatabase", "--help"},
			"SQL database deployment and migrations management tool\n" +
				"\n" +
				"Usage:\n" +
				"  environments defaultDatabase [command]\n" +
				"\n" +
				"Available Commands:\n" +
				"  create      Constructs and starts a new instance of this resource.\n" +
				"  destroy     Stops and removes running instance of this resource.\n" +
				"  migrations  Manages the lifecycle of a database migration.\n" +
				"  source      Print the data source name as a connection string.\n" +
				"\n" +
				"Flags:\n" +
				"      --help   Show help information.\n" +
				"\n" +
				"Use \"environments defaultDatabase [command] --help\" for more information about a command.\n",
		},
	}

	for i, testCase := range testCases {
		failMsg := fmt.Sprintf("testCase: %d %v", i, testCase)
		runner := func() {
			err := testCase.cmd.Execute("environments", testCase.buffer, testCase.args)
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

func TestEnvironmentsSuite(t *testing.T) {
	suite.Run(t, new(EnvironmentsSuite))
}
