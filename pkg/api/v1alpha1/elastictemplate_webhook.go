/*
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
	"github.com/Carrefour-Group/elastic-phenix-operator/pkg/utils"
	"fmt"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/validation/field"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

var (
	// log is for logging in this package.
	elastictemplatelog        = logf.Log.WithName("elastictemplate-resource")
	elastictemplateK8sClient  client.Client
	elastictemplateNamespaces []string
)

func (r *ElasticTemplate) SetupWebhookWithManager(mgr ctrl.Manager, namespaces []string) error {
	elastictemplateK8sClient = mgr.GetClient()
	elastictemplateNamespaces = namespaces

	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

// +kubebuilder:webhook:path=/mutate-elastic-carrefour-com-v1alpha1-elastictemplate,mutating=true,failurePolicy=fail,groups=elastic.carrefour.com,resources=elastictemplates,verbs=create;update,versions=v1alpha1,name=melastictemplate.kb.io

var _ webhook.Defaulter = &ElasticTemplate{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *ElasticTemplate) Default() {
	elastictemplatelog.Info("[Webhook] default", "namespace", r.Namespace, "name", r.Name)

	if r.Spec.Model != nil && r.Spec.NumberOfReplicas != nil && r.Spec.NumberOfShards != nil {
		//add settings to body
		modelWithSettings, _ := (&utils.EsModel{Model: *r.Spec.Model}).AddSettings(*r.Spec.NumberOfReplicas, *r.Spec.NumberOfShards)
		if compactedModel, err := utils.CompactJson(modelWithSettings); err == nil {
			*r.Spec.Model = compactedModel
		}
	}
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
// +kubebuilder:webhook:verbs=create;update;delete,path=/validate-elastic-carrefour-com-v1alpha1-elastictemplate,mutating=false,failurePolicy=fail,groups=elastic.carrefour.com,resources=elastictemplates,versions=v1alpha1,name=velastictemplate.kb.io

var _ webhook.Validator = &ElasticTemplate{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *ElasticTemplate) ValidateCreate() error {
	if len(elastictemplateNamespaces) == 0 || utils.ContainsString(elastictemplateNamespaces, r.ObjectMeta.Namespace) {
		elastictemplatelog.Info("[Webhook] validate create", "namespace", r.Namespace, "name", r.Name)

		var allErrs field.ErrorList

		_, err := (&utils.EsModel{Model: *r.Spec.Model}).IsValid(utils.Template)
		if err != nil {
			allErrs = append(allErrs, field.Invalid(field.NewPath("spec").Child("model"), r.Spec.Model, err.Error()))
		}

		if secret, err := utils.GetSecret(r.ObjectMeta.Namespace, r.Spec.ElasticURI.SecretKeyRef, elastictemplateK8sClient); err != nil {
			errMsg := fmt.Sprintf(`secret "%v" is required. %v`, r.Spec.ElasticURI.SecretKeyRef.Name, err.Error())
			allErrs = append(allErrs, field.Required(field.NewPath("spec").Child("elasticUri").Child("secretKeyRef"), errMsg))
		} else {
			secretKey := r.Spec.ElasticURI.SecretKeyRef.Key
			if esConfig, err := utils.BuildEsConfigFromExistingSecret(secret, secretKey); err != nil {
				errMsg := fmt.Sprintf(`error while parsing elasticsearch URI from secret "%v". %v`, r.Spec.ElasticURI.SecretKeyRef.Name, err.Error())
				allErrs = append(allErrs, field.Invalid(field.NewPath("spec").Child("elasticUri").Child("secretKeyRef"), string(secret.Data[secretKey]), errMsg))
			} else {
				if info, err := checkEsTemplateExists(*r.Spec.TemplateName, esConfig, elastictemplateK8sClient); err != nil {
					errMsg := fmt.Sprintf(`error while checking template "%v" existance from all kubernetes elastictemplate objects. %v`, r.Spec.TemplateName, err.Error())
					allErrs = append(allErrs, field.Invalid(field.NewPath("spec").Child("templateName"), r.Spec.TemplateName, errMsg))
				} else if info != nil {
					errMsg := fmt.Sprintf(`template "%v" for elasticsearch URI "%v:%v" was created by kubernetes elastictemplate "%v" in namespace "%v"`, info.esObjectName, info.Host, info.Port, info.Name, info.Namespace)
					allErrs = append(allErrs, field.Forbidden(field.NewPath("spec").Child("templateName"), errMsg))
				}
			}
		}

		if len(allErrs) == 0 {
			return nil
		}

		return apierrors.NewInvalid(
			schema.GroupKind{Group: "elastic.carrefour.com", Kind: "ElasticTemplate"},
			r.Name, allErrs)
	} else {
		elastictemplatelog.Info("[Webhook] ignore validate create", "namespace", r.Namespace, "name", r.Name)
		return nil
	}
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *ElasticTemplate) ValidateUpdate(old runtime.Object) error {
	if len(elastictemplateNamespaces) == 0 || utils.ContainsString(elastictemplateNamespaces, r.ObjectMeta.Namespace) {
		elastictemplatelog.Info("[Webhook] validate update", "namespace", r.Namespace, "name", r.Name)

		var allErrs field.ErrorList

		oldR := old.(*ElasticTemplate)

		if *r.Spec.TemplateName != *oldR.Spec.TemplateName {
			errMsg := fmt.Sprintf(`Cannot update templateName from "%v" to "%v"`, *r.Spec.TemplateName, *oldR.Spec.TemplateName)
			allErrs = append(allErrs, field.Invalid(field.NewPath("spec").Child("templateName"), r.Spec.TemplateName, errMsg))
		}

		if !utils.CompareJson(*r.Spec.Model, *oldR.Spec.Model) {
			_, err := (&utils.EsModel{Model: *r.Spec.Model}).IsValid(utils.Template)
			if err != nil {
				allErrs = append(allErrs, field.Invalid(field.NewPath("spec").Child("model"), r.Spec.Model, err.Error()))
			}
		}

		allErrs = ValidateUpdateSecret(allErrs, r.Namespace, r.Spec.ElasticURI.SecretKeyRef, oldR.Spec.ElasticURI.SecretKeyRef, elastictemplateK8sClient)

		if len(allErrs) == 0 {
			return nil
		}

		return apierrors.NewInvalid(
			schema.GroupKind{Group: "elastic.carrefour.com", Kind: "ElasticTemplate"},
			r.Name, allErrs)
	} else {
		elastictemplatelog.Info("[Webhook] ignore validate update", "namespace", r.Namespace, "name", r.Name)
		return nil
	}
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *ElasticTemplate) ValidateDelete() error {
	if len(elastictemplateNamespaces) == 0 || utils.ContainsString(elastictemplateNamespaces, r.ObjectMeta.Namespace) {
		elastictemplatelog.Info("[Webhook] validate delete", "namespace", r.Namespace, "name", r.Name)

		var allErrs field.ErrorList

		if _, err := utils.GetSecret(r.ObjectMeta.Namespace, r.Spec.ElasticURI.SecretKeyRef, elastictemplateK8sClient); err != nil {
			errMsg := fmt.Sprintf(`secret "%v" is required. %v`, r.Spec.ElasticURI.SecretKeyRef.Name, err.Error())
			allErrs = append(allErrs, field.Required(field.NewPath("spec").Child("elasticUri").Child("secretKeyRef"), errMsg))
		}

		if len(allErrs) == 0 {
			return nil
		}

		return apierrors.NewInvalid(
			schema.GroupKind{Group: "elastic.carrefour.com", Kind: "ElasticTemplate"},
			r.Name, allErrs)
	} else {
		elastictemplatelog.Info("[Webhook] ignore validate delete", "namespace", r.Namespace, "name", r.Name)
		return nil
	}
}

func checkEsTemplateExists(templateName string, esConfig *utils.EsConfig, k8sClient client.Client) (*EsObjectInfo, error) {
	var allElasticTemplate ElasticTemplateList
	if err := k8sClient.List(context.Background(), &allElasticTemplate); err != nil {
		return nil, err
	}
	for _, es := range allElasticTemplate.Items {
		esConfigToCheck, _ := utils.BuildEsConfigFromSecretSelector(es.Namespace, es.Spec.ElasticURI.SecretKeyRef, k8sClient)
		if templateName == *es.Spec.TemplateName && esConfig.Host == esConfigToCheck.Host && esConfig.Port == esConfigToCheck.Port {
			return &EsObjectInfo{
				Namespace:    es.Namespace,
				Name:         es.Name,
				esObjectName: *es.Spec.TemplateName,
				Host:         esConfigToCheck.Host,
				Port:         esConfigToCheck.Port}, nil
		}
	}
	return nil, nil
}
