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

package knative

import (
	"context"
	"reflect"

	olmclientset "github.com/operator-framework/operator-lifecycle-manager/pkg/api/client/clientset/versioned"
	"github.com/rhdhorchestrator/orchestrator-operator/internal/controller/kube"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	Knative "knative.dev/operator/pkg/apis/operator/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

const (
	KnativeAPIVersion              = "operator.knative.dev/v1beta1"
	KnativeServingKind             = "KnativeServing"
	KnativeServingNamespacedName   = "knative-serving"
	KnativeEventingKind            = "KnativeEventing"
	KnativeEventingNamespacedName  = "knative-eventing"
	KnativeEventingCRDName         = "knativeeventings.operator.knative.dev"
	KnativeServingCRDName          = "knativeservings.operator.knative.dev"
	KnativeOperatorGroupName       = "serverless-operator-group"
	KnativeSubscriptionName        = "serverless-operator"
	KnativeOperatorNamespace       = "openshift-serverless"
	KnativeSubscriptionChannel     = "stable"
	KnativeSubscriptionStartingCSV = "serverless-operator.v1.35.0"
)

func HandleKNativeOperatorInstallation(ctx context.Context, client client.Client, olmClientSet olmclientset.Interface) error {
	KnativeLogger := log.FromContext(ctx)

	if _, err := kube.CheckNamespaceExist(ctx, client, KnativeOperatorNamespace); err != nil {
		if apierrors.IsNotFound(err) {
			KnativeLogger.Info("Creating namespace", "NS", KnativeOperatorNamespace)
			if err := kube.CreateNamespace(ctx, client, KnativeOperatorNamespace); err != nil {
				KnativeLogger.Error(err, "Error occurred when creating namespace", "NS", KnativeOperatorNamespace)
				return err
			}
		}
	}

	serverlessSubscription := kube.CreateSubscriptionObject(
		KnativeSubscriptionName,
		KnativeOperatorNamespace,
		KnativeSubscriptionChannel,
		KnativeSubscriptionStartingCSV)

	// check if subscription exists
	subscriptionExists, existingSubscription, err := kube.CheckSubscriptionExists(ctx, olmClientSet, serverlessSubscription)
	if err != nil {
		KnativeLogger.Error(err, "Error occurred when checking subscription exists", "SubscriptionName", KnativeSubscriptionName)
		return err
	}
	if !subscriptionExists {
		if err := kube.InstallSubscriptionAndOperatorGroup(
			ctx, client, olmClientSet,
			KnativeOperatorGroupName, serverlessSubscription); err != nil {
			KnativeLogger.Error(err, "Error occurred when installing operator", "SubscriptionName", KnativeSubscriptionName)
			return err
		}
		KnativeLogger.Info("Operator successfully installed", "SubscriptionName", KnativeSubscriptionName)
	} else {
		// Compare the current and desired state
		if !reflect.DeepEqual(existingSubscription.Spec, serverlessSubscription.Spec) {
			// Update the existing subscription with the new Spec
			existingSubscription.Spec = serverlessSubscription.Spec
			if err := client.Update(ctx, existingSubscription); err != nil {
				KnativeLogger.Error(err, "Error occurred when updating subscription spec", "SubscriptionName", KnativeSubscriptionName)
				return err
			}
			KnativeLogger.Info("Successfully updated updating subscription spec", "SubscriptionName", KnativeSubscriptionName)
		}
	}

	// approve install plan
	if existingSubscription.Status.InstallPlanRef != nil && existingSubscription.Status.CurrentCSV == KnativeSubscriptionStartingCSV {
		installPlanName := existingSubscription.Status.InstallPlanRef.Name
		if err := kube.ApproveInstallPlan(client, ctx, installPlanName, existingSubscription.Namespace); err != nil {
			KnativeLogger.Error(err, "Error occurred while approving install plan for subscription", "SubscriptionName", installPlanName)
			return err
		}
	}

	return nil
}

func HandleKnativeCR(ctx context.Context, client client.Client) error {
	KnativeLogger := log.FromContext(ctx)
	KnativeLogger.Info("Handling Serverless Custom Resources...")

	// subscription exists; check if CRD exists for Knative eventing;
	if err := kube.CheckCRDExists(ctx, client, KnativeEventingCRDName); err != nil {
		if apierrors.IsNotFound(err) {
			KnativeLogger.Info("CRD resource not found or ready", "SubscriptionName", KnativeSubscriptionName)
			return err
		}
		KnativeLogger.Error(err, "Error occurred when retrieving CRD", "CRD", KnativeEventingCRDName)
		return err
	}
	// CRD exists; check and handle Knative eventing CR
	if err := HandleKnativeEventingCR(ctx, client); err != nil {
		KnativeLogger.Error(err, "Error occurred when creating Knative EventingCR", "CR-Name", KnativeEventingNamespacedName)
		return err
	}

	// subscription exists; check if Knative serving CRD exists;
	if err := kube.CheckCRDExists(ctx, client, KnativeServingCRDName); err != nil {
		if apierrors.IsNotFound(err) {
			KnativeLogger.Info("CRD resource not found or ready", "SubscriptionName", KnativeSubscriptionName)
			return err
		}
		KnativeLogger.Error(err, "Error occurred when retrieving CRD", "CRD", KnativeServingCRDName)
		return err
	}
	// CRD exist; check and handle Knative serving CR
	if err := HandleKnativeServingCR(ctx, client); err != nil {
		KnativeLogger.Error(err, "Error occurred when creating Knative ServingCR", "CR-Name", KnativeServingNamespacedName)
		return err
	}
	return nil
}

func HandleKnativeEventingCR(ctx context.Context, client client.Client) error {
	logger := log.FromContext(ctx)
	logger.Info("Handling K-Native Eventing CR")

	// check namespace exist; else create namespace
	namespaceExist, _ := kube.CheckNamespaceExist(ctx, client, KnativeEventingNamespacedName)
	if !namespaceExist {
		if err := kube.CreateNamespace(ctx, client, KnativeEventingNamespacedName); err != nil {
			logger.Error(err, "Error occurred when creating namespace", "NS", KnativeEventingNamespacedName)
			return err
		}
	}

	desiredKnEventingCR := &Knative.KnativeEventing{
		TypeMeta: metav1.TypeMeta{
			APIVersion: KnativeAPIVersion,
			Kind:       KnativeEventingKind,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      KnativeEventingNamespacedName,
			Namespace: KnativeEventingNamespacedName,
			Labels:    kube.GetOrchestratorLabel(),
		},
		Spec: Knative.KnativeEventingSpec{},
	}
	currentKnEventingCR := &Knative.KnativeEventing{}

	// check CR exists
	err := client.Get(ctx, types.NamespacedName{Name: KnativeEventingNamespacedName, Namespace: KnativeEventingNamespacedName}, currentKnEventingCR)

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

func HandleKnativeServingCR(ctx context.Context, client client.Client) error {
	logger := log.FromContext(ctx)
	logger.Info("Handling K-Native Serving CR")

	// check namespace exist; else create namespace
	namespaceExist, _ := kube.CheckNamespaceExist(ctx, client, KnativeServingNamespacedName)
	if !namespaceExist {
		if err := kube.CreateNamespace(ctx, client, KnativeServingNamespacedName); err != nil {
			logger.Error(err, "Error occurred when creating namespace", "NS", KnativeEventingNamespacedName)
			return err
		}
	}

	desiredKnServingCR := &Knative.KnativeServing{
		TypeMeta: metav1.TypeMeta{
			APIVersion: KnativeAPIVersion,
			Kind:       KnativeServingKind,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      KnativeServingNamespacedName,
			Namespace: KnativeServingNamespacedName,
			Labels:    kube.GetOrchestratorLabel(),
		},
		Spec: Knative.KnativeServingSpec{},
	}
	currentKnServingCR := &Knative.KnativeServing{}

	// check CR exists
	err := client.Get(ctx, types.NamespacedName{Name: KnativeServingNamespacedName, Namespace: KnativeServingNamespacedName}, currentKnServingCR)

	if err != nil {
		// CR does not exist. Create CR
		if apierrors.IsNotFound(err) {
			if err = client.Create(ctx, desiredKnServingCR); err != nil {
				logger.Error(err, "Error occurred when creating CR resource", "CR-Name", desiredKnServingCR.Name)
			}
			logger.Info("Successfully created Knative Eventing resource", "CR-Name", desiredKnServingCR.Name)
			return nil
		}
		logger.Error(err, "Error occurred when checking CR resource exist", "CR-Name", desiredKnServingCR.Name)
		return err
	}
	return nil
}

func HandleKnativeCleanUp(ctx context.Context, client client.Client) error {
	logger := log.FromContext(ctx)
	// remove all namespace
	if err := kube.CleanUpNamespace(ctx, KnativeEventingNamespacedName, client); err != nil {
		logger.Error(err, "Error occurred when deleting namespace", "NS", KnativeEventingNamespacedName)
		return err
	}
	if err := kube.CleanUpNamespace(ctx, KnativeServingNamespacedName, client); err != nil {
		logger.Error(err, "Error occurred when deleting namespace", "NS", KnativeServingNamespacedName)
		return err
	}

	// remove operator namespace
	if err := kube.CleanUpNamespace(ctx, KnativeOperatorNamespace, client); err != nil {
		logger.Error(err, "Error occurred when deleting namespace", "NS", KnativeOperatorNamespace)
		return err
	}

	return nil
}
