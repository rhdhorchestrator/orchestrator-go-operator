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

package kube

import (
	"context"
	"fmt"

	operatorsv1 "github.com/operator-framework/api/pkg/operators/v1"
	"github.com/operator-framework/api/pkg/operators/v1alpha1"
	operatorsv1alpha1 "github.com/operator-framework/api/pkg/operators/v1alpha1"
	olmclientset "github.com/operator-framework/operator-lifecycle-manager/pkg/api/client/clientset/versioned"
	corev1 "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

const (
	CatalogSourceNamespace      = "openshift-marketplace"
	CatalogSourceName           = "redhat-operators"
	CatalogSourceNameSonataFlow = "sonataflow-operator-catalog" // Remove after Sonataflow Release
	CreatedByLabelKey           = "rhdh.redhat.com/created-by"
	CreatedByLabelValue         = "orchestrator"
)

func CheckNamespaceExist(ctx context.Context, client client.Client, namespace string) (bool, error) {
	nsLogger := log.FromContext(ctx)
	nsLogger.Info("Checking namespace exist", "Namespace", namespace)
	namespaceObj := &corev1.Namespace{}
	// check if namespace exists
	if err := client.Get(ctx, types.NamespacedName{Name: namespace}, namespaceObj); err != nil {
		return false, err
	}
	// check and update missing labels
	labelExist := CheckLabelExist(namespaceObj.Labels)
	if !labelExist {
		// update namespace label
		if err := updateNamespaceLabel(namespaceObj, ctx, client); err != nil {
			nsLogger.Error(err, "Error occurred when updating namespace label", "NS", namespace)
		}
	}
	return true, nil
}

func CreateNamespace(ctx context.Context, client client.Client, namespace string) error {
	nsLogger := log.FromContext(ctx)
	nsLogger.Info("Creating namespace", "Namespace", namespace)
	// create new namespace
	newNamespace := &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: namespace, Labels: GetOrchestratorLabel()}}
	err := client.Create(ctx, newNamespace)
	if err != nil {
		nsLogger.Error(err, "Error occurred when creating namespace", "Namespace", namespace)
		return err
	}
	return nil
}

func InstallSubscriptionAndOperatorGroup(
	ctx context.Context, client client.Client,
	olmClientSet olmclientset.Interface,
	operatorGroupName string,
	subscription *v1alpha1.Subscription) error {

	logger := log.FromContext(ctx)
	subscriptionName := subscription.Name
	logger.Info("Starting subscription installation process", "SubscriptionName", subscriptionName)

	namespace := subscription.Namespace

	// check operator group exists
	err := getOperatorGroup(ctx, client, namespace, operatorGroupName)
	if err != nil {
		logger.Error(err, "Error occurred when checking operator group resource", "OperatorGroup", operatorGroupName)
		return err
	}
	// create subscription
	if _, err := olmClientSet.OperatorsV1alpha1().
		Subscriptions(namespace).
		Create(ctx, subscription, metav1.CreateOptions{}); err != nil {
		logger.Error(err, "Error occurred while creating Subscription", "SubscriptionName", subscriptionName)
		return err
	}
	return nil
}

func ApproveInstallPlan(client client.Client, ctx context.Context, installPlanName, namespace string) error {
	logger := log.FromContext(ctx)
	logger.Info("Starting approval for InstallPlan...")

	// get the InstallPlan
	installPlan := &operatorsv1alpha1.InstallPlan{}
	if err := client.Get(ctx, types.NamespacedName{
		Namespace: namespace,
		Name:      installPlanName,
	}, installPlan); err != nil {
		logger.Error(err, "Error occurred when retrieving InstallPlan", "InstallPlan", installPlan.Name)
		return err
	}

	// Approve the InstallPlan if manual approval is needed
	if !installPlan.Spec.Approved {
		logger.Info("Approving InstallPlan", "InstallPlanName", installPlan.Name)
		installPlan.Spec.Approved = true
		if err := client.Update(ctx, installPlan); err != nil {
			logger.Error(err, "Error occurred when approving InstallPlan", "InstallPlanName", installPlan.Name)
			return err
		}
		logger.Info("Successfully approved InstallPlan", "InstallPlanName", installPlan.Name)
	}
	return nil
}

func CheckCSVExists(ctx context.Context, client client.Client, installedSubscription *v1alpha1.Subscription) (bool, error) {
	logger := log.FromContext(ctx)

	installedCSV := installedSubscription.Status.InstalledCSV

	// Check the Subscription's status after installation
	if installedCSV == "" {
		logger.Info("Subscription has no installed CSV: CSV not ready or Incorrectly installed subscription", "Subscription", installedSubscription.Name)
		return false, fmt.Errorf("subscription not ready yet")
	}
	// Get the ClusterServiceVersion (CSV) for the Subscription installed
	sfcsv := &operatorsv1alpha1.ClusterServiceVersion{}
	if err := client.Get(ctx, types.NamespacedName{Name: installedCSV, Namespace: installedSubscription.Namespace}, sfcsv); err != nil {
		logger.Error(err, "Error occurred when retrieving CSV", "ClusterServiceVersion", installedCSV)
		return false, err
	}
	// Check if the CSV's phase is Succeeded
	if sfcsv != nil && sfcsv.Status.Phase == operatorsv1alpha1.CSVPhaseSucceeded {
		logger.Info("Successfully installed Operator Via Subscription", "SubscriptionName", installedSubscription.Name)
		return true, nil
	}
	return false, fmt.Errorf("failed to install operator via subscription")

}

func getOperatorGroup(ctx context.Context, client client.Client,
	namespace, operatorGroupName string) error {
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
	if err = client.Create(ctx, sfog); err != nil {
		logger.Error(err, "Error occurred when creating OperatorGroup resource", "Namespace", namespace)
		return err
	}
	logger.Info("Successfully created OperatorGroup", "OperatorGroup", operatorGroupName)
	return nil
}

func CreateSubscriptionObject(subscriptionName, namespace, channel, startingCSV string) *v1alpha1.Subscription {
	logger := log.Log.WithName("subscriptionObject")
	logger.Info("Creating subscription object")

	if subscriptionName == "sonataflow-operator" {
		subscriptionObject := &v1alpha1.Subscription{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: namespace,
				Name:      subscriptionName,
				Labels:    GetOrchestratorLabel(),
			},
			Spec: &v1alpha1.SubscriptionSpec{
				Channel:                channel,
				InstallPlanApproval:    v1alpha1.ApprovalManual,
				CatalogSource:          CatalogSourceNameSonataFlow,
				StartingCSV:            startingCSV,
				CatalogSourceNamespace: CatalogSourceNamespace,
				Package:                subscriptionName,
			},
		}
		return subscriptionObject
	}

	subscriptionObject := &v1alpha1.Subscription{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      subscriptionName,
			Labels:    GetOrchestratorLabel(),
		},
		Spec: &v1alpha1.SubscriptionSpec{
			Channel:                channel,
			InstallPlanApproval:    v1alpha1.ApprovalManual,
			CatalogSource:          CatalogSourceName,
			StartingCSV:            startingCSV,
			CatalogSourceNamespace: CatalogSourceNamespace,
			Package:                subscriptionName,
		},
	}
	return subscriptionObject
}

func CheckSubscriptionExists(
	ctx context.Context, olmClientSet olmclientset.Interface,
	existingSubscription *v1alpha1.Subscription) (bool, *v1alpha1.Subscription, error) {
	logger := log.FromContext(ctx)

	namespace := existingSubscription.Namespace
	subscriptionName := existingSubscription.Name

	subscription, err := olmClientSet.OperatorsV1alpha1().Subscriptions(namespace).Get(ctx, subscriptionName, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			logger.Info("Subscription resource not found.", "SubscriptionName", subscriptionName, "Namespace", namespace)
			return false, subscription, nil
		}
		logger.Error(err, "Failed to check Subscription does not exists", "SubscriptionName", subscriptionName)
		return false, subscription, err
	}
	logger.Info("Subscription exists", "SubscriptionName", subscription.Name)
	return true, subscription, nil
}

func CheckCRDExists(ctx context.Context, client client.Client, name string) error {
	crd := &apiextensionsv1.CustomResourceDefinition{}
	err := client.Get(ctx, types.NamespacedName{Name: name}, crd)
	if err != nil {
		return err
	}
	return nil
}

func CleanUpNamespace(ctx context.Context, namespaceName string, client client.Client) error {
	logger := log.FromContext(ctx)

	// check namespace exist
	logger.Info("Checking namespace exist", "Namespace", namespaceName)
	namespaceObj := &corev1.Namespace{}
	if err := client.Get(ctx, types.NamespacedName{Name: namespaceName}, namespaceObj); err != nil {
		if apierrors.IsNotFound(err) {
			logger.Info("Namespace does not exist", "Namespace", namespaceName)
			return nil
		}
		return err
	}

	// check label exist
	labelExist := CheckLabelExist(namespaceObj.Labels)

	if labelExist {
		// delete namespace
		if err := client.Delete(ctx, namespaceObj); err != nil && !apierrors.IsNotFound(err) {
			return err
		}
		logger.Info("Successfully deleted Namespace", "Namespace", namespaceName)
	}
	return nil
}

func CleanUpSubscriptionAndCSV(ctx context.Context, olmClientSet olmclientset.Interface, subscription *v1alpha1.Subscription) error {
	logger := log.FromContext(ctx)

	subscriptionName := subscription.Name
	subscriptionNamespace := subscription.Namespace

	subscriptionExists, _, err := CheckSubscriptionExists(ctx, olmClientSet, subscription)
	if err != nil {
		logger.Error(err, "Error occurred when checking subscription exists", "SubscriptionName", subscriptionName)
		return err
	}
	if subscriptionExists {
		// get name of csv before deletion
		csvName := subscription.Status.InstalledCSV

		// deleting subscription resource
		err = olmClientSet.OperatorsV1alpha1().Subscriptions(subscriptionNamespace).Delete(ctx, subscriptionName, metav1.DeleteOptions{})
		if err != nil {
			logger.Error(err, "Error occurred while deleting Subscription", "SubscriptionName", subscriptionName, "Namespace", subscriptionNamespace)
			return err
		}
		logger.Info("Successfully deleted Subscription", "SubscriptionName", subscriptionName)

		// cleanup csv
		// TODO verify CSV deletion happens. refactor code to use csvName instead of namespace
		csv, err := olmClientSet.OperatorsV1alpha1().ClusterServiceVersions(subscriptionNamespace).Get(ctx, csvName, metav1.GetOptions{})
		if err != nil {
			if apierrors.IsNotFound(err) {
				// ensure csvName is not empty: Do a check for csvName
				logger.Info("CSV resource not found", "CSV", csvName)
				return nil
			}
			logger.Error(err, "Error occurred when getting CSV", "CSV", csvName)
			return err
		}
		if err := olmClientSet.OperatorsV1alpha1().ClusterServiceVersions(csv.Namespace).Delete(ctx, csv.Name, metav1.DeleteOptions{}); err != nil && !apierrors.IsNotFound(err) {
			return err
		}
		logger.Info("Successfully deleted CSV", "CSV", csvName)
		return nil
	}
	return err
}

func GetOrchestratorLabel() map[string]string {
	return map[string]string{
		CreatedByLabelKey: CreatedByLabelValue,
	}
}

func CheckLabelExist(labels map[string]string) bool {
	labelValue, labelExist := labels[CreatedByLabelKey]
	if !labelExist {
		return false
	}
	return labelValue == CreatedByLabelValue
}

// updateNamespaceLabel adds a new label to namespace and updates the namespace object.
func updateNamespaceLabel(namespace *corev1.Namespace, ctx context.Context, client client.Client) error {
	nsLogger := log.FromContext(ctx)

	namespaceName := namespace.Name
	nsLogger.Info("Updating namespace with new label", "NS", namespaceName)

	// add new label to namespace label map
	namespace.Labels[CreatedByLabelKey] = CreatedByLabelValue
	if err := client.Update(ctx, namespace); err != nil {
		nsLogger.Info("Error occurred when updating namespace with new label", "NS", namespaceName)
		return err
	}
	nsLogger.Info("Successfully updated namespace with new label", "NS", namespaceName)
	return nil
}

// RemoveCustomResourcesInNamespace removes orchestrator labelled CR in a given namespace
// returns error
func RemoveCustomResourcesInNamespace[T client.ObjectList, I client.Object](ctx context.Context,
	k8client client.Client,
	objList T, getItems func(T) []I,
	namespace string) error {

	logger := log.FromContext(ctx)
	logger.Info("Removing custom resources in namespace", "NS", namespace)

	listOptions := []client.ListOption{
		client.InNamespace(namespace),
		client.MatchingLabels(GetOrchestratorLabel())}

	// List the CRs
	if err := k8client.List(ctx, objList, listOptions...); err != nil {
		if apierrors.IsNotFound(err) {
			logger.Info("Custom resource not found in namespace", "NS", namespace)
			return nil
		}
		logger.Error(err, "Error occurred when listing resources")
		return err
	}
	var errorList []error
	for _, item := range getItems(objList) {
		if err := k8client.Delete(ctx, item); err != nil {
			logger.Error(err, "Error occurred when deleting custom resource", "CR-Name", item.GetName())
			errorList = append(errorList, err)
		}
	}
	if len(errorList) > 0 {
		return errors.NewAggregate(errorList)
	}
	return nil
}
