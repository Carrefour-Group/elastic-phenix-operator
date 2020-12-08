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

package controllers

import (
	"context"
	"github.com/Carrefour-Group/elastic-phenix-operator/pkg/utils"
	"fmt"
	"github.com/go-logr/logr"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	elasticv1alpha1 "github.com/Carrefour-Group/elastic-phenix-operator/pkg/api/v1alpha1"
)

// ElasticTemplateReconciler reconciles a ElasticTemplate object
type ElasticTemplateReconciler struct {
	client.Client
	Log                   logr.Logger
	Scheme                *runtime.Scheme
	NamespacesRegexFilter string
	EnableDelete          bool
}

// +kubebuilder:rbac:groups=elastic.carrefour.com,resources=elastictemplates,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=elastic.carrefour.com,resources=elastictemplates/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=elastic.carrefour.com,resources=elastictemplates/finalizers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=secrets,verbs=get;list;watch
// +kubebuilder:rbac:groups=core,resources=configmaps,verbs=get;list;watch

func (r *ElasticTemplateReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {

	ctx := context.Background()
	//log := r.Log.WithValues("name", req.Name, "namespace", req.Namespace)
	log := r.Log.WithValues("elastictemplate", req.NamespacedName)

	log.Info("new reconciliation request")

	var elasticTemplate elasticv1alpha1.ElasticTemplate
	if err := r.Get(ctx, req.NamespacedName, &elasticTemplate); err != nil {
		if apierrors.IsNotFound(err) {
			log.Info("ElasticTemplate not found")
		} else {
			log.Error(err, "unable to fetch elasticTemplate object")
		}
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	var elasticsearch *utils.Elasticsearch
	if esConfig, err := utils.BuildEsConfigFromSecretSelector(elasticTemplate.ObjectMeta.Namespace, elasticTemplate.Spec.ElasticURI.SecretKeyRef, r.Client); err != nil {
		log.Error(err, "unable to build EsConfig from a secret")
		if templateStatusUpdated(&elasticTemplate.Status, &utils.EsStatus{Status: utils.StatusError, Message: err.Error()}, log) {
			r.Status().Update(ctx, &elasticTemplate)
		}
		return ctrl.Result{RequeueAfter: ErrorInterval}, nil
	} else {
		log.Info("esConfig generated from secret", "EsConfig", esConfig)
		if elasticsearch, err = (utils.Elasticsearch{}).NewClient(esConfig, log); err != nil {
			return ctrl.Result{}, err
		}
	}

	if deleteRequest, err := manageTemplateFinalizer(ctx, elasticTemplate, elasticsearch, log, r); err != nil {
		return ctrl.Result{}, err
	} else if !deleteRequest {
		if err := elasticsearch.PingES(ctx); err != nil {
			if templateStatusUpdated(&elasticTemplate.Status, &utils.EsStatus{Status: utils.StatusRetry, Message: err.Error()}, log) {
				if err := r.Status().Update(ctx, &elasticTemplate); err != nil {
					return ctrl.Result{RequeueAfter: RetryInterval}, nil
				}
			}
			return ctrl.Result{RequeueAfter: RetryInterval}, nil
		}
		log.Info("create/update ElasticTemplate", "templateName", elasticTemplate.Spec.TemplateName)
		esStatus, err := elasticsearch.CreateOrUpdateTemplate(ctx, *elasticTemplate.Spec.TemplateName, *elasticTemplate.Spec.Model)
		if templateStatusUpdated(&elasticTemplate.Status, esStatus, log) {
			if err := r.Status().Update(ctx, &elasticTemplate); err != nil {
				if apierrors.IsConflict(err) {
					log.Info("conflict: operation cannot be fulfilled on ElasticTemplate. Requeue to try again")
					return ctrl.Result{Requeue: true}, nil
				} else {
					log.Error(err, "unable to update ElasticTemplate status")
					return ctrl.Result{}, err
				}
			}
		}
		if esStatus.Status == utils.StatusError {
			//blocking error no need to Requeue or Requeue after a long interval
			return ctrl.Result{RequeueAfter: ErrorInterval}, nil
		} else if err != nil {
			return ctrl.Result{}, err
		}
	}

	if elasticTemplate.Status.Status == utils.StatusRetry {
		return ctrl.Result{RequeueAfter: RetryInterval}, nil
	} else {
		return ctrl.Result{}, nil
	}
}

func (r *ElasticTemplateReconciler) SetupWithManager(mgr ctrl.Manager) error {
	if r.NamespacesRegexFilter == "" {
		return ctrl.NewControllerManagedBy(mgr).
			For(&elasticv1alpha1.ElasticTemplate{}).
			Complete(r)
	} else {
		var regex = r.NamespacesRegexFilter
		namespacesRegexFilter := predicate.Funcs{
			CreateFunc:  func(ce event.CreateEvent) bool { return utils.FilterByNamespacesRegex(ce.Meta, regex, r.Log) },
			DeleteFunc:  func(ce event.DeleteEvent) bool { return utils.FilterByNamespacesRegex(ce.Meta, regex, r.Log) },
			UpdateFunc:  func(ce event.UpdateEvent) bool { return utils.FilterByNamespacesRegex(ce.MetaNew, regex, r.Log) },
			GenericFunc: func(ce event.GenericEvent) bool { return utils.FilterByNamespacesRegex(ce.Meta, regex, r.Log) },
		}

		return ctrl.NewControllerManagedBy(mgr).
			For(&elasticv1alpha1.ElasticTemplate{}).
			WithEventFilter(namespacesRegexFilter).
			Complete(r)
	}
}

func templateStatusUpdated(objectStatus *elasticv1alpha1.ElasticTemplateStatus, esStatus *utils.EsStatus, log logr.Logger) bool {
	if esStatus != nil && objectStatus.Status != esStatus.Status {
		log.Info("update status", "from", objectStatus.Status, "to", esStatus.Status)
		objectStatus.Status = esStatus.Status
		objectStatus.HttpCodeStatus = esStatus.HttpCodeStatus
		objectStatus.Message = esStatus.Message
		return true
	}

	return false
}

func manageTemplateFinalizer(ctx context.Context, elasticTemplate elasticv1alpha1.ElasticTemplate, elasticsearch *utils.Elasticsearch, log logr.Logger, r *ElasticTemplateReconciler) (bool, error) {
	finalizerName := fmt.Sprintf("finalizer.%v", elasticv1alpha1.GroupVersion.Group)
	deleteRequest := false

	if elasticTemplate.ObjectMeta.DeletionTimestamp.IsZero() {
		if !utils.ContainsString(elasticTemplate.ObjectMeta.Finalizers, finalizerName) {
			log.Info("register a finalizer")
			elasticTemplate.ObjectMeta.Finalizers = append(elasticTemplate.ObjectMeta.Finalizers, finalizerName)
			if err := r.Update(ctx, &elasticTemplate); err != nil {
				return deleteRequest, err
			}
		}
	} else {
		log.Info("elastictemplate is being deleted")
		deleteRequest = true
		if utils.ContainsString(elasticTemplate.ObjectMeta.Finalizers, finalizerName) {
			if r.EnableDelete {
				if err := elasticsearch.DeleteTemplate(ctx, *elasticTemplate.Spec.TemplateName); err != nil {
					log.Error(err, "error while deleting elasticTemplate", "templateName", *elasticTemplate.Spec.TemplateName)
				}
			} else {
				log.Info("elastictemplate deletion will not delete elasticsearch template", "templateName", *elasticTemplate.Spec.TemplateName, "enable-delete flag", r.EnableDelete)
			}

			// remove finalizer from the list and update it.
			elasticTemplate.ObjectMeta.Finalizers = utils.RemoveString(elasticTemplate.ObjectMeta.Finalizers, finalizerName)
			if err := r.Update(ctx, &elasticTemplate); err != nil {
				return deleteRequest, err
			}
		}
	}
	return deleteRequest, nil
}
