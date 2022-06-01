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
	"fmt"
	"github.com/Carrefour-Group/elastic-phenix-operator/pkg/utils"
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
	elasticindexlog        = logf.Log.WithName("elasticindex-resource")
	elasticindexK8sClient  client.Client
	elasticindexNamespaces []string
)

func (r *ElasticIndex) SetupWebhookWithManager(mgr ctrl.Manager, namespaces []string) error {
	elasticindexK8sClient = mgr.GetClient()
	elasticindexNamespaces = namespaces

	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

// +kubebuilder:webhook:path=/mutate-elastic-carrefour-com-v1alpha1-elasticindex,mutating=true,failurePolicy=fail,groups=elastic.carrefour.com,resources=elasticindices,verbs=create;update,versions=v1alpha1,name=melasticindex.kb.io,sideEffects=none,admissionReviewVersions=v1beta1

var _ webhook.Defaulter = &ElasticIndex{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *ElasticIndex) Default() {
	elasticindexlog.Info("[Webhook] default", "namespace", r.Namespace, "name", r.Name)

	if r.Spec.Model != nil && r.Spec.NumberOfReplicas != nil && r.Spec.NumberOfShards != nil {
		//add settings to body
		modelWithSettings, _ := (&utils.EsModel{Model: *r.Spec.Model}).AddSettings(*r.Spec.NumberOfReplicas, *r.Spec.NumberOfShards)
		if compactedModel, err := utils.CompactJson(modelWithSettings); err == nil {
			*r.Spec.Model = compactedModel
		}
	}
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
// +kubebuilder:webhook:verbs=create;update;delete,path=/validate-elastic-carrefour-com-v1alpha1-elasticindex,mutating=false,failurePolicy=fail,groups=elastic.carrefour.com,resources=elasticindices,versions=v1alpha1,name=velasticindex.kb.io,sideEffects=none,admissionReviewVersions=v1beta1

var _ webhook.Validator = &ElasticIndex{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *ElasticIndex) ValidateCreate() error {
	if len(elasticindexNamespaces) == 0 || utils.ContainsString(elasticindexNamespaces, r.ObjectMeta.Namespace) {
		elasticindexlog.Info("[Webhook] validate create", "namespace", r.Namespace, "name", r.Name)

		var allErrs field.ErrorList

		_, err := (&utils.EsModel{Model: *r.Spec.Model}).IsValid(utils.Index)
		if err != nil {
			allErrs = append(allErrs, field.Invalid(field.NewPath("spec").Child("model"), r.Spec.Model, err.Error()))
		}

		if secret, err := utils.GetSecret(r.ObjectMeta.Namespace, r.Spec.ElasticURI.SecretKeyRef, elasticindexK8sClient); err != nil {
			errMsg := fmt.Sprintf(`secret "%v" is required. %v`, r.Spec.ElasticURI.SecretKeyRef.Name, err.Error())
			allErrs = append(allErrs, field.Required(field.NewPath("spec").Child("elasticUri").Child("secretKeyRef"), errMsg))
		} else {
			secretKey := r.Spec.ElasticURI.SecretKeyRef.Key
			if esConfig, err := utils.BuildEsConfigFromExistingSecret(secret, secretKey); err != nil {
				errMsg := fmt.Sprintf(`error while parsing elasticsearch URI "%v" from secret. %v`, r.Spec.ElasticURI.SecretKeyRef.Name, err.Error())
				allErrs = append(allErrs, field.Invalid(field.NewPath("spec").Child("elasticUri").Child("secretKeyRef"), string(secret.Data[secretKey]), errMsg))
			} else {
				if info, err := checkEsIndexExists(*r.Spec.IndexName, esConfig, elasticindexK8sClient); err != nil {
					errMsg := fmt.Sprintf(`error while checking index "%v" existence from all kubernetes elasticindex objects. %v`, r.Spec.IndexName, err.Error())
					allErrs = append(allErrs, field.Invalid(field.NewPath("spec").Child("indexName"), r.Spec.IndexName, errMsg))
				} else if info != nil {
					errMsg := fmt.Sprintf(`index "%v" for elasticsearch URI "%v:%v" was created by kubernetes elasticindex "%v" in namespace "%v"`, info.esObjectName, info.Host, info.Port, info.Name, info.Namespace)
					allErrs = append(allErrs, field.Forbidden(field.NewPath("spec").Child("indexName"), errMsg))
				}
			}
		}

		if len(allErrs) == 0 {
			return nil
		}

		return apierrors.NewInvalid(
			schema.GroupKind{Group: "elastic.carrefour.com", Kind: "ElasticIndex"},
			r.Name, allErrs)
	}

	elasticindexlog.Info("[Webhook] ignore validate create", "namespace", r.Namespace, "name", r.Name)
	return nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *ElasticIndex) ValidateUpdate(old runtime.Object) error {
	if len(elasticindexNamespaces) == 0 || utils.ContainsString(elasticindexNamespaces, r.ObjectMeta.Namespace) {
		elasticindexlog.Info("[Webhook] validate update", "namespace", r.Namespace, "name", r.Name)

		var allErrs field.ErrorList

		_, err := (&utils.EsModel{Model: *r.Spec.Model}).IsValid(utils.Index)
		if err != nil {
			allErrs = append(allErrs, field.Invalid(field.NewPath("spec").Child("model"), r.Spec.Model, err.Error()))
		}

		oldR := old.(*ElasticIndex)

		if *r.Spec.IndexName != *oldR.Spec.IndexName {
			errMsg := fmt.Sprintf(`Cannot update indexName from "%v" to "%v"`, *r.Spec.IndexName, *oldR.Spec.IndexName)
			allErrs = append(allErrs, field.Invalid(field.NewPath("spec").Child("indexName"), r.Spec.IndexName, errMsg))
		}

		if *r.Spec.NumberOfShards != *oldR.Spec.NumberOfShards {
			allErrs = append(allErrs, field.Forbidden(field.NewPath("spec").Child("numberOfShards"), "cannot update numberOfShards setting for an Index"))
		}

		allErrs = ValidateUpdateSecret(allErrs, r.Namespace, r.Spec.ElasticURI.SecretKeyRef, oldR.Spec.ElasticURI.SecretKeyRef, elasticindexK8sClient)

		if len(allErrs) == 0 {
			return nil
		}

		return apierrors.NewInvalid(
			schema.GroupKind{Group: "elastic.carrefour.com", Kind: "ElasticIndex"},
			r.Name, allErrs)
	}

	elasticindexlog.Info("[Webhook] ignore validate update", "namespace", r.Namespace, "name", r.Name)
	return nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *ElasticIndex) ValidateDelete() error {
	if len(elasticindexNamespaces) == 0 || utils.ContainsString(elasticindexNamespaces, r.ObjectMeta.Namespace) {
		elasticindexlog.Info("[Webhook] validate delete", "namespace", r.Namespace, "name", r.Name)

		var allErrs field.ErrorList

		if _, err := utils.GetSecret(r.ObjectMeta.Namespace, r.Spec.ElasticURI.SecretKeyRef, elasticindexK8sClient); err != nil {
			errMsg := fmt.Sprintf(`secret "%v" is required. %v`, r.Spec.ElasticURI.SecretKeyRef.Name, err.Error())
			allErrs = append(allErrs, field.Required(field.NewPath("spec").Child("elasticUri").Child("secretKeyRef"), errMsg))
		}

		if len(allErrs) == 0 {
			return nil
		}

		return apierrors.NewInvalid(
			schema.GroupKind{Group: "elastic.carrefour.com", Kind: "ElasticIndex"},
			r.Name, allErrs)
	}

	elasticindexlog.Info("[Webhook] ignore validate delete", "namespace", r.Namespace, "name", r.Name)
	return nil
}

func checkEsIndexExists(indexName string, esConfig *utils.EsConfig, k8sClient client.Client) (*EsObjectInfo, error) {
	var allElasticIndex ElasticIndexList
	if err := k8sClient.List(context.Background(), &allElasticIndex); err != nil {
		return nil, err
	}
	for _, es := range allElasticIndex.Items {
		esConfigToCheck, _ := utils.BuildEsConfigFromSecretSelector(es.Namespace, es.Spec.ElasticURI.SecretKeyRef, k8sClient)
		if indexName == *es.Spec.IndexName && esConfig.Host == esConfigToCheck.Host && esConfig.Port == esConfigToCheck.Port {
			return &EsObjectInfo{
				Namespace:    es.Namespace,
				Name:         es.Name,
				esObjectName: *es.Spec.IndexName,
				Host:         esConfigToCheck.Host,
				Port:         esConfigToCheck.Port}, nil
		}
	}
	return nil, nil
}
