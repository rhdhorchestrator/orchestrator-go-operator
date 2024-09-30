/*
Copyright 2024 Red Hat, Inc.

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
	sonataapi "github.com/apache/incubator-kie-kogito-serverless-operator/api/v1alpha08"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/controller"

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
)

// OrchestratorReconciler reconciles an Orchestrator object
type OrchestratorReconciler struct {
	client.Client
	OLMClient olmclientset.Interface
	ClientSet *kubernetes.Clientset
	Scheme    *runtime.Scheme
	Recorder  record.EventRecorder
}

//+kubebuilder:rbac:groups=rhdh.redhat.com,resources=orchestrators,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=rhdh.redhat.com,resources=orchestrators/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=rhdh.redhat.com,resources=orchestrators/finalizers,verbs=update
//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=namespaces;events,verbs=list;get;create;delete;patch;watch
//+kubebuilder:rbac:groups=apiextensions.k8s.io,resources=customresourcedefinitions,verbs=get;list;watch
//+kubebuilder:rbac:groups=operators.coreos.com,resources=subscriptions;operatorgroups;catalogsources,verbs=get;list;watch;create;delete;patch
//+kubebuilder:rbac:groups=sonataflow.org,resources=sonataflows;sonataflowclusterplatforms;sonataflowplatforms,verbs=get;list;watch;create;delete;patch;update

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
	sonataFlowOperator := orchestrator.Spec.SonataFlowOperator
	err = r.reconcileSonataFlow(ctx, sonataFlowOperator, orchestrator)
	if err != nil {
		logger.Error(err, "Error occurred when installing SonataFlow resources")
		return ctrl.Result{}, err
	}

	// handle backstage
	//rhdhOperator := orchestrator.Spec.RhdhOperator

	return ctrl.Result{}, nil
}

func (r *OrchestratorReconciler) reconcileSonataFlow(
	ctx context.Context,
	sonataFlowOperator orchestratorv1alpha1.SonataFlowOperator,
	orchestrator *orchestratorv1alpha1.Orchestrator) error {

	sfLogger := log.FromContext(ctx)
	sfLogger.Info("Starting reconciliation for SonataFlow")
	subscriptionName := sonataFlowOperator.Subscription.Name
	namespace := sonataFlowOperator.Subscription.Namespace

	// if subscription is disabled; check if subscription exists and handle delete
	if !sonataFlowOperator.Enabled {
		// check if subscription exists using olm client
		subscriptionExists, err := checkSubscriptionExists(ctx, r.OLMClient, namespace, subscriptionName)
		if err != nil {
			sfLogger.Error(err, "Error occurred when checking subscription exists", "SubscriptionName", subscriptionName)
			return err
		}
		if subscriptionExists {
			// deleting subscription resource
			err = r.OLMClient.OperatorsV1alpha1().Subscriptions(namespace).Delete(ctx, subscriptionName, metav1.DeleteOptions{})
			if err != nil {
				sfLogger.Error(err, "Error occurred while deleting Subscription", "SubscriptionName", subscriptionName, "Namespace", namespace)
				//return ctrl.Result{RequeueAfter: 5 * time.Minute}, err
				return err
			}
			sfLogger.Info("Successfully deleted Subscription: %s", subscriptionName)
			return nil
		}
	}

	// Subscription is enabled; check if subscription exists
	subscriptionExists, err := checkSubscriptionExists(ctx, r.OLMClient, namespace, subscriptionName)
	if err != nil {
		sfLogger.Error(err, "Error occurred when checking subscription exists", "SubscriptionName", subscriptionName)
		return err
	}
	if !subscriptionExists {
		err := installOperatorViaSubscription(ctx, r.Client, r.OLMClient, namespace, subscriptionName, sonataFlowOperator)
		if err != nil {
			sfLogger.Error(err, "Error occurred when installing operator", "SubscriptionName", subscriptionName)
			return err
		}
		sfLogger.Info("Operator successfully installed", "SubscriptionName", subscriptionName)
	}

	// subscription exists; check if CRD exists;
	sonataFlowClusterPlatformCRD := &apiextensionsv1.CustomResourceDefinition{}
	err = r.Get(ctx, types.NamespacedName{Name: SonataFlowClusterPlatformCRDName, Namespace: namespace}, sonataFlowClusterPlatformCRD)

	if err != nil {
		if apierrors.IsNotFound(err) {
			// CRD does not exist
			sfLogger.Info("CRD resource not found.", "SubscriptionName", subscriptionName, "Namespace", namespace)
			return nil // do we want to re-attempt subscription installation?
		}
		sfLogger.Error(err, "Error occurred when retrieving CRD", "CRD", SonataFlowClusterPlatformCRDName)
	}

	// CRD exist; check and handle sonataflowclusterplatform CR
	err = handleSonataFlowClusterCR(ctx, r.Client, SonataFlowClusterPlatformCRName)
	if err != nil {
		sfLogger.Error(err, "Error occurred when creating SonataFlowClusterCR", "CR-Name", SonataFlowClusterPlatformCRName)
		return err

	} else {
		// create sonataflowplatform  CR
		err = createSonataFlowPlatformCR(ctx, r.Client, orchestrator, SonataFlowClusterPlatformCRName)
		if err != nil {
			sfLogger.Error(err, "Error occurred when creating SonataFlowPlatform", "CR-Name", SonataFlowClusterPlatformCRName)
			return err
		}
	}
	return err
}

// SetupWithManager sets up the controller with the Manager.
func (r *OrchestratorReconciler) SetupWithManager(mgr ctrl.Manager) error {
	config := mgr.GetConfig()

	// Create the OLM clientset using the config
	olmClient, err := olmclientset.NewForConfig(config)
	if err != nil {
		return nil
	}
	r.OLMClient = olmClient

	return ctrl.NewControllerManagedBy(mgr).
		For(&orchestratorv1alpha1.Orchestrator{}).
		Owns(&sonataapi.SonataFlow{}).
		Owns(&sonataapi.SonataFlowClusterPlatform{}).
		WithOptions(controller.Options{MaxConcurrentReconciles: 2}).
		Complete(r)
}
