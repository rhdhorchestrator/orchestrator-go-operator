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
	operatorsv1 "github.com/operator-framework/api/pkg/operators/v1"
	"github.com/operator-framework/api/pkg/operators/v1alpha1"
	operatorsv1alpha1 "github.com/operator-framework/api/pkg/operators/v1alpha1"
	olmclientset "github.com/operator-framework/operator-lifecycle-manager/pkg/api/client/clientset/versioned"
	orchestratorv1alpha1 "github.com/parodos-dev/orchestrator-operator/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

func installOperatorViaSubscription(
	ctx context.Context, client client.Client, olmClientSet olmclientset.Interface, namespace string,
	subscriptionName string, sonataFlowOperator orchestratorv1alpha1.SonataFlowOperator) error {

	logger := log.FromContext(ctx)
	logger.Info("Starting subscription installation process", "SubscriptionName", subscriptionName)

	logger.Info("Creating namespace", "Namespace", namespace)
	namespaceObj := &corev1.Namespace{}
	// check if namespace exists
	err := client.Get(ctx, types.NamespacedName{Name: namespace}, namespaceObj)
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
