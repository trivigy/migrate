package release

import (
	"bytes"
	"fmt"

	"github.com/blang/semver"
	"github.com/stretchr/testify/assert"
	v1apps "k8s.io/api/apps/v1"
	v1core "k8s.io/api/core/v1"
	v1meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	"github.com/trivigy/migrate/v2/config"
	"github.com/trivigy/migrate/v2/types"
)

func (r *ReleaseSuite) TestListCommand() {
	cmdName := "list"
	releases := types.Releases{
		{
			Name:    "create-unittest-table",
			Version: semver.MustParse("0.0.1"),
			Manifests: []interface{}{
				&v1core.Service{
					TypeMeta: v1meta.TypeMeta{
						APIVersion: "v1",
						Kind:       "Service",
					},
					ObjectMeta: v1meta.ObjectMeta{
						Name: "hello-kubernetes",
					},
					Spec: v1core.ServiceSpec{
						Type: v1core.ServiceTypeLoadBalancer,
						Ports: []v1core.ServicePort{
							{
								Port: 80,
								TargetPort: intstr.IntOrString{
									Type:   intstr.String,
									StrVal: "8080",
								},
							},
						},
						Selector: map[string]string{
							"app": "hello-kubernetes",
						},
					},
				},
				&v1apps.Deployment{
					TypeMeta: v1meta.TypeMeta{
						APIVersion: "apps/v1",
						Kind:       "Deployment",
					},
					ObjectMeta: v1meta.ObjectMeta{
						Name: "hello-kubernetes",
					},
					Spec: v1apps.DeploymentSpec{
						Replicas: &[]int32{1}[0],
						Selector: &v1meta.LabelSelector{
							MatchLabels: map[string]string{
								"app": "hello-kubernetes",
							},
						},
						Template: v1core.PodTemplateSpec{
							ObjectMeta: v1meta.ObjectMeta{
								Labels: map[string]string{
									"app": "hello-kubernetes",
								},
							},
							Spec: v1core.PodSpec{
								Containers: []v1core.Container{
									{
										Name:  "hello-kubernetes",
										Image: "paulbouwer/hello-kubernetes:1.5",
										Ports: []v1core.ContainerPort{
											{
												ContainerPort: 8080,
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

	defaultConfig := map[string]config.Cluster{
		"default": {
			Releases:  releases,
			Namespace: "unittest",
			Driver:    r.driver,
		},
	}

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
			[]string{},
			"+---+-----------------------+---------+------------------+------------+----------+\n" +
				"| # |         NAME          | VERSION |     MANIFEST     |    KIND    |  STATUS  |\n" +
				"+---+-----------------------+---------+------------------+------------+----------+\n" +
				"| 1 | create-unittest-table | 0.0.1   | hello-kubernetes | Service    | NotFound |\n" +
				"|   |                       |         | hello-kubernetes | Deployment | NotFound |\n" +
				"+---+-----------------------+---------+------------------+------------+----------+\n",
		},
	}

	for i, testCase := range testCases {
		failMsg := fmt.Sprintf("testCase: %d %v", i, testCase)
		runner := func() {
			command := NewList(defaultConfig)
			err := command.Execute(cmdName, testCase.buffer, testCase.args)
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
