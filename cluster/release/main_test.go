package release

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/Pallinder/go-randomdata"
	"github.com/blang/semver"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	v1apps "k8s.io/api/apps/v1"
	v1core "k8s.io/api/core/v1"
	v1meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	"github.com/trivigy/migrate/v2/config"
	"github.com/trivigy/migrate/v2/driver/provider"
	"github.com/trivigy/migrate/v2/types"
)

type ReleaseSuite struct {
	suite.Suite
	config map[string]config.Cluster
}

func (r *ReleaseSuite) SetupSuite() {
	releases := &types.Releases{
		{
			Name:    "create-unittest-cluster",
			Version: semver.Version{Major: 0, Minor: 0, Patch: 1},
			Manifests: []interface{}{
				&v1core.Service{
					TypeMeta: v1meta.TypeMeta{
						APIVersion: "v1",
						Kind:       "Service",
					},
					ObjectMeta: v1meta.ObjectMeta{
						Name: "locker",
						Labels: map[string]string{
							"app": "locker",
						},
					},
					Spec: v1core.ServiceSpec{
						Ports: []v1core.ServicePort{
							{
								Port:       80,
								TargetPort: intstr.FromInt(80),
							},
						},
						Selector: map[string]string{
							"app": "locker",
						},
					},
				},
				&v1apps.Deployment{
					TypeMeta: v1meta.TypeMeta{
						APIVersion: "apps/v1",
						Kind:       "Deployment",
					},
					ObjectMeta: v1meta.ObjectMeta{
						Name: "unittest",
					},
					Spec: v1apps.DeploymentSpec{
						Replicas: &[]int32{1}[0],
						Selector: &v1meta.LabelSelector{
							MatchLabels: map[string]string{
								"app": "unittest",
							},
						},
						Template: v1core.PodTemplateSpec{
							ObjectMeta: v1meta.ObjectMeta{
								Labels: map[string]string{
									"app": "unittest",
								},
							},
							Spec: v1core.PodSpec{
								Containers: []v1core.Container{
									{
										Name:  "unittest",
										Image: "nginx:latest",
										Ports: []v1core.ContainerPort{
											{
												ContainerPort: 80,
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	r.config = map[string]config.Cluster{
		"default": {
			Namespace: "unittest",
			Releases:  releases,
			Driver: provider.Kind{
				Name: strings.ToLower(randomdata.SillyName()),
			},
		},
	}

	buffer := bytes.NewBuffer(nil)
	assert.Nil(r.T(), r.config["default"].Driver.Setup(buffer))
}

func (r *ReleaseSuite) TearDownSuite() {
	buffer := bytes.NewBuffer(nil)
	assert.Nil(r.T(), r.config["default"].Driver.TearDown(buffer))
}

func (r *ReleaseSuite) TearDownTest() {
	// buffer := bytes.NewBuffer(nil)
	// command := NewUninstall(r.config)
	// assert.Nil(r.T(), command.Execute("down", buffer, []string{}))
}

func (r *ReleaseSuite) TestReleaseCommand() {
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
			"Manages the lifecycle of a kubernetes release\n" +
				"\n" +
				"Usage:\n" +
				"  release [command]\n" +
				"\n" +
				"Available Commands:\n" +
				"  generate    Adds a new release template.\n" +
				"  history     Prints revisions history of deployed releases.\n" +
				"  inspect     Prints release resources detail information.\n" +
				"  install     Deploys release resources on running cluster.\n" +
				"  list        List registered releases with states information.\n" +
				"  uninstall   Stops a running release and removes the resources.\n" +
				"  upgrade     Redeploy a modified release and track revision version.\n" +
				"\n" +
				"Flags:\n" +
				"  -e, --env ENV   Run with env ENV configurations. (default \"default\")\n" +
				"      --help      Show help information.\n" +
				"\n" +
				"Use \"release [command] --help\" for more information about a command.\n",
		},
	}

	for i, testCase := range testCases {
		failMsg := fmt.Sprintf("testCase: %d %v", i, testCase)
		runner := func() {
			command := NewRelease(r.config)
			err := command.Execute("release", testCase.buffer, testCase.args)
			if err != nil {
				panic(testCase.buffer.String())
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
