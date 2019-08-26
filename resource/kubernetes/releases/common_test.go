package releases

import (
	"bytes"
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

	"github.com/trivigy/migrate/v2/driver/docker"
	"github.com/trivigy/migrate/v2/types"
)

type ReleasesSuite struct {
	suite.Suite
	Namespace string          `json:"namespace" yaml:"namespace"`
	Releases  *types.Releases `json:"releases" yaml:"releases"`
	Driver    interface {
		types.Creator
		types.Destroyer
		types.KubeConfiged
	} `json:"driver" yaml:"driver"`
}

func (r *ReleasesSuite) SetupSuite() {
	r.Namespace = "unittest"
	r.Releases = &types.Releases{
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

	r.Driver = docker.KIND{
		Name: strings.ToLower(randomdata.SillyName()),
	}

	buffer := bytes.NewBuffer(nil)
	assert.Nil(r.T(), r.Driver.Create(buffer))
}

func (r *ReleasesSuite) TearDownSuite() {
	buffer := bytes.NewBuffer(nil)
	assert.Nil(r.T(), r.Driver.Destroy(buffer))
}

func (r *ReleasesSuite) TearDownTest() {
	buffer := bytes.NewBuffer(nil)
	uninstall := Uninstall{Namespace: r.Namespace, Releases: r.Releases, Driver: r.Driver}
	assert.Nil(r.T(), uninstall.Execute("uninstall", buffer, []string{}))
}

func TestReleasesSuite(t *testing.T) {
	suite.Run(t, new(ReleasesSuite))
}
