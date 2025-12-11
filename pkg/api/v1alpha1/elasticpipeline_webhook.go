/*
Copyright 2023.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	"context"
	"fmt"
	"github.com/Carrefour-Group/elastic-phenix-operator/pkg/utils"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/validation/field"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

var (
	// log is for logging in this package.
	elasticpipelinelog        = logf.Log.WithName("elasticpipeline-resource")
	elasticpipelineK8sClient  client.Client
	elasticpipelineNamespaces []string
)

func (r *ElasticPipeline) SetupWebhookWithManager(mgr ctrl.Manager, namespaces []string) error {
	elasticpipelineK8sClient = mgr.GetClient()
	elasticpipelineNamespaces = namespaces
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// +kubebuilder:webhook:path=/mutate-elastic-carrefour-com-v1alpha1-elasticpipeline,mutating=true,failurePolicy=fail,sideEffects=None,groups=elastic.carrefour.com,resources=elasticpipelines,verbs=create;update,versions=v1alpha1,name=melasticpipeline.kb.io,admissionReviewVersions=v1beta1;v1
var _ webhook.Defaulter = &ElasticPipeline{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *ElasticPipeline) Default() {
	elasticpipelinelog.Info("default", "namespace", r.Namespace, "name", r.Name)
}

//+kubebuilder:webhook:verbs=create;update;delete,path=/validate-elastic-carrefour-com-v1alpha1-elasticpipeline,mutating=false,failurePolicy=fail,groups=elastic.carrefour.com,resources=elasticpipelines,versions=v1alpha1,name=velasticpipeline.kb.io,sideEffects=none,admissionReviewVersions=v1beta1;v1

var _ webhook.Validator = &ElasticPipeline{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *ElasticPipeline) ValidateCreate() error {
	if len(elasticpipelineNamespaces) == 0 || utils.ContainsString(elasticpipelineNamespaces, r.ObjectMeta.Namespace) {
		elasticpipelinelog.Info("[Webhook] validate create", "namespace", r.Namespace, "name", r.Name)
		var allErrs field.ErrorList
		_, err := (&utils.EsModel{Model: r.Spec.Model}).IsValid(utils.Pipeline)
		if err != nil {
			allErrs = append(allErrs, field.Invalid(field.NewPath("spec").Child("model"), r.Spec.Model, err.Error()))
		}

		if secret, err := utils.GetSecret(r.ObjectMeta.Namespace, r.Spec.ElasticURI.SecretKeyRef, elasticpipelineK8sClient); err != nil {
			errMsg := fmt.Sprintf(`secret "%v" is required. %v`, r.Spec.ElasticURI.SecretKeyRef.Name, err.Error())
			allErrs = append(allErrs, field.Required(field.NewPath("spec").Child("elasticUri").Child("secretKeyRef"), errMsg))
		} else {
			secretKey := r.Spec.ElasticURI.SecretKeyRef.Key
			if esConfig, err := utils.BuildEsConfigFromExistingSecret(secret, secretKey); err != nil {
				errMsg := fmt.Sprintf(`error while parsing elasticsearch URI from secret "%v". %v`, r.Spec.ElasticURI.SecretKeyRef.Name, err.Error())
				allErrs = append(allErrs, field.Invalid(field.NewPath("spec").Child("elasticUri").Child("secretKeyRef"), string(secret.Data[secretKey]), errMsg))
			} else {
				if info, err := checkEsPipelineExists(r.Spec.PipelineName, esConfig, elasticpipelineK8sClient); err != nil {
					errMsg := fmt.Sprintf(`error while checking pipeline "%v" existence from all kubernetes elasticpipeline objects. %v`, r.Spec.PipelineName, err.Error())
					allErrs = append(allErrs, field.Invalid(field.NewPath("spec").Child("pipelineName"), r.Spec.PipelineName, errMsg))
				} else if info != nil {
					errMsg := fmt.Sprintf(`pipeline "%v" for elasticsearch URI "%v:%v" was created by kubernetes elasticpipeline "%v" in namespace "%v"`, info.esObjectName, info.Host, info.Port, info.Name, info.Namespace)
					allErrs = append(allErrs, field.Forbidden(field.NewPath("spec").Child("pipelineName"), errMsg))
				}
			}
		}

		if len(allErrs) == 0 {
			return nil
		}

		return apierrors.NewInvalid(
			schema.GroupKind{Group: "elastic.carrefour.com", Kind: "ElasticPipeline"},
			r.Name, allErrs)
	}

	elasticpipelinelog.Info("[Webhook] ignore validate create", "namespace", r.Namespace, "name", r.Name)
	return nil

}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *ElasticPipeline) ValidateUpdate(old runtime.Object) error {
	if len(elasticpipelineNamespaces) == 0 || utils.ContainsString(elasticpipelineNamespaces, r.ObjectMeta.Namespace) {
		elasticpipelinelog.Info("[Webhook] validate update", "namespace", r.Namespace, "name", r.Name)

		var allErrs field.ErrorList

		_, err := (&utils.EsModel{Model: r.Spec.Model}).IsValid(utils.Pipeline)
		if err != nil {
			allErrs = append(allErrs, field.Invalid(field.NewPath("spec").Child("model"), r.Spec.Model, err.Error()))
		}

		oldR := old.(*ElasticPipeline)

		if r.Spec.PipelineName != oldR.Spec.PipelineName {
			errMsg := fmt.Sprintf(`Cannot update PipelineName from "%v" to "%v"`, r.Spec.PipelineName, oldR.Spec.PipelineName)
			allErrs = append(allErrs, field.Invalid(field.NewPath("spec").Child("pipelineName"), r.Spec.PipelineName, errMsg))
		}

		allErrs = ValidateUpdateSecret(allErrs, r.Namespace, r.Spec.ElasticURI.SecretKeyRef, oldR.Spec.ElasticURI.SecretKeyRef, elasticpipelineK8sClient)

		if len(allErrs) == 0 {
			return nil
		}

		return apierrors.NewInvalid(
			schema.GroupKind{Group: "elastic.carrefour.com", Kind: "ElasticPipeline"},
			r.Name, allErrs)
	}

	elasticpipelinelog.Info("[Webhook] ignore validate update", "namespace", r.Namespace, "name", r.Name)
	return nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *ElasticPipeline) ValidateDelete() error {
	if len(elasticpipelineNamespaces) == 0 || utils.ContainsString(elasticpipelineNamespaces, r.ObjectMeta.Namespace) {
		elasticpipelinelog.Info("[Webhook] validate delete", "namespace", r.Namespace, "name", r.Name)

		var allErrs field.ErrorList

		if _, err := utils.GetSecret(r.ObjectMeta.Namespace, r.Spec.ElasticURI.SecretKeyRef, elasticpipelineK8sClient); err != nil {
			errMsg := fmt.Sprintf(`secret "%v" is required. %v`, r.Spec.ElasticURI.SecretKeyRef.Name, err.Error())
			allErrs = append(allErrs, field.Required(field.NewPath("spec").Child("elasticUri").Child("secretKeyRef"), errMsg))
		}

		if len(allErrs) == 0 {
			return nil
		}

		return apierrors.NewInvalid(
			schema.GroupKind{Group: "elastic.carrefour.com", Kind: "ElasticPipeline"},
			r.Name, allErrs)
	}

	elasticpipelinelog.Info("[Webhook] ignore validate delete", "namespace", r.Namespace, "name", r.Name)
	return nil
}

func checkEsPipelineExists(pipelineName string, esConfig *utils.EsConfig, k8sClient client.Client) (*EsObjectInfo, error) {
	var allElasticPipeline ElasticPipelineList
	if err := k8sClient.List(context.Background(), &allElasticPipeline); err != nil {
		return nil, err
	}
	for _, es := range allElasticPipeline.Items {
		esConfigToCheck, _ := utils.BuildEsConfigFromSecretSelector(es.Namespace, es.Spec.ElasticURI.SecretKeyRef, k8sClient)
		if pipelineName == es.Spec.PipelineName && esConfig.Host == esConfigToCheck.Host && esConfig.Port == esConfigToCheck.Port {
			return &EsObjectInfo{
				Namespace:    es.Namespace,
				Name:         es.Name,
				esObjectName: es.Spec.PipelineName,
				Host:         esConfigToCheck.Host,
				Port:         esConfigToCheck.Port}, nil
		}
	}
	return nil, nil
}
