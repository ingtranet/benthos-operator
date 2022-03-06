/*
Copyright 2022.

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
	ingtranetv1alpha1 "github.com/ingtranet/benthos-operator/api/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"reflect"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"time"
)

// BenthosPipelineReconciler reconciles a BenthosPipeline object
type BenthosPipelineReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=ingtra.net,resources=benthospipelines,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=ingtra.net,resources=benthospipelines/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=ingtra.net,resources=benthospipelines/finalizers,verbs=update
//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=configmaps,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the BenthosPipeline object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.11.0/pkg/reconcile
func (r *BenthosPipelineReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx).WithValues("resource", req.Namespace)

	pipeline := &ingtranetv1alpha1.BenthosPipeline{}
	if err := r.Get(ctx, req.NamespacedName, pipeline); err != nil {
		if errors.IsNotFound(err) {
			logger.Info("Cannot find resource")
			return ctrl.Result{}, nil
		}
		logger.Error(err, "Finding resource failed")
		return ctrl.Result{}, err
	}

	// reconcile the configMap
	result, err := r.reconcileConfigMap(ctx, req, pipeline)
	if err != nil {
		logger.Error(err, "Reconciling configMap failed")
		return ctrl.Result{}, err
	}
	if result != nil {
		return *result, nil
	}

	// reconcile the deployment
	result, err = r.reconcileDeployment(ctx, req, pipeline)
	if err != nil {
		logger.Error(err, "Reconciling deployment failed")
		return ctrl.Result{}, err
	}
	if result != nil {
		return *result, nil
	}

	logger.Info("Nothing to do")
	return ctrl.Result{}, nil
}

func (r *BenthosPipelineReconciler) reconcileConfigMap(ctx context.Context, req ctrl.Request, p *ingtranetv1alpha1.BenthosPipeline) (*ctrl.Result, error) {
	logger := log.FromContext(ctx).WithValues("resource", req.Namespace)

	desiredCm, err := getConfigMapFor(p)
	if err != nil {
		logger.Error(err, "Error in configMap spec")
		return nil, err
	}

	cm := &corev1.ConfigMap{}
	err = r.Get(ctx, types.NamespacedName{Name: p.Name, Namespace: p.Namespace}, cm)
	if err != nil && errors.IsNotFound(err) {
		logger.Info("Creating new configMap")
		ctrl.SetControllerReference(p, desiredCm, r.Scheme)
		if err := r.Create(ctx, desiredCm); err != nil {
			logger.Error(err, "Creating configMap failed")
			return nil, err
		}
		logger.Info("ConfigMap creation succeed. Re-queuing...")
		return &ctrl.Result{Requeue: true}, nil
	} else if err != nil {
		logger.Error(err, "Finding configMap failed")
		return nil, err
	}

	logger.Info("Found configMap. Now checking")
	if !reflect.DeepEqual(cm.Data, desiredCm.Data) {
		logger.Info("Updating config.yaml of the configMap")
		cm.Data = desiredCm.Data
		if err := r.Update(ctx, cm); err != nil {
			logger.Error(err, "Updating config.yaml failed")
			return nil, err
		}
	}
	logger.Info("ConfigMap update not required")
	return nil, nil
}

func (r *BenthosPipelineReconciler) reconcileDeployment(ctx context.Context, req ctrl.Request, p *ingtranetv1alpha1.BenthosPipeline) (*ctrl.Result, error) {
	logger := log.FromContext(ctx).WithValues("resource", req.Namespace)

	desiredDeployment, err := getDeploymentFor(p)
	if err != nil {
		logger.Error(err, "Error in deployment spec")
		return nil, err
	}

	dep := &appsv1.Deployment{}
	err = r.Get(ctx, types.NamespacedName{Name: p.Name, Namespace: p.Namespace}, dep)
	if err != nil && errors.IsNotFound(err) {
		logger.Info("Creating new deployment")
		ctrl.SetControllerReference(p, desiredDeployment, r.Scheme)
		if err := r.Create(ctx, desiredDeployment); err != nil {
			logger.Error(err, "Creating deployment failed")
			return nil, err
		}
		logger.Info("Deployment creation succeed. Re-queuing...")
		return &ctrl.Result{Requeue: true, RequeueAfter: 10 * time.Second}, nil
	} else if err != nil {
		logger.Error(err, "Finding deployment failed")
		return nil, err
	}

	logger.Info("Found deployment. Now checking")
	if !reflect.DeepEqual(dep.Spec, desiredDeployment.Spec) {
		logger.Info("Updating deployment")
		dep.Spec = desiredDeployment.Spec
		if err := r.Update(ctx, dep); err != nil {
			logger.Error(err, "Updating deployment failed")
			return nil, err
		}
	}

	logger.Info("Deployment update not required")
	return nil, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *BenthosPipelineReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&ingtranetv1alpha1.BenthosPipeline{}).
		Owns(&corev1.ConfigMap{}).
		Owns(&appsv1.Deployment{}).
		Complete(r)
}
