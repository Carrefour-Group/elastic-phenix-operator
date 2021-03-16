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
	"fmt"
	"github.com/Carrefour-Group/elastic-phenix-operator/pkg/utils"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ElasticURISource struct {
	// +kubebuilder:validation:Required
	SecretKeyRef *v1.SecretKeySelector `json:"secretKeyRef" protobuf:"bytes,4,opt,name=secretKeyRef"`
}

type EsObjectInfo struct {
	Namespace    string
	Name         string
	esObjectName string
	Host         string
	Port         string
}

func ValidateUpdateSecret(allErrs field.ErrorList, namespace string, newSecretSelector *v1.SecretKeySelector, oldSecretSelector *v1.SecretKeySelector, k8sClient client.Client) field.ErrorList {
	if secret, err := utils.GetSecret(namespace, newSecretSelector, k8sClient); err != nil {
		errMsg := fmt.Sprintf(`secret "%v" is required. %v`, newSecretSelector.Name, err.Error())
		allErrs = append(allErrs, field.Required(field.NewPath("spec").Child("elasticUri").Child("secretKeyRef"), errMsg))
	} else if secret != nil {
		secretKey := newSecretSelector.Key
		if esConfig, err := utils.BuildEsConfigFromExistingSecret(secret, secretKey); err != nil {
			errMsg := fmt.Sprintf(`error while parsing elasticsearch URI from secret "%v". %v`, newSecretSelector.Name, err.Error())
			allErrs = append(allErrs, field.Invalid(field.NewPath("spec").Child("elasticUri").Child("secretKeyRef"), string(secret.Data[secretKey]), errMsg))
		} else {
			if oldEsConfig, _ := utils.BuildEsConfigFromSecretSelector(namespace, oldSecretSelector, k8sClient); oldEsConfig != nil {
				if esConfig.Host != oldEsConfig.Host || esConfig.Port != oldEsConfig.Port {
					errMsg := fmt.Sprintf(`cannot update elasticsearch host:port from "%v:%v" to "%v:%v"`, oldEsConfig.Host, oldEsConfig.Port, esConfig.Host, esConfig.Port)
					allErrs = append(allErrs, field.Forbidden(field.NewPath("spec").Child("elasticUri").Child("secretKeyRef"), errMsg))
				}
			}
		}
	}
	return allErrs
}
