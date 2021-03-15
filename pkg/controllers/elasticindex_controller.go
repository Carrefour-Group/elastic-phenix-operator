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
	elasticv1alpha1 "github.com/Carrefour-Group/elastic-phenix-operator/pkg/api/v1alpha1"
	"github.com/Carrefour-Group/elastic-phenix-operator/pkg/utils"
	"fmt"
	"github.com/go-logr/logr"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

// ElasticIndexReconciler reconciles a ElasticIndex object
type ElasticIndexReconciler struct {
	client.Client
	Log                   logr.Logger
	Scheme                *runtime.Scheme
	NamespacesRegexFilter string
}

// +kubebuilder:rbac:groups=elastic.carrefour.com,resources=elasticindices,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=elastic.carrefour.com,resources=elasticindices/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=elastic.carrefour.com,resources=elasticindices/finalizers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=secrets,verbs=get;list;watch
// +kubebuilder:rbac:groups=core,resources=configmaps,verbs=get;list;watch

func (r *ElasticIndexReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("elasticindex", req.NamespacedName)

	log.Info("new reconciliation request")

	var elasticIndex elasticv1alpha1.ElasticIndex
	if err := r.Get(ctx, req.NamespacedName, &elasticIndex); err != nil {
		if apierrors.IsNotFound(err) {
			log.Info("ElasticIndex not found")
		} else {
			log.Error(err, "unable to fetch elasticIndex object")
		}
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	var elasticsearch *utils.Elasticsearch
	if esConfig, err := utils.BuildEsConfigFromSecretSelector(elasticIndex.ObjectMeta.Namespace, elasticIndex.Spec.ElasticURI.SecretKeyRef, r.Client); err != nil {
		log.Error(err, "unable to build EsConfig from a secret")
		if indexStatusUpdated(&elasticIndex.Status, &utils.EsStatus{Status: utils.StatusError, Message: err.Error()}, log) {
			r.Status().Update(ctx, &elasticIndex)
		}
		return ctrl.Result{RequeueAfter: ErrorInterval}, nil
	} else {
		log.Info("esConfig generated from secret", "EsConfig", esConfig)
		if elasticsearch, err = (utils.Elasticsearch{}).NewClient(esConfig, log); err != nil {
			return ctrl.Result{}, err
		}
	}

	if deleteRequest, err := manageIndexFinalizer(ctx, elasticIndex, elasticsearch, log, r); err != nil {
		return ctrl.Result{}, err
	} else if !deleteRequest {
		if err := elasticsearch.PingES(ctx); err != nil {
			if indexStatusUpdated(&elasticIndex.Status, &utils.EsStatus{Status: utils.StatusRetry, Message: err.Error()}, log) {
				r.Status().Update(ctx, &elasticIndex)
			}
			return ctrl.Result{RequeueAfter: RetryInterval}, nil
		}
		log.Info("create/update ElasticIndex", "indexName", elasticIndex.Spec.IndexName)
		esStatus, err := elasticsearch.CreateOrUpdateIndex(ctx, *elasticIndex.Spec.IndexName, *elasticIndex.Spec.Model)
		if indexStatusUpdated(&elasticIndex.Status, esStatus, log) {
			if err := r.Status().Update(ctx, &elasticIndex); err != nil {
				if apierrors.IsConflict(err) {
					log.Info("conflict: operation cannot be fulfilled on ElasticIndex. Requeue to try again")
					return ctrl.Result{Requeue: true}, nil
				} else {
					log.Error(err, "unable to update ElasticIndex status")
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

	if elasticIndex.Status.Status == utils.StatusRetry {
		return ctrl.Result{RequeueAfter: RetryInterval}, nil
	} else {
		return ctrl.Result{}, nil
	}
}

func (r *ElasticIndexReconciler) SetupWithManager(mgr ctrl.Manager) error {
	if r.NamespacesRegexFilter == "" {
		return ctrl.NewControllerManagedBy(mgr).
			For(&elasticv1alpha1.ElasticIndex{}).
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
			For(&elasticv1alpha1.ElasticIndex{}).
			WithEventFilter(namespacesRegexFilter).
			Complete(r)
	}
}

func indexStatusUpdated(objectStatus *elasticv1alpha1.ElasticIndexStatus, esStatus *utils.EsStatus, log logr.Logger) bool {
	if esStatus != nil &&
		(objectStatus.Status != esStatus.Status ||
			(objectStatus.Status != utils.StatusCreated && objectStatus.Message != esStatus.Message)) {
		log.Info("update status", "from", objectStatus.Status, "to", esStatus.Status)
		objectStatus.Status = esStatus.Status
		objectStatus.HttpCodeStatus = esStatus.HttpCodeStatus
		objectStatus.Message = esStatus.Message
		return true
	}

	return false
}

func manageIndexFinalizer(ctx context.Context, elasticIndex elasticv1alpha1.ElasticIndex, elasticsearch *utils.Elasticsearch, log logr.Logger, r *ElasticIndexReconciler) (bool, error) {
	finalizerName := fmt.Sprintf("finalizer.%v", elasticv1alpha1.GroupVersion.Group)
	deleteRequest := false

	if elasticIndex.ObjectMeta.DeletionTimestamp.IsZero() {
		if !utils.ContainsString(elasticIndex.ObjectMeta.Finalizers, finalizerName) {
			log.Info("register a finalizer")
			elasticIndex.ObjectMeta.Finalizers = append(elasticIndex.ObjectMeta.Finalizers, finalizerName)
			if err := r.Update(ctx, &elasticIndex); err != nil {
				return deleteRequest, err
			}
		}
	} else {
		log.Info("elasticindex is being deleted")
		deleteRequest = true
		if utils.ContainsString(elasticIndex.ObjectMeta.Finalizers, finalizerName) {
			if elasticIndex.Annotations[DeleteInClusterAnnotation] == "true" {
				if err := elasticsearch.DeleteIndex(ctx, *elasticIndex.Spec.IndexName); err != nil {
					log.Error(err, "error while deleting elasticIndex", "indexName", *elasticIndex.Spec.IndexName)
				}
			} else {
				log.Info("elasticindex deletion will not delete elasticsearch index", "indexName", *elasticIndex.Spec.IndexName)
			}

			// remove finalizer from the list and update it.
			elasticIndex.ObjectMeta.Finalizers = utils.RemoveString(elasticIndex.ObjectMeta.Finalizers, finalizerName)
			if err := r.Update(ctx, &elasticIndex); err != nil {
				return deleteRequest, err
			}
		}
	}
	return deleteRequest, nil
}
