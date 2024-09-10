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
	"fmt"
	appsv1 "k8s.io/api/apps/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/controller"

	"github.com/operator-framework/api/pkg/operators/v1alpha1"
	olmclientset "github.com/operator-framework/operator-lifecycle-manager/pkg/api/client/clientset/versioned"
	orchestratorv1alpha1 "github.com/parodos-dev/orchestrator-operator/api/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// Definition to manage Orchestrator condition status.
const (
	TypeAvailable string = "Available"
	//TypeProgressing string = "Progressing"
	//TypeDegraded    string = "Degraded"
)

// OrchestratorReconciler reconciles a Orchestrator object
type OrchestratorReconciler struct {
	client.Client
	ClientSet *kubernetes.Clientset
	Scheme    *runtime.Scheme
	Recorder  record.EventRecorder
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
	logger := log.FromContext(ctx)
	logger.Info("Starting reconciliation")

	// Fetch the Orchestrator instance
	// The purpose is to check if the Custom Resource for the Kind Orchestrator
	// is applied on the cluster if not we return nil to stop the reconciliation
	orchestrator := &orchestratorv1alpha1.Orchestrator{}

	err := r.Get(ctx, req.NamespacedName, orchestrator) // Lookup the Orchestrator instance for this reconcile request
	if err != nil {
		if apierrors.IsNotFound(err) {
			// If the custom resource is not found then it usually means that it was deleted or not created
			// In this way, we will stop the reconciliation
			logger.Info("orchestrator resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		logger.Error(err, "Failed to get orchestrator")
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
			logger.Error(err, "Failed to update Orchestrator status")
			return ctrl.Result{}, err
		}
		// Re-fetch orchestrator Custom Resource after updating the status
		if err := r.Get(ctx, req.NamespacedName, orchestrator); err != nil {
			logger.Error(err, "Failed to re fetch orchestrator")
			return ctrl.Result{}, err
		}
	}

	// for creating and deleting use case

	sonataFlowOperator := orchestrator.Spec.SonataFlowOperator
	subscriptionName := sonataFlowOperator.Subscription.Name
	namespace := sonataFlowOperator.Subscription.Namespace

	// check if the sonataflow subscription operator is enabled
	// if disabled,
	if !sonataFlowOperator.Enabled {
		// check if subscription exists, delete it, then requeue for x amount of time (not hoard the thread)
		// Use SubscriptionLister to check if a Subscription exists
		sonataFlowSubscription, err := getSubscription(ctx, subscriptionName, namespace)
		if err != nil {
			logger.Error(err, "Subscription does not exists: %v")
		}
		logger.Info("Subscription exists: %s", sonataFlowSubscription.Name)
	}

	// if enabled, check if CRD exists, if not install the sonataflow operator

	// if the CRD exists, check if CR exists, if CR does not exist, create CR
	// if CR exists, check desired state is the same as current state.

	// testing

	// for updating the spec use case
	// check if the sonataflow CR spec matches the current state
	//
	// check status of resource, update it if not in desired state.

	//sonataflow := orchestrator.Spec.SonataFlowOperator

	return ctrl.Result{}, nil
}

// getSubscription to retrieve a Subscription via OLM
func getSubscription(ctx context.Context, subscriptionName, namespace string) (*v1alpha1.Subscription, error) {
	logger := log.FromContext(ctx)

	config, err := rest.InClusterConfig()
	if err != nil {
		logger.Error(err, "Error creating Kubernetes config: %v")
	}

	// Create the OLM clientset using the config
	olmClient, err := olmclientset.NewForConfig(config)
	if err != nil {
		logger.Error(err, "Failed to create OLM client")
	}
	// Retrieve the Subscription object from the cluster using OLM client
	subscription, err := olmClient.OperatorsV1alpha1().Subscriptions(namespace).Get(context.TODO(), subscriptionName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get Subscription %s in namespace %s: %v", subscriptionName, namespace, err)
	}

	if err != nil {
		if apierrors.IsNotFound(err) {
			logger.Info("Subscription %s in namespace %s not found", subscriptionName, namespace)
		}
		logger.Error(err, "error getting Subscription %s in namespace %s: %v", subscriptionName, namespace)
	}

	logger.Info("Subscription %s found in namespace %s with status: %v", subscriptionName, namespace, subscription.Status)
	return subscription, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *OrchestratorReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&orchestratorv1alpha1.Orchestrator{}).
		Owns(&appsv1.Deployment{}).
		WithOptions(controller.Options{MaxConcurrentReconciles: 2}).
		Complete(r)
}
