package release

func (r *ReleaseSuite) TestInstallCommand() {
	// cmdName := "install"
	// releases := types.Releases{
	// 	{
	// 		Name:    "create-unittest-table",
	// 		Version: semver.MustParse("0.0.1"),
	// 		Manifests: []interface{}{
	// 			&v1core.Service{
	// 				TypeMeta: v1meta.TypeMeta{
	// 					APIVersion: "v1",
	// 					Kind:       "Service",
	// 				},
	// 				ObjectMeta: v1meta.ObjectMeta{
	// 					Name: "hello-kubernetes",
	// 				},
	// 				Spec: v1core.ServiceSpec{
	// 					Type: v1core.ServiceTypeLoadBalancer,
	// 					Ports: []v1core.ServicePort{
	// 						{
	// 							Port: 80,
	// 							TargetPort: intstr.IntOrString{
	// 								Type:   intstr.String,
	// 								StrVal: "8080",
	// 							},
	// 						},
	// 					},
	// 					Selector: map[string]string{
	// 						"app": "hello-kubernetes",
	// 					},
	// 				},
	// 			},
	// 			&v1apps.Deployment{
	// 				TypeMeta: v1meta.TypeMeta{
	// 					APIVersion: "apps/v1",
	// 					Kind:       "Deployment",
	// 				},
	// 				ObjectMeta: v1meta.ObjectMeta{
	// 					Name: "hello-kubernetes",
	// 				},
	// 				Spec: v1apps.DeploymentSpec{
	// 					Replicas: &[]int32{1}[0],
	// 					Selector: &v1meta.LabelSelector{
	// 						MatchLabels: map[string]string{
	// 							"app": "hello-kubernetes",
	// 						},
	// 					},
	// 					Template: v1core.PodTemplateSpec{
	// 						ObjectMeta: v1meta.ObjectMeta{
	// 							Labels: map[string]string{
	// 								"app": "hello-kubernetes",
	// 							},
	// 						},
	// 						Spec: v1core.PodSpec{
	// 							Containers: []v1core.Container{
	// 								{
	// 									Name:  "hello-kubernetes",
	// 									Image: "paulbouwer/hello-kubernetes:1.5",
	// 									Ports: []v1core.ContainerPort{
	// 										{
	// 											ContainerPort: 8080,
	// 										},
	// 									},
	// 								},
	// 							},
	// 						},
	// 					},
	// 				},
	// 			},
	// 		},
	// 	},
	// }
	//
	// defaultConfig := map[string]config.Cluster{}
	// if os.Getenv("CI") != "true" {
	// 	defaultConfig = map[string]config.Cluster{
	// 		"default": {
	// 			Releases:  releases,
	// 			Namespace: "unittest",
	// 			Driver:    r.driver,
	// 		},
	// 	}
	// } else {
	// 	defaultConfig = map[string]config.Cluster{
	// 		"default": {
	// 			Releases: releases,
	// 			// Driver: provider.GCP{
	// 			// 	Name:        branchName,
	// 			// 	ProjectID:   "digicontract-248304",
	// 			// 	Location:    "us-east4-b",
	// 			// 	MachineType: "n1-standard-1",
	// 			// 	ImageType:   "ubuntu",
	// 			// },
	// 		},
	// 	}
	// }
	//
	// testCases := []struct {
	// 	shouldFail bool
	// 	onFail     string
	// 	buffer     *bytes.Buffer
	// 	args       []string
	// 	output     string
	// }{
	// 	{
	// 		false, "",
	// 		bytes.NewBuffer(nil),
	// 		[]string{},
	// 		"---",
	// 	},
	// }
	//
	// for i, testCase := range testCases {
	// 	failMsg := fmt.Sprintf("testCase: %d %v", i, testCase)
	// 	runner := func() {
	// 		command := NewInstall(defaultConfig)
	// 		err := command.Execute(cmdName, testCase.buffer, testCase.args)
	// 		if err != nil {
	// 			panic(testCase.buffer.String())
	// 		}
	//
	// 		if testCase.output != testCase.buffer.String() {
	// 			panic(testCase.buffer.String())
	// 		}
	// 	}
	//
	// 	if testCase.shouldFail {
	// 		assert.PanicsWithValue(r.T(), testCase.onFail, runner, failMsg)
	// 	} else {
	// 		assert.NotPanics(r.T(), runner, failMsg)
	// 	}
	// }
}
