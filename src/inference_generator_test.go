package main

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"kusionstack.io/kusion-module-framework/pkg/module"
	apiv1 "kusionstack.io/kusion/pkg/apis/api.kusion.io/v1"
	v1 "kusionstack.io/kusion/pkg/apis/api.kusion.io/v1"
)

func TestInferenceModule_Generator(t *testing.T) {
	r := &module.GeneratorRequest{
		Project: "test-project",
		Stack:   "test-stack",
		App:     "test-app",
		Workload: &apiv1.Workload{
			Header: apiv1.Header{
				Type: "Service",
			},
			Service: &apiv1.Service{},
		},
	}

	testcases := []struct {
		name            string
		devModuleConfig apiv1.Accessory
		platformConfig  apiv1.GenericConfig
		expectedErr     error
	}{
		{
			name: "Generate inference by Ollama",
			devModuleConfig: apiv1.Accessory{
				"model":     "llama3",
				"framework": "Ollama",
			},
			expectedErr: nil,
		},
		// {
		// 	name: "Unsupported framework",
		// 	devModuleConfig: apiv1.Accessory{
		// 		"model":     "llama3",
		// 		"framework": "unsupported-type",
		// 	},
		// 	expectedErr: errors.New("unsupported framework type"),
		// },
	}

	for _, tc := range testcases {
		inference := &Inference{}
		t.Run(tc.name, func(t *testing.T) {
			r.DevConfig = tc.devModuleConfig
			r.PlatformConfig = tc.platformConfig

			res, err := inference.Generate(context.Background(), r)
			if tc.expectedErr != nil {
				assert.ErrorContains(t, err, tc.expectedErr.Error())
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, res)
			}
		})
	}
}

func TestInferenceModule_GenerateDeployment(t *testing.T) {
	r := &module.GeneratorRequest{
		Project: "test-project",
		Stack:   "test-stack",
		App:     "test-app",
		Workload: &v1.Workload{
			Header: v1.Header{
				Type: "Service",
			},
			Service: &v1.Service{},
		},
	}

	inference := &Inference{
		Model:     "qwen",
		Framework: "ollama",
	}

	res, err := inference.generateDeployment(r)

	assert.NotNil(t, res)
	assert.NoError(t, err)
}

func TestInferenceModule_GeneratePodSpec(t *testing.T) {
	r := &module.GeneratorRequest{
		Project: "test-project",
		Stack:   "test-stack",
		App:     "test-app",
		Workload: &v1.Workload{
			Header: v1.Header{
				Type: "Service",
			},
			Service: &v1.Service{},
		},
	}

	inference := &Inference{
		Model:     "qwen",
		Framework: "ollama",
	}

	res, err := inference.generatePodSpec(r)

	assert.NotNil(t, res)
	assert.NoError(t, err)
}

func TestInferenceModule_GenerateService(t *testing.T) {
	r := &module.GeneratorRequest{
		Project: "test-project",
		Stack:   "test-stack",
		App:     "test-app",
		Workload: &v1.Workload{
			Header: v1.Header{
				Type: "Service",
			},
			Service: &v1.Service{},
		},
	}

	inference := &Inference{
		Model:     "qwen",
		Framework: "ollama",
	}

	res, svcName, err := inference.generateService(r)

	assert.NotNil(t, res)
	assert.NotNil(t, svcName)
	assert.NoError(t, err)
}
