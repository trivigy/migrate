package types

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/blang/semver"
	"github.com/davecgh/go-spew/spew"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	v1core "k8s.io/api/core/v1"
	v1ext "k8s.io/api/extensions/v1beta1"
	v1meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
)

type ReleaseSuite struct {
	suite.Suite
}

func (r *ReleaseSuite) TestRelease_MarshalUnmarshal() {
	testCases := []struct {
		shouldFail bool
		onFail     string
		release    *Release
	}{
		{
			false, "",
			&Release{
				Name:    "coredb",
				Version: semver.MustParse("0.0.1"),
				Manifests: []runtime.Object{
					&v1core.Service{
						TypeMeta: v1meta.TypeMeta{
							APIVersion: "v1",
							Kind:       "Service",
						},
						ObjectMeta: v1meta.ObjectMeta{
							Name: "unittest",
						},
						Spec: v1core.ServiceSpec{
							Ports: []v1core.ServicePort{
								{
									Name:       "postgres",
									Port:       5432,
									TargetPort: intstr.FromInt(5432),
								},
							},
						},
					},
					&v1ext.Ingress{
						TypeMeta: v1meta.TypeMeta{
							APIVersion: "extensions/v1beta1",
							Kind:       "Ingress",
						},
						ObjectMeta: v1meta.ObjectMeta{
							Name: "unittest",
						},
						Spec: v1ext.IngressSpec{
							Rules: []v1ext.IngressRule{
								{
									IngressRuleValue: v1ext.IngressRuleValue{
										HTTP: &v1ext.HTTPIngressRuleValue{
											Paths: []v1ext.HTTPIngressPath{
												{
													Backend: v1ext.IngressBackend{
														ServiceName: "unittest",
														ServicePort: intstr.FromInt(80),
													},
												},
											},
										},
									},
								},
								{
									IngressRuleValue: v1ext.IngressRuleValue{
										HTTP: &v1ext.HTTPIngressRuleValue{
											Paths: []v1ext.HTTPIngressPath{
												{
													Backend: v1ext.IngressBackend{
														ServiceName: "unittest",
														ServicePort: intstr.FromInt(80),
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
			},
		},
	}

	for i, tc := range testCases {
		failMsg := fmt.Sprintf("test: %d %v", i, spew.Sprint(tc))
		runner := func() {
			rbytes, err := json.Marshal(tc.release)
			if err != nil {
				panic(err.Error())
			}

			actual := &Release{}
			err = json.Unmarshal(rbytes, actual)
			if err != nil {
				panic(err.Error())
			}

			assert.EqualValues(r.T(), tc.release, actual)
		}

		if tc.shouldFail {
			assert.PanicsWithValue(r.T(), tc.onFail, runner, failMsg)
		} else {
			assert.NotPanics(r.T(), runner, failMsg)
		}
	}
}

func TestReleaseSuite(t *testing.T) {
	suite.Run(t, new(ReleaseSuite))
}
