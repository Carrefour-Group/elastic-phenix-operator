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

package utils

import (
	"context"
	"github.com/go-logr/logr"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"regexp"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strings"
)

func FilterByNamespacesRegex(meta metav1.Object, regex string, log logr.Logger) bool {
	match, _ := regexp.MatchString(regex, meta.GetNamespace())
	if !match {
		log.Info("/!\\ event filtered", "namespace", meta.GetNamespace(), "name", meta.GetName())
	}
	return match
}

func GetSecret(namespace string, secretKeySelector *v1.SecretKeySelector, k8sClient client.Client) (*v1.Secret, error) {
	var secret v1.Secret
	name := strings.TrimSpace(secretKeySelector.Name)
	if err := k8sClient.Get(context.Background(), client.ObjectKey{Namespace: namespace, Name: name}, &secret); err != nil {
		return nil, err
	}
	return &secret, nil
}

func BuildEsConfigFromSecretSelector(namespace string, secretKeySelector *v1.SecretKeySelector, k8sClient client.Client) (*EsConfig, error) {
	if secret, err := GetSecret(namespace, secretKeySelector, k8sClient); err != nil {
		return nil, err
	} else {
		return BuildEsConfigFromExistingSecret(secret, secretKeySelector.Key)
	}
}

func BuildEsConfigFromExistingSecret(secret *v1.Secret, key string) (*EsConfig, error) {
	elasticURI := string(secret.Data[key])
	if esConfig, err := (&EsConfig{}).FromURI(elasticURI); err != nil {
		return nil, err
	} else {
		return esConfig, nil
	}
}
