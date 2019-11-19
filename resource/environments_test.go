package resource

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/trivigy/migrate/v2/resource/database"
	"github.com/trivigy/migrate/v2/resource/database/migrations"
	"github.com/trivigy/migrate/v2/resource/kubernetes"
	"github.com/trivigy/migrate/v2/resource/kubernetes/releases"
	"github.com/trivigy/migrate/v2/resource/primitive"
	"github.com/trivigy/migrate/v2/types"
)

type EnvironmentsSuite struct {
	suite.Suite
}

func (r *EnvironmentsSuite) TestEnvironmentsCommand() {
	Command := Environments{
		"development": Environment{
			"developmentDatabase": Database{
				"create":  primitive.Create{},
				"destroy": primitive.Destroy{},
				"source":  primitive.Source{},
				"migrations": database.Migrations{
					"generate": migrations.Generate{},
					"up":       migrations.Up{},
					"down":     migrations.Down{},
					"report":   migrations.Report{},
				},
			},
			"developmentKubernetes": Kubernetes{
				"create":  primitive.Create{},
				"destroy": primitive.Destroy{},
				"source":  primitive.Source{},
				"releases": kubernetes.Releases{
					"generate":  releases.Generate{},
					"install":   releases.Install{},
					"uninstall": releases.Uninstall{},
					"update":    releases.Update{},
					"describe":  releases.Describe{},
				},
			},
			"developmentDomainName": DomainName{
				"create":  primitive.Create{},
				"destroy": primitive.Destroy{},
			},
		},
		"staging": Environment{
			"stagingDatabase": Database{
				"create":  primitive.Create{},
				"destroy": primitive.Destroy{},
				"source":  primitive.Source{},
				"migrations": database.Migrations{
					"generate": migrations.Generate{},
					"up":       migrations.Up{},
					"down":     migrations.Down{},
					"report":   migrations.Report{},
				},
			},
			"stagingKubernetes": Kubernetes{
				"create":  primitive.Create{},
				"destroy": primitive.Destroy{},
				"source":  primitive.Source{},
				"releases": kubernetes.Releases{
					"generate":  releases.Generate{},
					"install":   releases.Install{},
					"uninstall": releases.Uninstall{},
					"update":    releases.Update{},
					"describe":  releases.Describe{},
				},
			},
			"stagingDomainName": DomainName{
				"create":  primitive.Create{},
				"destroy": primitive.Destroy{},
			},
		},
	}

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
				"Usage:\n" +
				"  environments [flags] COMMAND\n" +
				"\n" +
				"Available Commands:\n" +
				"  developmentDatabase   SQL database deployment and migrations management tool.\n" +
				"  developmentDomainName Controls instance of domain name service resource.\n" +
				"  developmentKubernetes Kubernetes cluster release and deployment controller.\n" +
				"\n" +
				"Flags:\n" +
				"  -e, --env ENV   Run with env ENV configurations. (default \"development\")\n" +
				"      --help      Show help information.\n",
			Command,
			bytes.NewBuffer(nil),
			[]string{},
			"",
		},
		{
			false, "",
			Command,
			bytes.NewBuffer(nil),
			[]string{"--help"},
			"Usage:\n" +
				"  environments [flags] COMMAND\n" +
				"\n" +
				"Available Commands:\n" +
				"  developmentDatabase   SQL database deployment and migrations management tool.\n" +
				"  developmentDomainName Controls instance of domain name service resource.\n" +
				"  developmentKubernetes Kubernetes cluster release and deployment controller.\n" +
				"\n" +
				"Flags:\n" +
				"  -e, --env ENV   Run with env ENV configurations. (default \"development\")\n" +
				"      --help      Show help information.\n",
		},
		{
			true,
			"accepts 1 arg(s), received 0 for \"environments\"\n" +
				"\n" +
				"Usage:\n" +
				"  environments [flags] COMMAND\n" +
				"\n" +
				"Available Commands:\n" +
				"  stagingDatabase   SQL database deployment and migrations management tool.\n" +
				"  stagingDomainName Controls instance of domain name service resource.\n" +
				"  stagingKubernetes Kubernetes cluster release and deployment controller.\n" +
				"\n" +
				"Flags:\n" +
				"  -e, --env ENV   Run with env ENV configurations. (default \"development\")\n" +
				"      --help      Show help information.\n",
			Command,
			bytes.NewBuffer(nil),
			[]string{"-e", "staging"},
			"",
		},
		{
			false, "",
			Command,
			bytes.NewBuffer(nil),
			[]string{"-e", "staging", "--help"},
			"Usage:\n" +
				"  environments [flags] COMMAND\n" +
				"\n" +
				"Available Commands:\n" +
				"  stagingDatabase   SQL database deployment and migrations management tool.\n" +
				"  stagingDomainName Controls instance of domain name service resource.\n" +
				"  stagingKubernetes Kubernetes cluster release and deployment controller.\n" +
				"\n" +
				"Flags:\n" +
				"  -e, --env ENV   Run with env ENV configurations. (default \"development\")\n" +
				"      --help      Show help information.\n",
		},
		{
			false, "",
			Command,
			bytes.NewBuffer(nil),
			[]string{"--help", "-e", "staging"},
			"Usage:\n" +
				"  environments [flags] COMMAND\n" +
				"\n" +
				"Available Commands:\n" +
				"  stagingDatabase   SQL database deployment and migrations management tool.\n" +
				"  stagingDomainName Controls instance of domain name service resource.\n" +
				"  stagingKubernetes Kubernetes cluster release and deployment controller.\n" +
				"\n" +
				"Flags:\n" +
				"  -e, --env ENV   Run with env ENV configurations. (default \"development\")\n" +
				"      --help      Show help information.\n",
		},
		{
			false, "",
			Command,
			bytes.NewBuffer(nil),
			[]string{"--help", "defaultDatabase"},
			"Usage:\n" +
				"  environments [flags] COMMAND\n" +
				"\n" +
				"Available Commands:\n" +
				"  developmentDatabase   SQL database deployment and migrations management tool.\n" +
				"  developmentDomainName Controls instance of domain name service resource.\n" +
				"  developmentKubernetes Kubernetes cluster release and deployment controller.\n" +
				"\n" +
				"Flags:\n" +
				"  -e, --env ENV   Run with env ENV configurations. (default \"development\")\n" +
				"      --help      Show help information.\n",
		},
		{
			false, "",
			Command,
			bytes.NewBuffer(nil),
			[]string{"--help", "defaultDatabase", "--help"},
			"Usage:\n" +
				"  environments [flags] COMMAND\n" +
				"\n" +
				"Available Commands:\n" +
				"  developmentDatabase   SQL database deployment and migrations management tool.\n" +
				"  developmentDomainName Controls instance of domain name service resource.\n" +
				"  developmentKubernetes Kubernetes cluster release and deployment controller.\n" +
				"\n" +
				"Flags:\n" +
				"  -e, --env ENV   Run with env ENV configurations. (default \"development\")\n" +
				"      --help      Show help information.\n",
		},
		{
			false, "",
			Command,
			bytes.NewBuffer(nil),
			[]string{"developmentDatabase", "--help"},
			"SQL database deployment and migrations management tool\n" +
				"\n" +
				"Usage:\n" +
				"  environments developmentDatabase COMMAND [flags]\n" +
				"\n" +
				"Available Commands:\n" +
				"  create      Constructs and starts a new instance of this resource.\n" +
				"  destroy     Stops and removes running instance of this resource.\n" +
				"  migrations  Manages the lifecycle of a database migration.\n" +
				"  source      Prints the data source name as a connection string.\n" +
				"\n" +
				"Flags:\n" +
				"      --help   Show help information.\n",
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
