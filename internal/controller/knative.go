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
	olmclientset "github.com/operator-framework/operator-lifecycle-manager/pkg/api/client/clientset/versioned"
	"github.com/rhdhorchestrator/orchestrator-operator/internal/controller/kube"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	knative "knative.dev/operator/pkg/apis/operator/v1beta1"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

const (
	knativeAPIVersion              = "operator.knative.dev/v1beta1"
	knativeServingKind             = "KnativeServing"
	knativeServingNamespacedName   = "knative-serving"
	knativeEventingKind            = "KnativeEventing"
	knativeEventingNamespacedName  = "knative-eventing"
	knativeEventingCRDName         = "knativeeventings.operator.knative.dev"
	knativeServingCRDName          = "knativeservings.operator.knative.dev"
	knativeOperatorGroupName       = "serverless-operator-group"
	knativeSubscriptionName        = "serverless-operator"
	knativeOperatorNamespace       = "openshift-serverless"
	knativeSubscriptionChannel     = "stable"
	knativeSubscriptionStartingCSV = "serverless-operator.v1.35.1"
)

func handleKNativeOperatorInstallation(ctx context.Context, client client.Client, olmClientSet olmclientset.Interface) error {
	knativeLogger := log.FromContext(ctx)

	if _, err := kube.CheckNamespaceExist(ctx, client, knativeOperatorNamespace); err != nil {
		if apierrors.IsNotFound(err) {
			knativeLogger.Info("Creating namespace", "NS", knativeOperatorNamespace)
			if err := kube.CreateNamespace(ctx, client, knativeOperatorNamespace); err != nil {
				knativeLogger.Error(err, "Error occurred when creating namespace", "NS", knativeOperatorNamespace)
				return nil
			}
		}
		knativeLogger.Error(err, "Error occurred when checking namespace exist", "NS", knativeOperatorNamespace)
		return err
	}

	serverlessSubscription := kube.CreateSubscriptionObject(
		knativeSubscriptionName,
		knativeOperatorNamespace,
		knativeSubscriptionChannel,
		knativeSubscriptionStartingCSV)

	// check if subscription exists
	subscriptionExists, existingSubscription, err := kube.CheckSubscriptionExists(ctx, olmClientSet, serverlessSubscription)
	if err != nil {
		knativeLogger.Error(err, "Error occurred when checking subscription exists", "SubscriptionName", knativeSubscriptionName)
		return err
	}
	if !subscriptionExists {
		if err := kube.InstallSubscriptionAndOperatorGroup(
			ctx, client, olmClientSet,
			knativeOperatorGroupName, serverlessSubscription); err != nil {
			knativeLogger.Error(err, "Error occurred when installing operator", "SubscriptionName", knativeSubscriptionName)
			return err
		}
		knativeLogger.Info("Operator successfully installed", "SubscriptionName", knativeSubscriptionName)
	} else {
		// Compare the current and desired state
		if !reflect.DeepEqual(existingSubscription.Spec, serverlessSubscription.Spec) {
			// Update the existing subscription with the new Spec
			existingSubscription.Spec = serverlessSubscription.Spec
			if err := client.Update(ctx, existingSubscription); err != nil {
				knativeLogger.Error(err, "Error occurred when updating subscription spec", "SubscriptionName", knativeSubscriptionName)
				return err
			}
			knativeLogger.Info("Successfully updated updating subscription spec", "SubscriptionName", knativeSubscriptionName)
		}
	}

	// approve install plan
	if existingSubscription.Status.InstallPlanRef != nil && existingSubscription.Status.CurrentCSV == knativeSubscriptionStartingCSV {
		installPlanName := existingSubscription.Status.InstallPlanRef.Name
		if err := kube.ApproveInstallPlan(client, ctx, installPlanName, existingSubscription.Namespace); err != nil {
			knativeLogger.Error(err, "Error occurred while approving install plan for subscription", "SubscriptionName", installPlanName)
			return err
		}
	}
	return nil
}

func handleKnativeCR(ctx context.Context, client client.Client) error {
	knativeLogger := log.FromContext(ctx)
	knativeLogger.Info("Handling Serverless Custom Resources...")

	// subscription exists; check if CRD exists for knative eventing;
	if err := kube.CheckCRDExists(ctx, client, knativeEventingCRDName); err != nil {
		if apierrors.IsNotFound(err) {
			knativeLogger.Info("CRD resource not found or ready", "SubscriptionName", knativeSubscriptionName)
			return err
		}
		knativeLogger.Error(err, "Error occurred when retrieving CRD", "CRD", knativeEventingCRDName)
		return err
	}
	// CRD exist; check and handle knative eventing CR
	if err := handleKnativeEventingCR(ctx, client); err != nil {
		knativeLogger.Error(err, "Error occurred when creating Knative EventingCR", "CR-Name", knativeEventingNamespacedName)
		return err
	}

	// subscription exists; check if CRD exists knative serving;
	if err := kube.CheckCRDExists(ctx, client, knativeServingCRDName); err != nil {
		if apierrors.IsNotFound(err) {
			knativeLogger.Info("CRD resource not found or ready", "SubscriptionName", knativeSubscriptionName)
			return err
		}
		knativeLogger.Error(err, "Error occurred when retrieving CRD", "CRD", knativeServingCRDName)
		return err
	}
	// CRD exist; check and handle knative eventing CR
	if err := handleKnativeServingCR(ctx, client); err != nil {
		knativeLogger.Error(err, "Error occurred when creating Knative ServingCR", "CR-Name", knativeServingNamespacedName)
		return err
	}
	return nil
}

func handleKnativeEventingCR(ctx context.Context, client client.Client) error {
	logger := log.FromContext(ctx)
	logger.Info("Handling K-Native Eventing CR")

	// check namespace exist; else create namespace
	namespaceExist, _ := kube.CheckNamespaceExist(ctx, client, knativeEventingNamespacedName)
	if !namespaceExist {
		if err := kube.CreateNamespace(ctx, client, knativeEventingNamespacedName); err != nil {
			logger.Error(err, "Error occurred when creating namespace", "NS", knativeEventingNamespacedName)
			return err
		}
	}

	desiredKnEventingCR := &knative.KnativeEventing{
		TypeMeta: metav1.TypeMeta{
			APIVersion: knativeAPIVersion,
			Kind:       knativeEventingKind,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      knativeEventingNamespacedName,
			Namespace: knativeEventingNamespacedName,
			Labels:    kube.GetOrchestratorLabel(),
		},
		Spec: knative.KnativeEventingSpec{},
	}
	currentKnEventingCR := &knative.KnativeEventing{}

	// check CR exists
	err := client.Get(ctx, types.NamespacedName{Name: knativeEventingNamespacedName, Namespace: knativeEventingNamespacedName}, currentKnEventingCR)

	if err != nil {
		// CR does not exist. Create CR
		if apierrors.IsNotFound(err) {
			if err = client.Create(ctx, desiredKnEventingCR); err != nil {
				logger.Error(err, "Error occurred when creating CR resource", "CR-Name", desiredKnEventingCR.Name)
			}
			logger.Info("Successfully created Knative Eventing resource", "CR-Name", desiredKnEventingCR.Name)
			return nil
		}
		logger.Error(err, "Error occurred when checking CR resource exist", "CR-Name", desiredKnEventingCR.Name)
		return err
	}
	return nil
}

func handleKnativeServingCR(ctx context.Context, client client.Client) error {
	logger := log.FromContext(ctx)
	logger.Info("Handling K-Native Serving CR")

	// check namespace exist; else create namespace
	namespaceExist, _ := kube.CheckNamespaceExist(ctx, client, knativeServingNamespacedName)
	if !namespaceExist {
		if err := kube.CreateNamespace(ctx, client, knativeServingNamespacedName); err != nil {
			logger.Error(err, "Error occurred when creating namespace", "NS", knativeEventingNamespacedName)
			return err
		}
	}

	desiredKnServingCR := &knative.KnativeServing{
		TypeMeta: metav1.TypeMeta{
			APIVersion: knativeAPIVersion,
			Kind:       knativeServingKind,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      knativeServingNamespacedName,
			Namespace: knativeServingNamespacedName,
			Labels:    kube.GetOrchestratorLabel(),
		},
		Spec: knative.KnativeServingSpec{},
	}
	currentKnServingCR := &knative.KnativeServing{}

	// check CR exists
	err := client.Get(ctx, types.NamespacedName{Name: knativeServingNamespacedName, Namespace: knativeServingNamespacedName}, currentKnServingCR)

	if err != nil {
		// CR does not exist. Create CR
		if apierrors.IsNotFound(err) {
			if err = client.Create(ctx, desiredKnServingCR); err != nil {
				logger.Error(err, "Error occurred when creating CR resource", "CR-Name", desiredKnServingCR.Name)
			}
			logger.Info("Successfully created knative Eventing resource", "CR-Name", desiredKnServingCR.Name)
			return nil
		}
		logger.Error(err, "Error occurred when checking CR resource exist", "CR-Name", desiredKnServingCR.Name)
		return err
	}
	return nil
}

func handleKnativeCleanUp(ctx context.Context, client client.Client) error {
	logger := log.FromContext(ctx)
	// remove all namespace
	if err := kube.CleanUpNamespace(ctx, knativeEventingNamespacedName, client); err != nil {
		logger.Error(err, "Error occurred when deleting namespace", "NS", knativeEventingNamespacedName)
		return err
	}
	if err := kube.CleanUpNamespace(ctx, knativeServingNamespacedName, client); err != nil {
		logger.Error(err, "Error occurred when deleting namespace", "NS", knativeServingNamespacedName)
		return err
	}

	// remove operator namespace
	if err := kube.CleanUpNamespace(ctx, knativeOperatorNamespace, client); err != nil {
		logger.Error(err, "Error occurred when deleting namespace", "NS", knativeOperatorNamespace)
		return err
	}

	return nil
}
