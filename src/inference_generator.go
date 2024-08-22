package main

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"gopkg.in/yaml.v2"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"kusionstack.io/kusion-module-framework/pkg/module"
	"kusionstack.io/kusion-module-framework/pkg/server"
	apiv1 "kusionstack.io/kusion/pkg/apis/api.kusion.io/v1"
	"kusionstack.io/kusion/pkg/log"
	"kusionstack.io/kusion/pkg/modules"
)

func main() {
	server.Start(&Inference{})
}

// Inference implements the Kusion Module generator interface.
type Inference struct {
	Model       string  `yaml:"model,omitempty" json:"model,omitempty"`
	Framework   string  `yaml:"framework,omitempty" json:"framework,omitempty"`
	System      string  `yaml:"system,omitempty" json:"system,omitempty"`
	Template    string  `yaml:"template,omitempty" json:"template,omitempty"`
	TopK        int     `yaml:"top_k,omitempty" json:"top_k,omitempty"`
	TopP        float64 `yaml:"top_p,omitempty" json:"top_p,omitempty"`
	Temperature float64 `yaml:"temperature,omitempty" json:"temperature,omitempty"`
	NumPredict  int     `yaml:"num_predict,omitempty" json:"num_predict,omitempty"`
	NumCtx      int     `yaml:"num_ctx,omitempty" json:"num_ctx,omitempty"`
}

func (infer *Inference) Generate(_ context.Context, request *module.GeneratorRequest) (*module.GeneratorResponse, error) {
	defer func() {
		if r := recover(); r != nil {
			log.Debugf("failed to generate inference module: %v", r)
		}
	}()

	// Inference module does not exist in AppConfiguration configs.
	if request.DevConfig == nil {
		log.Info("Inference does not exist in AppConfig config")
		return nil, nil
	}

	/*
		向Inference对象填充配置
		生成资源
	*/
	// Get the complete inference module configs.
	if err := infer.CompleteConfig(request.DevConfig, request.PlatformConfig); err != nil {
		log.Debugf("failed to get complete inference module configs: %v", err)
		return nil, err
	}

	// Validate the completed inference module configs.
	if err := infer.ValidateConfig(); err != nil {
		log.Debugf("failed to validate the inference module configs: %v", err)
		return nil, err
	}

	var resources []apiv1.Resource
	var patcher *apiv1.Patcher

	// Generate the Kubernetes Service related resource.
	resource, patcher, err := infer.GenerateInferenceResource(request)
	if err != nil {
		return nil, err
	}
	resources = append(resources, *resource)

	// Return the Kusion generator response.
	return &module.GeneratorResponse{
		Resources: resources,
		Patcher:   patcher,
	}, nil
}

// CompleteConfig completes the inference module configs with both devModuleConfig and platformModuleConfig.
func (infer *Inference) CompleteConfig(devConfig apiv1.Accessory, platformConfig apiv1.GenericConfig) error {
	// Retrieve the config items the developers are concerned about.
	if devConfig != nil {
		devCfgYamlStr, err := yaml.Marshal(devConfig)
		if err != nil {
			return err
		}

		if err = yaml.Unmarshal(devCfgYamlStr, infer); err != nil {
			return err
		}
	}
	// Retrieve the config items the platform engineers care about.
	if platformConfig != nil {
		platformCfgYamlStr, err := yaml.Marshal(platformConfig)
		if err != nil {
			return err
		}

		if err = yaml.Unmarshal(platformCfgYamlStr, infer); err != nil {
			return err
		}
	}
	return nil
}

// ValidateConfig validates the completed inference configs are valid or not.
func (infer *Inference) ValidateConfig() error {
	if infer.Framework != "Ollama" && infer.Framework != "KubeRay" {
		return errors.New("framework must be Ollama or KubeRay")
	}
	if infer.TopK <= 0 {
		return errors.New("topK must be greater than 0 if exist")
	}
	if infer.TopP <= 0 || infer.TopP > 1 {
		return errors.New("topP must be greater than 0 and less than or equal to 1 if exist")
	}
	if infer.Temperature <= 0 {
		return errors.New("temperature must be greater than 0 if exist")
	}
	if infer.NumPredict < -2 {
		return errors.New("numPredict must be greater than or equal to -2")
	}
	if infer.NumCtx <= 0 {
		return errors.New("numCtx must be greater than 0 if exist")
	}
	return nil
}

// GenerateInferenceResource generates the Kubernetes Service related to the inference module service.
//
// Note that we will use the SDK provided by the kusion module framework to wrap the Kubernetes resource
// into Kusion resource.
func (infer *Inference) GenerateInferenceResource(request *module.GeneratorRequest) (*apiv1.Resource, *apiv1.Patcher, error) {
	/*
		生成 Ollama Deployment 资源
		生成 Ollama Service 资源
		Patcher 中配置 Ollama 服务路径
	*/

	appUniqueName := "ollama-" + modules.UniqueAppName(request.Project, request.Stack, request.App)
	// Deployment: Container Volume
	deployment := &appsv1.Deployment{
		TypeMeta:   typeMeta,
		ObjectMeta: objectMeta,
		Spec:       spec,
	}

	// Service
	svc := &v1.Service{
		TypeMeta: metav1.TypeMeta{
			APIVersion: v1.SchemeGroupVersion.String(),
			Kind:       "Service",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      appUniqueName,
			Namespace: request.Project,
		},
		Spec: v1.ServiceSpec{
			Ports: []v1.ServicePort{
				{
					Name: fmt.Sprintf("%s-%d-%s",
						appUniqueName, infer.Service.Port, strings.ToLower(infer.Service.Protocol)),
					Port:       int32(infer.Service.Port),
					TargetPort: intstr.FromInt(infer.Service.TargetPort),
					Protocol:   v1.Protocol(infer.Service.Protocol),
				},
			},
			Selector: selector,
			Type:     svcType,
		},
	}

	// Patch
	envVars := []v1.EnvVar{
		{
			Name:  "INFERENCE_PATH",
			Value: "",
		},
	}
	patcher := &apiv1.Patcher{
		Environments: envVars,
	}

	// Generate the unique application name with project, stack and app name.
	appUniqueName := modules.UniqueAppName(request.Project, request.Stack, request.App)
	svcType := v1.ServiceTypeClusterIP

	// Generate the selector for the Service workload with the unique app labels SDK
	// provided by Kusion.
	selector := modules.UniqueAppLabels(request.Project, request.App)

	// Add the labels and annotations in inference module to the Service.
	if len(svc.Labels) == 0 {
		svc.Labels = make(map[string]string)
	}
	if len(svc.Annotations) == 0 {
		svc.Annotations = make(map[string]string)
	}

	for k, v := range infer.Service.Labels {
		svc.Labels[k] = v
	}
	for k, v := range infer.Service.Annotations {
		svc.Annotations[k] = v
	}

	// Generate Kusion resource ID and wrap the Kubernetes Service into Kusion resource
	// with the SDK provided by kusion module framework.
	resourceID := module.KubernetesResourceID(svc.TypeMeta, svc.ObjectMeta)
	resource, err := module.WrapK8sResourceToKusionResource(resourceID, svc)
	if err != nil {
		return nil, nil, err
	}

	return resource, patcher, nil
}
