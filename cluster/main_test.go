package cluster

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/Pallinder/go-randomdata"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/trivigy/migrate/v2/config"
	"github.com/trivigy/migrate/v2/driver/provider"
)

type ClusterSuite struct {
	suite.Suite
	config map[string]config.Cluster
}

func (r *ClusterSuite) SetupSuite() {
	r.config = map[string]config.Cluster{
		"default": {
			Driver: provider.Kind{
				Name: strings.ToLower(randomdata.SillyName()),
			},
		},
	}
}

func (r *ClusterSuite) TearDownSuite() {
	buffer := bytes.NewBuffer(nil)
	assert.Nil(r.T(), r.config["default"].Driver.TearDown(buffer))
}

func (r *ClusterSuite) TestClusterCommand() {
	testCases := []struct {
		shouldFail bool
		onFail     string
		buffer     *bytes.Buffer
		args       []string
		output     string
	}{
		{
			false, "",
			bytes.NewBuffer(nil),
			[]string{"--help"},
			"Kubernetes cluster release and deployment controller\n" +
				"\n" +
				"Usage:\n" +
				"  cluster [command]\n" +
				"\n" +
				"Available Commands:\n" +
				"  create      Initializes a new kubernetes cluster.\n" +
				"  destroy     Stops an existing running kubernetes cluster.\n" +
				"  release     Manages the lifecycle of a kubernetes release.\n" +
				"\n" +
				"Flags:\n" +
				"  -e, --env ENV   Run with env ENV configurations. (default \"default\")\n" +
				"      --help      Show help information.\n" +
				"\n" +
				"Use \"cluster [command] --help\" for more information about a command.\n",
		},
	}

	for i, testCase := range testCases {
		failMsg := fmt.Sprintf("testCase: %d %v", i, testCase)
		runner := func() {
			command := NewCluster(r.config)
			err := command.Execute("cluster", testCase.buffer, testCase.args)
			if err != nil {
				panic(testCase.buffer.String())
			}

			if testCase.output != testCase.buffer.String() {
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
