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
	"github.com/operator-framework/api/pkg/operators/v1alpha1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/controller"

	olmclientset "github.com/operator-framework/operator-lifecycle-manager/pkg/api/client/clientset/versioned"
	// olmv1alpha1 "github.com/operator-framework/operator-lifecycle-manager/pkg/api/client/clientset/versioned/typed/operators/v1alpha1"

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
	OLMClient olmclientset.Interface
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

	// sonataFlowOperator
	// for creating and deleting use case
	sonataFlowOperator := orchestrator.Spec.SonataFlowOperator
	subscriptionName := sonataFlowOperator.Subscription.Name
	namespace := sonataFlowOperator.Subscription.Namespace

	// check if the sonataflow subscription operator is enabled
	// subscription is disabled,
	if !sonataFlowOperator.Enabled {
		// check if subscription exists using olm client
		sonataFlowSubscription, err := r.OLMClient.OperatorsV1alpha1().Subscriptions(namespace).Get(context.TODO(), subscriptionName, metav1.GetOptions{})
		if err != nil {
			logger.Error(err, "Subscription does not exists: %v")
			return ctrl.Result{}, err
		}
		logger.Info("Subscription exists: %s", sonataFlowSubscription.Name)

		// deleting subscription resource;
		//then requeue for x amount of time (not hoard the thread)
		err = r.OLMClient.OperatorsV1alpha1().Subscriptions(namespace).Delete(ctx, subscriptionName, metav1.DeleteOptions{})
		if err != nil {
			logger.Error(err, "Error occurred while deleting Subscription: %s", subscriptionName)
			return ctrl.Result{}, err
		}
		logger.Info("Successfully deleted Subscription: %s", subscriptionName)
	}

	// Subscription is enabled
	// check if CRD exists;
	// if the CRD exists, check if CR exists, if CR does not exist, create CR
	// if CR exists, check desired state is the same as current state.

	// if not install the sonataflow operator
	subscriptionObject := createSubscriptionObject(subscriptionName, namespace, sonataFlowOperator)
	installedSubscription, err := r.OLMClient.OperatorsV1alpha1().Subscriptions(namespace).Create(ctx, subscriptionObject, metav1.CreateOptions{})

	if err != nil {
		logger.Error(err, "Error occurred while creating Subscription: %s", subscriptionName)
		return ctrl.Result{}, err
	}

	logger.Info("Successfully installed Operator Subscription: %s", installedSubscription.Name)

	// testing

	// for updating the spec use case
	// check if the sonataflow CR spec matches the current state
	//
	// check status of resource, update it if not in desired state.

	return ctrl.Result{}, nil
}

func createSubscriptionObject(subscriptionName string, namespace string, sonataFlowOperator orchestratorv1alpha1.SonataFlowOperator) *v1alpha1.Subscription {
	sonataFlowSubscriptionDetails := sonataFlowOperator.Subscription
	subscriptionObject := &v1alpha1.Subscription{
		ObjectMeta: metav1.ObjectMeta{Namespace: namespace, Name: subscriptionName},
		Spec: &v1alpha1.SubscriptionSpec{
			Channel:                sonataFlowSubscriptionDetails.Channel,
			InstallPlanApproval:    v1alpha1.Approval(sonataFlowSubscriptionDetails.InstallPlanApproval),
			CatalogSource:          sonataFlowSubscriptionDetails.SourceName,
			StartingCSV:            sonataFlowSubscriptionDetails.StartingCSV,
			CatalogSourceNamespace: sonataFlowSubscriptionDetails.Namespace,
			Package:                sonataFlowSubscriptionDetails.Name,
		},
	}
	return subscriptionObject
}

// SetupWithManager sets up the controller with the Manager.
func (r *OrchestratorReconciler) SetupWithManager(mgr ctrl.Manager) error {
	//config, err := rest.InClusterConfig()
	config := mgr.GetConfig()

	// Create the OLM clientset using the config
	olmClient, err := olmclientset.NewForConfig(config)
	if err != nil {
		return nil
	}
	r.OLMClient = olmClient

	return ctrl.NewControllerManagedBy(mgr).
		For(&orchestratorv1alpha1.Orchestrator{}).
		//Owns(&appsv1.Deployment{}).
		WithOptions(controller.Options{MaxConcurrentReconciles: 2}).
		Complete(r)
}
