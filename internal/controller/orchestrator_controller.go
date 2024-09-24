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

	sonataapi "github.com/apache/incubator-kie-kogito-serverless-operator/api/v1alpha08"
	operatorsv1 "github.com/operator-framework/api/pkg/operators/v1"
	"github.com/operator-framework/api/pkg/operators/v1alpha1"
	operatorsv1alpha1 "github.com/operator-framework/api/pkg/operators/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"time"

	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
	//TypeProgressing string = "Progressing"
	//TypeDegraded    string = "Degraded"
)

// OrchestratorReconciler reconciles an Orchestrator object
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
	// for creating and deleting use case
	sonataFlowOperator := orchestrator.Spec.SonataFlowOperator
	subscriptionName := sonataFlowOperator.Subscription.Name
	namespace := sonataFlowOperator.Subscription.Namespace

	// if subscription is disabled; check if subscription exists
	if !sonataFlowOperator.Enabled {
		// check if subscription exists using olm client
		subscriptionExists, err := checkSubscriptionExists(ctx, r.OLMClient, namespace, subscriptionName)
		if err != nil {
			logger.Error(err, "Error occurred when checking subscription exists", "SubscriptionName", subscriptionName)
			return ctrl.Result{}, err
		}
		if subscriptionExists {
			// deleting subscription resource
			err = r.OLMClient.OperatorsV1alpha1().Subscriptions(namespace).Delete(ctx, subscriptionName, metav1.DeleteOptions{})
			if err != nil {
				logger.Error(err, "Error occurred while deleting Subscription", "SubscriptionName", subscriptionName, "Namespace", namespace)
				return ctrl.Result{RequeueAfter: 5 * time.Minute}, err
			}
			logger.Info("Successfully deleted Subscription: %s", subscriptionName)
			return ctrl.Result{}, nil
		}
	}

	// Subscription is enabled; check if subscription exists
	subscriptionExists, err := checkSubscriptionExists(ctx, r.OLMClient, namespace, subscriptionName)
	if err != nil {
		logger.Error(err, "Error occurred when checking subscription exists", "SubscriptionName", subscriptionName)
		return ctrl.Result{RequeueAfter: 2 * time.Second}, err
	}
	if !subscriptionExists {
		err := installOperatorSubscription(ctx, r.Client, r.OLMClient, namespace, subscriptionName, sonataFlowOperator)
		if err != nil {
			return ctrl.Result{RequeueAfter: 2 * time.Second}, err
		}
	}

	// subscription exists; check if CRD exists;
	crdGroupName := "sonataflowclusterplatforms.sonataflow.org"
	sonataCRD := &apiextensionsv1.CustomResourceDefinition{}
	err = r.Get(ctx, types.NamespacedName{Name: crdGroupName, Namespace: namespace}, sonataCRD)

	if err != nil {
		if apierrors.IsNotFound(err) {
			// CRD does not exist
			logger.Info("CRD resource not found.", "SubscriptionName", subscriptionName, "Namespace", namespace)
			return ctrl.Result{}, nil // do we want to attempt subscription installation?
		}
		logger.Error(err, "Error occurred when retrieving CRD", "CRD", crdGroupName)
	}
	// CRD exists; check CR exists
	sfcCR := &sonataapi.SonataFlowClusterPlatform{}
	sonataClusterName := "cluster-platform"
	err = r.Get(ctx, types.NamespacedName{Name: sonataClusterName}, sfcCR)
	if err == nil {
		// check for CR updates
		if apierrors.IsNotFound(err) {
			logger.Info("CR resource not found.", "CR-Name", sonataClusterName)
		}
		logger.Error(err, "Error occurred when retrieving CR", "CR-Name", sonataClusterName)
		return ctrl.Result{}, err
	}

	// CR does not exists; create CR
	err = createSonataFlowClusterCR(ctx, r.Client, sonataClusterName)
	return ctrl.Result{}, nil
}

func createSonataFlowClusterCR(ctx context.Context, client client.Client, crName string) error {
	logger := log.FromContext(ctx)

	// Create sonataflow cluster CR object
	sonataFlowClusterCR := &sonataapi.SonataFlowClusterPlatform{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "sonataflow.org/v1alpha08",  // CRD group and version
			Kind:       "SonataFlowClusterPlatform", // CRD kind
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      crName,             // Name of the CR
			Namespace: "sonataflow-infra", // Namespace of the CR
		},
		Spec: sonataapi.SonataFlowClusterPlatformSpec{
			PlatformRef: sonataapi.SonataFlowPlatformRef{
				Name:      "sonataflow-platform",
				Namespace: "sonataflow-infra",
			},
		},
	}

	// Create sonataflow cluster CR
	if err := client.Create(ctx, sonataFlowClusterCR); err != nil {
		logger.Error(err, "Error occurred when creating Custom Resource", "CR-Name", crName)
		return err
	}
	logger.Info("Successfully created SonataFlow Cluster resource %s", sonataFlowClusterCR.Name)
	return nil
}

//func getSonataFlowPersistence(orchestrator *orchestratorv1alpha1.Orchestrator) orchestratorv1alpha1.Persistence {
//	return orchestratorv1alpha1.Persistence{
//		Postgresql: orchestratorv1alpha1.Postgresql{
//			SecretRef: orchestratorv1alpha1.PostgresAuthSecret{
//				SecretName:  orchestrator.Spec.PostgresDB.AuthSecret.SecretName,
//				UserKey:     orchestrator.Spec.PostgresDB.AuthSecret.UserKey,
//				PasswordKey: orchestrator.Spec.PostgresDB.AuthSecret.PasswordKey,
//			},
//			ServiceRef: orchestratorv1alpha1.ServiceRef{
//				Name:      orchestrator.Spec.PostgresDB.ServiceName,
//				Namespace: orchestrator.Spec.PostgresDB.ServiceNameSpace,
//			},
//		}}
//}

func getOperatorGroup(ctx context.Context, client client.Client,
	namespace string, operatorGroupName string) error {
	logger := log.FromContext(ctx)
	// check if operator group exists
	operatorGroup := &operatorsv1.OperatorGroup{}
	err := client.Get(ctx, types.NamespacedName{Name: operatorGroupName, Namespace: namespace}, operatorGroup)
	if err == nil {
		logger.Info("Operator Group already exists", "Operator Group", operatorGroupName)
		return nil
	}
	// create operator group
	sfog := &operatorsv1.OperatorGroup{
		ObjectMeta: metav1.ObjectMeta{Name: operatorGroupName, Namespace: namespace},
	}
	err = client.Create(ctx, sfog)
	if err != nil {
		logger.Error(err, "Error occurred when creating OperatorGroup resource", "Namespace", namespace)
		return err
	}
	return nil
}

func checkSubscriptionExists(
	ctx context.Context, olmClientSet olmclientset.Interface,
	namespace string, subscriptionName string) (bool, error) {
	logger := log.FromContext(ctx)

	subscription, err := olmClientSet.OperatorsV1alpha1().Subscriptions(namespace).Get(ctx, subscriptionName, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			logger.Info("Subscription resource not found.", "SubscriptionName", subscriptionName, "Namespace", namespace)
			return false, nil
		}
		logger.Error(err, "Failed to check Subscription does not exists", "SubscriptionName", subscriptionName)
		return false, err
	}
	logger.Info("Subscription exists", "SubscriptionName", subscription.Name)
	return true, nil
}

func createSubscriptionObject(
	subscriptionName string, namespace string,
	sonataFlowOperator orchestratorv1alpha1.SonataFlowOperator) *v1alpha1.Subscription {
	logger := log.Log.WithName("subscriptionObject")
	logger.Info("Creating subscription object")

	sonataFlowSubscriptionDetails := sonataFlowOperator.Subscription
	subscriptionObject := &v1alpha1.Subscription{
		ObjectMeta: metav1.ObjectMeta{Namespace: namespace, Name: subscriptionName},
		Spec: &v1alpha1.SubscriptionSpec{
			Channel:                sonataFlowSubscriptionDetails.Channel,
			InstallPlanApproval:    v1alpha1.Approval(sonataFlowSubscriptionDetails.InstallPlanApproval),
			CatalogSource:          sonataFlowSubscriptionDetails.SourceName,
			StartingCSV:            sonataFlowSubscriptionDetails.StartingCSV,
			CatalogSourceNamespace: "openshift-marketplace",
			Package:                sonataFlowSubscriptionDetails.Name,
		},
	}
	return subscriptionObject
}

func installOperatorSubscription(
	ctx context.Context, client client.Client, olmClientSet olmclientset.Interface, namespace string,
	subscriptionName string, sonataFlowOperator orchestratorv1alpha1.SonataFlowOperator) error {

	logger := log.FromContext(ctx)
	logger.Info("Starting subscription installation process", "SubscriptionName", subscriptionName)

	logger.Info("Creating namespace", "Namespace", namespace)
	serverlessLogicNamespace := &corev1.Namespace{}
	// check if namespace exists
	err := client.Get(ctx, types.NamespacedName{Name: namespace}, serverlessLogicNamespace)
	if err != nil {
		if apierrors.IsNotFound(err) {
			// create new namespace
			newNamespace := &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: namespace}}
			err = client.Create(ctx, newNamespace)
			if err != nil {
				logger.Error(err, "Error occurred when creating namespace", "Namespace", namespace)
			}
		}
		logger.Error(err, "Error occurred when checking namespace exists", "Namespace", namespace)
	}
	// check operator group exists
	operatorGroupName := "openshift-serverless-logic"
	err = getOperatorGroup(ctx, client, namespace, operatorGroupName)
	if err != nil {
		logger.Error(err, "Failed to get operator group resource", "OperatorGroup", operatorGroupName)
	}
	// install subscription
	subscriptionObject := createSubscriptionObject(subscriptionName, namespace, sonataFlowOperator)
	installedSubscription, err := olmClientSet.OperatorsV1alpha1().
		Subscriptions(namespace).
		Create(context.Background(), subscriptionObject, metav1.CreateOptions{})

	if err != nil {
		logger.Error(err, "Error occurred while creating Subscription", "SubscriptionName", subscriptionName)
	}
	// Check the Subscription's status after installation
	installedCSV := installedSubscription.Status.InstalledCSV
	if installedCSV == "" {
		logger.Info("Subscription has no installed CSV: Incorrectly installed subscription", "Subscription", subscriptionName)
	}
	// Get the ClusterServiceVersion (CSV) for the Subscription installed
	sfcsv := &operatorsv1alpha1.ClusterServiceVersion{}
	err = client.Get(ctx, types.NamespacedName{Name: installedCSV, Namespace: namespace}, sfcsv)
	if err != nil {
		logger.Error(err, "Error occurred when retrieving CSV", "ClusterServiceVersion", installedCSV)

	}
	// Check if the CSV's phase is "Succeeded"
	if sfcsv.Status.Phase == operatorsv1alpha1.CSVPhaseSucceeded {
		logger.Info("Successfully installed Operator Subscription", "SubscriptionName", installedSubscription.Name)
		return nil
	}
	logger.Info("Successfully installed Operator Subscription", "SubscriptionName", installedSubscription.Name)
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
