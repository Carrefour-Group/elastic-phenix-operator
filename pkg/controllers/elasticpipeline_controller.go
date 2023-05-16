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

package controllers

import (
	"context"
	"fmt"
	elasticv1alpha1 "github.com/Carrefour-Group/elastic-phenix-operator/pkg/api/v1alpha1"
	"github.com/Carrefour-Group/elastic-phenix-operator/pkg/utils"
	"github.com/go-logr/logr"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// ElasticPipelineReconciler reconciles a ElasticPipeline object
type ElasticPipelineReconciler struct {
	client.Client
	Scheme                *runtime.Scheme
	Log                   logr.Logger
	NamespacesRegexFilter string
}

//+kubebuilder:rbac:groups=elastic.carrefour.com,resources=elasticpipelines,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=elastic.carrefour.com,resources=elasticpipelines/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=elastic.carrefour.com,resources=elasticpipelines/finalizers,verbs=update
//+kubebuilder:rbac:groups=core,resources=secrets,verbs=get;list;watch
//+kubebuilder:rbac:groups=core,resources=configmaps,verbs=get;list;watch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the ElasticPipeline object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.14.1/pkg/reconcile
func (r *ElasticPipelineReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()

	log := r.Log.WithValues("elasticpipeline", req.NamespacedName)

	log.Info("new reconciliation request")

	var elasticPipeline elasticv1alpha1.ElasticPipeline
	if err := r.Get(ctx, req.NamespacedName, &elasticPipeline); err != nil {
		if apierrors.IsNotFound(err) {
			log.Info("ElasticPipeline not found")
		} else {
			log.Error(err, "unable to fetch elasticpipeline object")
		}
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	esConfig, err := utils.BuildEsConfigFromSecretSelector(elasticPipeline.ObjectMeta.Namespace, elasticPipeline.Spec.ElasticURI.SecretKeyRef, r.Client)
	if err != nil {
		log.Error(err, "unable to build EsConfig from a secret")
		if pipelineStatusUpdated(&elasticPipeline.Status, &utils.EsStatus{Status: utils.StatusError, Message: err.Error()}, log) {
			err := r.Status().Update(ctx, &elasticPipeline)
			if err != nil {
				log.Error(err, "Cannot update elasticpipeline status")
				return ctrl.Result{}, err
			}
		}
		return ctrl.Result{RequeueAfter: ErrorInterval}, nil
	}
	log.Info("esConfig generated from secret", "EsConfig", esConfig, "EsVersion", esConfig.Version)
	var elasticsearch = buildElasticsearchFromVersion(esConfig.Version)
	if err = elasticsearch.NewClient(esConfig, log); err != nil {
		log.Error(err, "Cannot initialize elasticsearch client")
		return ctrl.Result{}, err
	}

	if deleteRequest, err := managePipelineFinalizer(ctx, elasticPipeline, elasticsearch, log, r); err != nil {
		return ctrl.Result{}, err
	} else if !deleteRequest {
		if err := elasticsearch.PingES(ctx); err != nil {
			if pipelineStatusUpdated(&elasticPipeline.Status, &utils.EsStatus{Status: utils.StatusRetry, Message: err.Error()}, log) {
				err := r.Status().Update(ctx, &elasticPipeline)
				if err != nil {
					log.Error(err, "Could not update status for pipeline", "pipelineName", elasticPipeline.Spec.PipelineName)
					return ctrl.Result{}, err
				}
			}
			return ctrl.Result{RequeueAfter: RetryInterval}, nil
		}
		log.Info("create/update Elastic Pipeline", "pipelineName", elasticPipeline.Spec.PipelineName)
		esStatus, err := elasticsearch.CreateOrUpdatePipeline(ctx, elasticPipeline.Spec.PipelineName, elasticPipeline.Spec.Model)
		if err := updatePipelineESStatus(ctx, r, req, esStatus, log); err != nil {
			if apierrors.IsConflict(err) {
				log.Info("conflict: operation cannot be fulfilled on ElasticPipeline. Requeue to try again")
				return ctrl.Result{Requeue: true}, nil
			}
			log.Error(err, "unable to update ElasticPipeline status")
			return ctrl.Result{}, err
		}
		if esStatus.Status == utils.StatusError {
			//blocking error no need to Requeue or Requeue after a long interval
			return ctrl.Result{RequeueAfter: ErrorInterval}, nil
		} else if err != nil {
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

func updatePipelineESStatus(ctx context.Context, r *ElasticPipelineReconciler, req ctrl.Request, esStatus *utils.EsStatus, log logr.Logger) error {
	var elasticPipeline elasticv1alpha1.ElasticPipeline
	if err := r.Get(ctx, req.NamespacedName, &elasticPipeline); err != nil {
		if apierrors.IsNotFound(err) {
			log.Info("ElasticPipeline not found")
		} else {
			log.Error(err, "unable to fetch elasticpipeline object")
		}
		return client.IgnoreNotFound(err)
	}
	if pipelineStatusUpdated(&elasticPipeline.Status, esStatus, log) {
		if err := r.Status().Update(ctx, &elasticPipeline); err != nil {
			log.Error(err, "unable to update elasticpipeline object")
			return err
		}
	}
	return nil
}

func pipelineStatusUpdated(objectStatus *elasticv1alpha1.ElasticPipelineStatus, esStatus *utils.EsStatus, log logr.Logger) bool {
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

func managePipelineFinalizer(ctx context.Context, elasticPipeline elasticv1alpha1.ElasticPipeline, elasticsearch utils.Elasticsearch, log logr.Logger, r *ElasticPipelineReconciler) (bool, error) {
	finalizerName := fmt.Sprintf("finalizer.%v", elasticv1alpha1.GroupVersion.Group)
	deleteRequest := false

	if elasticPipeline.ObjectMeta.DeletionTimestamp.IsZero() {
		if !utils.ContainsString(elasticPipeline.ObjectMeta.Finalizers, finalizerName) {
			log.Info("register a finalizer")
			elasticPipeline.ObjectMeta.Finalizers = append(elasticPipeline.ObjectMeta.Finalizers, finalizerName)
			if err := r.Update(ctx, &elasticPipeline); err != nil {
				return false, err
			}
		}
	} else {
		log.Info("elasticpipeline is being deleted")
		deleteRequest = true
		if utils.ContainsString(elasticPipeline.ObjectMeta.Finalizers, finalizerName) {
			if elasticPipeline.Annotations[DeleteInClusterAnnotation] == "true" {
				if err := elasticsearch.DeletePipeline(ctx, elasticPipeline.Spec.PipelineName); err != nil {
					log.Error(err, "error while deleting elasticpipeline", "pipelineName", elasticPipeline.Spec.PipelineName)
				}
			} else {
				log.Info("elasticindex deletion will not delete elasticsearch pipeline", "pipelineName", elasticPipeline.Spec.PipelineName)
			}

			// remove finalizer from the list and update it.
			elasticPipeline.ObjectMeta.Finalizers = utils.RemoveString(elasticPipeline.ObjectMeta.Finalizers, finalizerName)
			if err := r.Update(ctx, &elasticPipeline); err != nil {
				return deleteRequest, err
			}
		}
	}
	return deleteRequest, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ElasticPipelineReconciler) SetupWithManager(mgr ctrl.Manager) error {
	if r.NamespacesRegexFilter == "" {
		return ctrl.NewControllerManagedBy(mgr).
			For(&elasticv1alpha1.ElasticPipeline{}).
			Complete(r)
	}
	var regex = r.NamespacesRegexFilter
	namespacesRegexFilter := predicate.Funcs{
		CreateFunc:  func(ce event.CreateEvent) bool { return utils.FilterByNamespacesRegex(ce.Meta, regex, r.Log) },
		DeleteFunc:  func(ce event.DeleteEvent) bool { return utils.FilterByNamespacesRegex(ce.Meta, regex, r.Log) },
		UpdateFunc:  func(ce event.UpdateEvent) bool { return utils.FilterByNamespacesRegex(ce.MetaNew, regex, r.Log) },
		GenericFunc: func(ce event.GenericEvent) bool { return utils.FilterByNamespacesRegex(ce.Meta, regex, r.Log) },
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&elasticv1alpha1.ElasticPipeline{}).
		WithEventFilter(namespacesRegexFilter).
		Complete(r)
}
