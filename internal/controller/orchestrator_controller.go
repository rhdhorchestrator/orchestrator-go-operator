/*
Copyright 2024.

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

package controller

import (
	"context"
	appsv1 "k8s.io/api/apps/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/controller"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	orchestratorv1alpha1 "github.com/parodos-dev/orchestrator-operator/api/v1alpha1"
)

// Definition to manage Orchestrator condition status.
const (
	TypeAvailable   string = "Available"
	TypeProgressing string = "Progressing"
	TypeDegraded    string = "Degraded"
)

// OrchestratorReconciler reconciles a Orchestrator object
type OrchestratorReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
}

//+kubebuilder:rbac:groups=orchestrator.parodos.dev,resources=orchestrators,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=orchestrator.parodos.dev,resources=orchestrators/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=orchestrator.parodos.dev,resources=orchestrators/finalizers,verbs=update
//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=events,verbs=create;patch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// The Reconcile function compares the state specified by
// the Orchestrator object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.17.3/pkg/reconcile
func (r *OrchestratorReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	log.Info("Starting reconciliation")

	// Fetch the Orchestrator instance
	// The purpose is to check if the Custom Resource for the Kind Orchestrator
	// is applied on the cluster if not we return nil to stop the reconciliation
	orchestrator := &orchestratorv1alpha1.Orchestrator{}

	err := r.Get(ctx, req.NamespacedName, orchestrator) // Lookup the Orchestrator instance for this reconcile request
	if err != nil {
		if apierrors.IsNotFound(err) {
			// If the custom resource is not found then it usually means that it was deleted or not created
			// In this way, we will stop the reconciliation
			log.Info("orchestrator resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		log.Error(err, "Failed to get orchestrator")
		return ctrl.Result{}, err
	}
	// Set the status to Unknown when no status is available - usually initial reconciliation.
	if orchestrator.Status.Conditions == nil || len(orchestrator.Status.Conditions) == 0 {
		meta.SetStatusCondition(
			&orchestrator.Status.Conditions,
			metav1.Condition{
				Type:    TypeAvailable,
				Status:  metav1.ConditionUnknown,
				Reason:  "Reconciling",
				Message: "Starting Reconciliation",
			},
		)
		if err = r.Status().Update(ctx, orchestrator); err != nil {
			log.Error(err, "Failed to update Orchestrator status")
			return ctrl.Result{}, err
		}
		// Re-fetch orchestrator Custom Resource after updating the status
		if err := r.Get(ctx, req.NamespacedName, orchestrator); err != nil {
			log.Error(err, "Failed to re fetch orchestrator")
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *OrchestratorReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&orchestratorv1alpha1.Orchestrator{}).
		Owns(&appsv1.Deployment{}).
		WithOptions(controller.Options{MaxConcurrentReconciles: 2}).
		Complete(r)
}
