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
			platformConfig: nil,
			expectedErr:    nil,
		},
		// {
		// 	name: "Unsupported framework",
		// 	devModuleConfig: apiv1.Accessory{
		// 		"model":     "llama3",
		// 		"framework": "unsupported-type",
		// 	},
		// platformConfig: nil,
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

func TestInferenceModule_CompleteConfig(t *testing.T) {
	testcases := []struct {
		name              string
		devModuleConfig   apiv1.Accessory
		platformConfig    apiv1.GenericConfig
		expectedInference *Inference
	}{
		{
			name: "Default inference config",
			devModuleConfig: apiv1.Accessory{
				"model":     "qwen",
				"framework": "Ollama",
			},
			platformConfig: nil,
			expectedInference: &Inference{
				Model:       "qwen",
				Framework:   "Ollama",
				System:      "",
				Template:    "",
				TopK:        40,
				TopP:        0.9,
				Temperature: 0.8,
				NumPredict:  128,
				NumCtx:      2048,
			},
		},
		{
			name: "Custom inference config",
			devModuleConfig: apiv1.Accessory{
				"model":       "qwen",
				"framework":   "Ollama",
				"top_k":       50,
				"top_p":       0.5,
				"temperature": 0.5,
				"num_predict": 256,
				"num_ctx":     4096,
			},
			platformConfig: nil,
			expectedInference: &Inference{
				Model:       "qwen",
				Framework:   "Ollama",
				System:      "",
				Template:    "",
				TopK:        50,
				TopP:        0.5,
				Temperature: 0.5,
				NumPredict:  256,
				NumCtx:      4096,
			},
		},
	}

	for _, tc := range testcases {
		infer := &Inference{}
		t.Run(tc.name, func(t *testing.T) {
			_ = infer.CompleteConfig(tc.devModuleConfig, tc.platformConfig)
			assert.Equal(t, tc.expectedInference, infer)
		})
	}
}

func TestInferenceModule_ValidateConfig(t *testing.T) {
	t.Run("test framework", func(t *testing.T) {
		infer := &Inference{
			Model:       "qwen",
			Framework:   "unsupport_framework",
			System:      "",
			Template:    "",
			TopK:        40,
			TopP:        0.9,
			Temperature: 0.8,
			NumPredict:  128,
			NumCtx:      2048,
		}
		err := infer.ValidateConfig()
		assert.ErrorContains(t, err, ErrUnsupportFramework.Error())
	})

	t.Run("test top_k", func(t *testing.T) {
		infer := &Inference{
			Model:       "qwen",
			Framework:   "Ollama",
			System:      "",
			Template:    "",
			TopK:        0,
			TopP:        0.9,
			Temperature: 0.8,
			NumPredict:  128,
			NumCtx:      2048,
		}
		err := infer.ValidateConfig()
		assert.ErrorContains(t, err, ErrRangeTopK.Error())
	})

	t.Run("test top_p", func(t *testing.T) {
		infer := &Inference{
			Model:       "qwen",
			Framework:   "Ollama",
			System:      "",
			Template:    "",
			TopK:        40,
			TopP:        2,
			Temperature: 0.8,
			NumPredict:  128,
			NumCtx:      2048,
		}
		err := infer.ValidateConfig()
		assert.ErrorContains(t, err, ErrRangeTopP.Error())
	})

	t.Run("test temperature", func(t *testing.T) {
		infer := &Inference{
			Model:       "qwen",
			Framework:   "Ollama",
			System:      "",
			Template:    "",
			TopK:        40,
			TopP:        0.9,
			Temperature: 0,
			NumPredict:  128,
			NumCtx:      2048,
		}
		err := infer.ValidateConfig()
		assert.ErrorContains(t, err, ErrRangeTemperature.Error())
	})

	t.Run("test num_predict", func(t *testing.T) {
		infer := &Inference{
			Model:       "qwen",
			Framework:   "Ollama",
			System:      "",
			Template:    "",
			TopK:        40,
			TopP:        0.9,
			Temperature: 0.8,
			NumPredict:  -100,
			NumCtx:      2048,
		}
		err := infer.ValidateConfig()
		assert.ErrorContains(t, err, ErrRangeNumPredict.Error())
	})

	t.Run("test num_ctx", func(t *testing.T) {
		infer := &Inference{
			Model:       "qwen",
			Framework:   "Ollama",
			System:      "",
			Template:    "",
			TopK:        40,
			TopP:        0.9,
			Temperature: 0.8,
			NumPredict:  128,
			NumCtx:      -100,
		}
		err := infer.ValidateConfig()
		assert.ErrorContains(t, err, ErrRangeNumCtx.Error())
	})
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
