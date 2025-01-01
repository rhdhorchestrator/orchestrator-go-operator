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
	"github.com/parodos-dev/orchestrator-operator/internal/controller/kube"
	"github.com/parodos-dev/orchestrator-operator/internal/controller/rhdh"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"time"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/controller"

	configv1 "github.com/openshift/api/config/v1"
	olmclientset "github.com/operator-framework/operator-lifecycle-manager/pkg/api/client/clientset/versioned"
	orchestratorv1alpha2 "github.com/parodos-dev/orchestrator-operator/api/v1alpha2"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

const (
	// Definition to manage Orchestrator condition status.
	TypeAvailable   string = "Available"
	TypeProgressing string = "Progressing"
	TypeDegrading   string = "Degrading"

	// Finalizer Definition
	FinalizerCRCleanup = "rhdh.redhat.com/orchestrator-cleanup"

	RequeueAfterTime = 5
)

// OrchestratorReconciler reconciles an Orchestrator object
type OrchestratorReconciler struct {
	client.Client
	OLMClient olmclientset.Clientset
	Scheme    *runtime.Scheme
	Recorder  record.EventRecorder
}

//+kubebuilder:rbac:groups=rhdh.redhat.com,resources=orchestrators,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=rhdh.redhat.com,resources=orchestrators/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=rhdh.redhat.com,resources=orchestrators/finalizers,verbs=update
//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=secrets;configmaps;namespaces;events,verbs=list;get;create;delete;patch;watch;update
//+kubebuilder:rbac:groups=apiextensions.k8s.io,resources=customresourcedefinitions,verbs=get;list;watch
//+kubebuilder:rbac:groups=operators.coreos.com,resources=subscriptions;operatorgroups;clusterserviceversions;catalogsources;installplans,verbs=get;list;watch;create;delete;patch;update
//+kubebuilder:rbac:groups=sonataflow.org,resources=sonataflows;sonataflowclusterplatforms;sonataflowplatforms,verbs=get;list;watch;create;delete;patch;update
//+kubebuilder:rbac:groups=operator.knative.dev,resources=knativeeventings;knativeservings,verbs=get;list;watch;create;delete;patch;update
//+kubebuilder:rbac:groups=rhdh.redhat.com,resources=backstages,verbs=get;list;create;delete;patch;watch
//+kubebuilder:rbac:groups=config.openshift.io,resources=ingresses,verbs=get;list;watch
// +kubebuilder:rbac:groups=networking.k8s.io,resources=ingresses;networkpolicies,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=tekton.dev,resources=pipelines,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=tekton.dev,resources=tasks,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=argoproj.io,resources=appprojects,verbs=get;list;watch;create;update;patch;delete

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
	orchestrator := &orchestratorv1alpha2.Orchestrator{}

	err := r.Get(ctx, req.NamespacedName, orchestrator) // Lookup the Orchestrator instance for this reconcile request
	if err != nil {
		if apierrors.IsNotFound(err) {
			// If the custom resource is not found then it usually means that it was deleted or not created
			// In this way, we will stop the reconciliation
			logger.Info("Orchestrator resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		logger.Error(err, "Failed to get orchestrator")
		return ctrl.Result{}, err
	}

	if !orchestrator.DeletionTimestamp.IsZero() {
		err := r.handleCleanUp(ctx, orchestrator)
		if err != nil {
			return ctrl.Result{RequeueAfter: RequeueAfterTime * time.Minute}, err
		}
		// Remove the finalizer to complete deletion
		controllerutil.RemoveFinalizer(orchestrator, FinalizerCRCleanup)
		if err := r.Update(ctx, orchestrator); err != nil {
			return ctrl.Result{}, err
		}
		logger.Info("Successfully removed Orchestrator Custom Resource")
	}

	// Add finalizer if not present
	if err := r.addFinalizers(ctx, orchestrator); err != nil {
		return ctrl.Result{}, err
	}

	// Set the status to Unknown when no status is available - usually initial reconciliation.
	if orchestrator.Status.Conditions == nil || len(orchestrator.Status.Conditions) == 0 {
		if err := r.UpdateStatus(ctx, orchestrator, orchestratorv1alpha2.RunningPhase, metav1.Condition{
			Type:               TypeAvailable,
			Status:             metav1.ConditionUnknown,
			Reason:             "Reconciling",
			Message:            "Starting Reconciliation",
			LastTransitionTime: metav1.Now(),
		}); err != nil {
			return ctrl.Result{}, err
		}
		// Re-fetch orchestrator Custom Resource after updating the status
		if err := r.Get(ctx, req.NamespacedName, orchestrator); err != nil {
			logger.Error(err, "Failed to re fetch orchestrator")
			return ctrl.Result{}, err
		}
	}

	argoCDEnabled := orchestrator.Spec.ArgoCd.Enabled
	tektonEnabled := orchestrator.Spec.Tekton.Enabled
	serverlessWorkflowNamespace := orchestrator.Spec.PlatformConfig.Namespace

	// handle serverless logic
	serverlessLogicOperator := orchestrator.Spec.ServerlessLogicOperator
	if err = r.reconcileServerlessLogic(ctx, serverlessLogicOperator, orchestrator); err != nil {
		logger.Error(err, "Error occurred when installing Serverless Logic resources")
		_ = r.UpdateStatus(ctx, orchestrator, orchestratorv1alpha2.FailedPhase, metav1.Condition{
			Type:               TypeDegrading,
			Status:             metav1.ConditionFalse,
			Reason:             "ReconcilingOSLResourcesFailed",
			Message:            err.Error(),
			LastTransitionTime: metav1.Now(),
		})
		return ctrl.Result{RequeueAfter: RequeueAfterTime * time.Minute}, err
	}

	//handle knative
	serverlessOperator := orchestrator.Spec.ServerlessOperator
	if err := r.reconcileKnative(ctx, serverlessOperator); err != nil {
		logger.Error(err, "Error occurred when installing K-Native resources")
		_ = r.UpdateStatus(ctx, orchestrator, orchestratorv1alpha2.FailedPhase, metav1.Condition{
			Type:               TypeDegrading,
			Status:             metav1.ConditionFalse,
			Reason:             "ReconcilingKNativeResourcesFailed",
			Message:            err.Error(),
			LastTransitionTime: metav1.Now(),
		})
		return ctrl.Result{RequeueAfter: RequeueAfterTime * time.Minute}, err
	}

	// handle RHDH
	rhdhConfig := orchestrator.Spec.RHDHConfig
	if err = r.reconcileRHDH(ctx, serverlessWorkflowNamespace, argoCDEnabled, tektonEnabled, rhdhConfig); err != nil {
		logger.Error(err, "Error occurred when installing RHDH resources")
		_ = r.UpdateStatus(ctx, orchestrator, orchestratorv1alpha2.FailedPhase, metav1.Condition{
			Type:               TypeDegrading,
			Status:             metav1.ConditionFalse,
			Reason:             "ReconcilingRHDHResourcesFailed",
			Message:            err.Error(),
			LastTransitionTime: metav1.Now(),
		})
		return ctrl.Result{Requeue: true, RequeueAfter: 3 * time.Minute}, err
	}
	if err = r.reconcileNetworkPolicy(ctx, orchestrator); err != nil {
		logger.Error(err, "Error occurred when installing NetworkPolicy")
		_ = r.UpdateStatus(ctx, orchestrator, orchestratorv1alpha2.FailedPhase, metav1.Condition{
			Type:               TypeDegrading,
			Status:             metav1.ConditionFalse,
			Reason:             "ReconcilingNetworkPolicyFailed",
			Message:            err.Error(),
			LastTransitionTime: metav1.Now(),
		})
		return ctrl.Result{Requeue: true, RequeueAfter: 3 * time.Minute}, err
	}
	return ctrl.Result{}, nil
}

func (r *OrchestratorReconciler) reconcileServerlessLogic(
	ctx context.Context,
	serverlessLogicOperator orchestratorv1alpha2.ServerlessLogicOperator,
	orchestrator *orchestratorv1alpha2.Orchestrator) error {

	sfLogger := log.FromContext(ctx)
	sfLogger.Info("Starting reconciliation for Serverless Logic")

	serverlessWorkflowNamespace := orchestrator.Spec.PlatformConfig.Namespace

	// if subscription is disabled;
	// check if subscription exists and handle clean up if necessary
	if !serverlessLogicOperator.InstallOperator {
		// handle clean up
		if err := handleServerlessLogicCleanUp(ctx, r.Client, r.OLMClient, serverlessWorkflowNamespace); err != nil {
			return err
		}
	}
	// Subscription is enabled; check namespace exist
	if _, err := kube.CheckNamespaceExist(ctx, r.Client, serverlessWorkflowNamespace); err != nil {
		if apierrors.IsNotFound(err) {
			sfLogger.Info("Creating namespace", "NS", serverlessWorkflowNamespace)
			if err := kube.CreateNamespace(ctx, r.Client, serverlessWorkflowNamespace); err != nil {
				sfLogger.Error(err, "Error occurred when creating namespace", "NS", serverlessWorkflowNamespace)
				return err
			}
		}
		sfLogger.Error(err, "Error occurred when checking namespace exists", "NS", serverlessWorkflowNamespace)
		return err
	}

	if err := handleServerlessLogicOperatorInstallation(ctx, r.Client, r.OLMClient); err != nil {
		sfLogger.Error(err, "Error occurred when installing OSL Operator resources")
		return err
	}

	// subscription exists; check if CRD exists;
	sonataFlowClusterPlatformCRD := &apiextensionsv1.CustomResourceDefinition{}
	if err := r.Get(ctx, types.NamespacedName{Name: sonataFlowClusterPlatformCRDName, Namespace: serverlessWorkflowNamespace}, sonataFlowClusterPlatformCRD); err != nil {
		if apierrors.IsNotFound(err) {
			// CRD does not exist
			sfLogger.Info("CRD resource not found.", "SubscriptionName", serverlessLogicSubscriptionName, "Namespace", serverlessWorkflowNamespace)
			return err
		}
		sfLogger.Error(err, "Error occurred when retrieving CRD", "CRD", sonataFlowClusterPlatformCRDName)
		return err
	}

	// handle serveless logic CRs
	if err := handleServerlessLogicCR(ctx, r.Client, orchestrator); err != nil {
		return err
	}
	sfLogger.Info("Successfully created ServerlessLogic Resources")
	return nil
}

func (r *OrchestratorReconciler) reconcileKnative(ctx context.Context, serverlessOperator orchestratorv1alpha2.ServerlessOperator) error {
	knativeLogger := log.FromContext(ctx)
	knativeLogger.Info("Starting Reconciliation for K-Native Serverless")

	// if subscription is disabled; check if subscription exists and handle delete
	if !serverlessOperator.InstallOperator {
		// handle cleanup
		if err := handleKnativeCleanUp(ctx, r.Client, r.OLMClient); err != nil {
			return err
		}
	}

	// Subscription is enabled;
	if err := handleKNativeOperatorInstallation(ctx, r.Client, r.OLMClient); err != nil {
		knativeLogger.Error(err, "Error occurred when installing Knative Operator resources")
		return err
	}

	// handle knative CRs
	if err := handleKnativeCR(ctx, r.Client); err != nil {
		knativeLogger.Error(err, "Error occurred when handling Knative Custom Resources")
		return err
	}
	knativeLogger.Info("Successfully created Knative Custom Resources")
	return nil
}

func (r *OrchestratorReconciler) reconcileRHDH(
	ctx context.Context, serverlessWorkflowNamespace string,
	argoCDEnabled, tektonEnabled bool,
	rhdhConfig orchestratorv1alpha2.RHDHConfig) error {

	logger := log.FromContext(ctx)
	logger.Info("Starting Reconciliation for RHDH")

	subscriptionName := rhdhConfig.Name
	namespace := rhdhConfig.Namespace

	// if subscription is disabled; check if subscription exists and handle delete
	if !rhdhConfig.InstallOperator {
		if err := rhdh.HandleRHDHCleanUp(ctx, r.Client, r.OLMClient, namespace); err != nil {
			logger.Error(err, "Error occurred when cleaning up RHDH", "SubscriptionName", subscriptionName)
			return err
		}
	}

	if err := rhdh.HandleRHDHOperatorInstallation(ctx, r.Client, r.OLMClient); err != nil {
		logger.Error(err, "Error occurred when installing RHDH Operator resources")
		return err
	}

	clusterDomain, _ := r.getClusterDomain(ctx)
	// create secret
	if err := rhdh.CreateRHDHSecret(namespace, ctx, r.Client); err != nil {
		return err
	}
	// create configmap
	bsConfigMapList := rhdh.GetConfigmapList(ctx, r.Client, clusterDomain, serverlessWorkflowNamespace, argoCDEnabled, tektonEnabled, rhdhConfig)
	logger.Info("Configmap list", "CM-List", bsConfigMapList)
	// handle RHDH CR
	if err := rhdh.HandleRHDHCR(rhdhConfig, bsConfigMapList, ctx, r.Client); err != nil {
		return err
	}
	return nil
}

// getClusterDomain retrieves the OpenShift cluster domain from the Ingress resource
func (r *OrchestratorReconciler) getClusterDomain(ctx context.Context) (string, error) {
	gcdLogger := log.FromContext(ctx)
	ingress := &configv1.Ingress{}
	err := r.Get(ctx, client.ObjectKey{Name: "cluster"}, ingress)
	if err != nil {
		gcdLogger.Error(err, "Unable to retrieve OpenShift Ingress resource")
		return "", err
	}

	clusterDomain := ingress.Spec.Domain
	if ingress.Spec.Domain == "" {
		gcdLogger.Error(err, "Cluster domain not set in Ingress resource")
		return "", err
	}
	gcdLogger.Info("Successfully retrieved cluster domain", "Domain", clusterDomain)
	return clusterDomain, nil
}

func (r *OrchestratorReconciler) addFinalizers(ctx context.Context, orchestrator *orchestratorv1alpha2.Orchestrator) error {
	if !controllerutil.ContainsFinalizer(orchestrator, FinalizerCRCleanup) {
		controllerutil.AddFinalizer(orchestrator, FinalizerCRCleanup)
		if err := r.Update(ctx, orchestrator); err != nil {
			return err
		}
	}
	return nil
}

func (r *OrchestratorReconciler) handleCleanUp(ctx context.Context, orchestrator *orchestratorv1alpha2.Orchestrator) error {
	// cleanup Knative
	if err := handleKnativeCleanUp(ctx, r.Client, r.OLMClient); err != nil {
		return err
	}
	// cleanup Serverless Logic
	if err := handleServerlessLogicCleanUp(ctx, r.Client, r.OLMClient, orchestrator.Spec.PlatformConfig.Namespace); err != nil {
		return err
	}
	// cleanup RHDH
	if err := rhdh.HandleRHDHCleanUp(ctx, r.Client, r.OLMClient, orchestrator.Spec.RHDHConfig.Namespace); err != nil {
		return err
	}
	return nil
}

// UpdateStatus sets the status of orchestrator.
func (r *OrchestratorReconciler) UpdateStatus(ctx context.Context, orchestrator *orchestratorv1alpha2.Orchestrator, phase orchestratorv1alpha2.OrchestratorPhase, condition metav1.Condition) error {
	logger := log.FromContext(ctx)
	orchestrator.Status.Phase = phase
	meta.SetStatusCondition(&orchestrator.Status.Conditions, condition)

	err := r.Status().Update(ctx, orchestrator)
	if err != nil {
		logger.Error(err, "Failed to update Orchestrator status")
		return err
	}
	return nil
}

func (r *OrchestratorReconciler) reconcileNetworkPolicy(ctx context.Context, orchestrator *orchestratorv1alpha2.Orchestrator) error {
	logger := log.FromContext(ctx)
	logger.Info("Reconciling Network Policy...")
	if err := handleNetworkPolicy(r.Client, ctx, orchestrator.Spec.PlatformConfig.Namespace, orchestrator.Spec.RHDHConfig.Namespace, orchestrator.Spec.PostgresConfig.Namespace); err != nil {
		logger.Error(err, "Error occurred when reconciling Network Policy", "NP", NetworkPolicyName)
		return err
	}
	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *OrchestratorReconciler) SetupWithManager(mgr ctrl.Manager) error {
	config := mgr.GetConfig()

	// Create the OLM clientset using the config
	olmClient, err := olmclientset.NewForConfig(config)
	if err != nil {
		return err
	}
	r.OLMClient = *olmClient

	return ctrl.NewControllerManagedBy(mgr).
		For(&orchestratorv1alpha2.Orchestrator{}).
		Owns(&orchestratorv1alpha2.Orchestrator{}).
		WithOptions(controller.Options{MaxConcurrentReconciles: 2}).
		Complete(r)
}
