//
// Copyright (c) 2024 Red Hat, Inc.
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package controller

import (
	"context"
	sonataapi "github.com/apache/incubator-kie-kogito-serverless-operator/api/v1alpha08"
	olmclientset "github.com/operator-framework/operator-lifecycle-manager/pkg/api/client/clientset/versioned"
	orchestratorv1alpha2 "github.com/parodos-dev/orchestrator-operator/api/v1alpha2"
	"github.com/parodos-dev/orchestrator-operator/internal/controller/kube"
	"github.com/parodos-dev/orchestrator-operator/internal/controller/util"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

const (
	sonataFlowAPIVersion                   = "sonataflow.org/v1alpha08"
	sonataFlowPlatformCRName               = "sonataflow-platform"
	sonataFlowPlatformKind                 = "SonataFlowPlatform"
	sonataFlowClusterPlatformKind          = "SonataFlowClusterPlatform"
	sonataFlowClusterPlatformCRName        = "cluster-platform"
	sonataFlowClusterPlatformCRDName       = "sonataflowclusterplatforms.sonataflow.org"
	serverlessOperatorGroupName            = "serverless-operator-group"
	serverlessLogicSubscriptionChannel     = "alpha"
	serverlessLogicOperatorNamespace       = "openshift-serverless-logic"
	serverlessLogicSubscriptionName        = "logic-operator-rhel8"
	serverlessLogicSubscriptionStartingCSV = "logic-operator-rhel8.v1.34.0"
)

func handleServerlessLogicOperatorInstallation(ctx context.Context, client client.Client, olmClientSet olmclientset.Clientset) error {
	sfLogger := log.FromContext(ctx)

	// create namespace for operator
	if _, err := kube.CheckNamespaceExist(ctx, client, serverlessLogicOperatorNamespace); err != nil {
		if apierrors.IsNotFound(err) {
			if err := kube.CreateNamespace(ctx, client, serverlessLogicOperatorNamespace); err != nil {
				sfLogger.Error(err, "Error occurred when creating namespace for Serverless Logic operator", "NS", serverlessLogicOperatorNamespace)
				return nil
			}
		}
		sfLogger.Error(err, "Error occurred when checking namespace exist for Serverless Logic operator", "NS", serverlessLogicOperatorNamespace)
		return err
	}

	// check if subscription exist
	oslSubscription := kube.CreateSubscriptionObject(
		serverlessLogicSubscriptionName,
		serverlessLogicOperatorNamespace,
		serverlessLogicSubscriptionChannel,
		serverlessLogicSubscriptionStartingCSV)

	subscriptionExists, existingSubscription, err := kube.CheckSubscriptionExists(ctx, olmClientSet, oslSubscription)
	if err != nil {
		sfLogger.Error(err, "Error occurred when checking subscription exists", "SubscriptionName", serverlessLogicSubscriptionName)
		return err
	}
	if !subscriptionExists {
		err := kube.InstallSubscriptionAndOperatorGroup(
			ctx, client, olmClientSet,
			serverlessOperatorGroupName,
			oslSubscription)
		if err != nil {
			sfLogger.Error(err, "Error occurred when installing operator via Subscription", "SubscriptionName", serverlessLogicSubscriptionName)
			return err
		}
		sfLogger.Info("Operator successfully installed via Subscription", "SubscriptionName", serverlessLogicSubscriptionName)
	} else {
		// Compare the current and desired state
		if !reflect.DeepEqual(existingSubscription.Spec, oslSubscription.Spec) {
			// Update the existing subscription with the desired spec
			existingSubscription.Spec = oslSubscription.Spec
			if err := client.Update(ctx, existingSubscription); err != nil {
				sfLogger.Error(err, "Error occurred when updating subscription spec", "SubscriptionName", serverlessLogicSubscriptionName)
				return err
			}
			sfLogger.Info("Successfully updated updating subscription spec", "SubscriptionName", serverlessLogicSubscriptionName)
		}
	}

	// approve install plan
	if existingSubscription.Status.InstallPlanRef != nil && existingSubscription.Status.CurrentCSV == serverlessLogicSubscriptionStartingCSV {
		installPlanName := existingSubscription.Status.InstallPlanRef.Name
		if err := kube.ApproveInstallPlan(client, ctx, installPlanName, existingSubscription.Namespace); err != nil {
			sfLogger.Error(err, "Error occurred while approving install plan for subscription", "SubscriptionName", installPlanName)
			return err
		}
	}
	return nil
}

// handleServerlessLogicCR performs the creation of serverless logic namespace and CRs
func handleServerlessLogicCR(ctx context.Context, client client.Client, orchestrator *orchestratorv1alpha2.Orchestrator) error {
	sfLogger := log.FromContext(ctx)
	sfLogger.Info("Handling ServerlessLogic CR...")
	serverlessWorkflowNamespace := orchestrator.Spec.PlatformConfig.Namespace

	// check namespace for workflow
	if _, err := kube.CheckNamespaceExist(ctx, client, serverlessWorkflowNamespace); err != nil {
		if apierrors.IsNotFound(err) {
			sfLogger.Info("Workflow namespace does not exist. Please create workflow namespace", "NS", serverlessWorkflowNamespace)
		}
		sfLogger.Error(err, "Error occurred when checking namespace exist for Workflow operator", "NS", serverlessWorkflowNamespace)
		return err
	}

	if err := handleSonataFlowClusterCR(ctx, client, sonataFlowClusterPlatformCRName, serverlessWorkflowNamespace); err != nil {
		sfLogger.Error(err, "Error occurred when creating SonataFlowClusterCR", "CR-Name", sonataFlowClusterPlatformCRName)
		return err

	}
	// create sonataflowplatform  CR
	if err := handleSonataFlowPlatformCR(ctx, client, orchestrator, sonataFlowClusterPlatformCRName, serverlessWorkflowNamespace); err != nil {
		sfLogger.Error(err, "Error occurred when creating SonataFlowPlatform", "CR-Name", sonataFlowClusterPlatformCRName)
		return err
	}
	return nil
}

func getServerlessLogicPersistence(orchestrator *orchestratorv1alpha2.Orchestrator) *sonataapi.PersistenceOptionsSpec {
	return &sonataapi.PersistenceOptionsSpec{
		PostgreSQL: &sonataapi.PersistencePostgreSQL{
			SecretRef: sonataapi.PostgreSQLSecretOptions{
				Name:        orchestrator.Spec.PostgresConfig.AuthSecret.SecretName,
				UserKey:     orchestrator.Spec.PostgresConfig.AuthSecret.UserKey,
				PasswordKey: orchestrator.Spec.PostgresConfig.AuthSecret.PasswordKey,
			},
			ServiceRef: &sonataapi.PostgreSQLServiceOptions{
				SQLServiceOptions: &sonataapi.SQLServiceOptions{
					Name:         orchestrator.Spec.PostgresConfig.Name,
					Namespace:    orchestrator.Spec.PostgresConfig.Namespace,
					DatabaseName: orchestrator.Spec.PostgresConfig.DatabaseName,
				},
			},
		},
	}
}

func handleSonataFlowClusterCR(ctx context.Context, client client.Client, crName, namespace string) error {
	logger := log.FromContext(ctx)
	// check sonataflowlusterplatform CR exists
	sfcCR := &sonataapi.SonataFlowClusterPlatform{}

	err := client.Get(ctx, types.NamespacedName{Name: crName, Namespace: namespace}, sfcCR)
	if err == nil {
		// CR exists; check for CR updates
		logger.Info("CR resource  found.", "CR-Name", crName, "NS", namespace)
		sfcCR.Spec = getSonataFlowClusterSpec(namespace)
		if err = client.Update(ctx, sfcCR); err != nil {
			logger.Error(err, "Failed to update CR", "CR-Name", sfcCR.Name)
		}
		return nil
	} else {
		if apierrors.IsNotFound(err) {
			// Create sonataflowcluster CR object
			sonataFlowClusterCR := &sonataapi.SonataFlowClusterPlatform{
				TypeMeta: metav1.TypeMeta{
					APIVersion: sonataFlowAPIVersion,
					Kind:       sonataFlowClusterPlatformKind,
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      sonataFlowClusterPlatformCRName,
					Namespace: namespace,
					Labels:    kube.AddLabel(),
				},
				Spec: getSonataFlowClusterSpec(namespace),
			}

			// Create sonataflow cluster CR
			if err := client.Create(ctx, sonataFlowClusterCR); err != nil {
				logger.Error(err, "Error occurred when creating Custom Resource", "CR-Name", crName)
				return err
			}
			logger.Info("Successfully created SonataFlowClusterPlatform resource", "CR-Name", sonataFlowClusterCR.Name)
			return nil
		}
		logger.Error(err, "Error occurred when retrieving SonataFlowClusterPlatform CR", "CR-Name", crName)
	}
	return err
}

func getSonataFlowClusterSpec(namespace string) sonataapi.SonataFlowClusterPlatformSpec {
	return sonataapi.SonataFlowClusterPlatformSpec{
		PlatformRef: sonataapi.SonataFlowPlatformRef{
			Name:      sonataFlowClusterPlatformCRName,
			Namespace: namespace,
		},
	}
}

func handleSonataFlowPlatformCR(
	ctx context.Context, client client.Client,
	orchestrator *orchestratorv1alpha2.Orchestrator, crName, namespace string) error {
	logger := log.FromContext(ctx)

	logger.Info("Starting CR creation for SonataFlowPlatform...")

	sfpCR := &sonataapi.SonataFlowPlatform{}
	err := client.Get(ctx, types.NamespacedName{
		Namespace: namespace,
		Name:      sonataFlowPlatformCRName,
	}, sfpCR)

	if err == nil {
		// CR exists; check for CR updates
		logger.Info("CR resource  found.", "CR-Name", crName, "Namespace", namespace)
		err = client.Update(ctx, sfpCR)

		return nil
	} else {
		if apierrors.IsNotFound(err) {
			logger.Info("SonataFlowPlatform not found. Proceed to creating CR...")

			// Create sonataflow platform CR object
			sonataFlowPlatformCR := &sonataapi.SonataFlowPlatform{
				TypeMeta: metav1.TypeMeta{
					APIVersion: sonataFlowAPIVersion,
					Kind:       sonataFlowPlatformKind,
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      sonataFlowPlatformCRName,
					Namespace: namespace,
					Labels:    kube.AddLabel(),
				},
				Spec: getSonataFlowPlatformSpec(orchestrator),
			}
			logger.Info("Persistence function", "Persistent", getServerlessLogicPersistence(orchestrator))
			// Create sonataflowplatform CR
			if err := client.Create(ctx, sonataFlowPlatformCR); err != nil {
				logger.Error(err, "Failed to create Custom Resource", "CR-Name", crName)
				return err
			}
			logger.Info("Successfully created SonataFlowPlatform CR", "CR-Name", sonataFlowPlatformCR.Name)
			return nil
		}
		logger.Error(err, "Error occurred when retrieving SonataFlowPlatform CR", "CR-Name", crName)
	}
	return err
}

func getSonataFlowPlatformSpec(orchestrator *orchestratorv1alpha2.Orchestrator) sonataapi.SonataFlowPlatformSpec {
	limitResourceMap := make(map[corev1.ResourceName]resource.Quantity)

	cpuQuantity, _ := resource.ParseQuantity(orchestrator.Spec.PlatformConfig.Resources.Limits.Cpu)
	memoryQuantity, _ := resource.ParseQuantity(orchestrator.Spec.PlatformConfig.Resources.Limits.Memory)
	limitResourceMap[corev1.ResourceCPU] = cpuQuantity
	limitResourceMap[corev1.ResourceMemory] = memoryQuantity

	requestResourceMap := make(map[corev1.ResourceName]resource.Quantity)
	requestCpuQuantity, _ := resource.ParseQuantity(orchestrator.Spec.PlatformConfig.Resources.Requests.Cpu)
	requestMemoryQuantity, _ := resource.ParseQuantity(orchestrator.Spec.PlatformConfig.Resources.Requests.Memory)
	requestResourceMap[corev1.ResourceCPU] = requestCpuQuantity
	requestResourceMap[corev1.ResourceMemory] = requestMemoryQuantity

	return sonataapi.SonataFlowPlatformSpec{
		Build: sonataapi.BuildPlatformSpec{
			Template: sonataapi.BuildTemplate{
				Resources: corev1.ResourceRequirements{
					Limits:   limitResourceMap,
					Requests: requestResourceMap,
				},
			}},
		Services: &sonataapi.ServicesPlatformSpec{
			DataIndex: &sonataapi.ServiceSpec{
				Enabled:     util.MakePointer(true),
				Persistence: getServerlessLogicPersistence(orchestrator),
			},
			JobService: &sonataapi.ServiceSpec{
				Enabled:     util.MakePointer(true),
				Persistence: getServerlessLogicPersistence(orchestrator),
			},
		},
	}
}

func handleServerlessLogicCleanUp(ctx context.Context, client client.Client, olmClientSet olmclientset.Clientset, namespace string) error {
	logger := log.FromContext(ctx)

	// remove all namespace
	if err := kube.CleanUpNamespace(ctx, namespace, client); err != nil {
		logger.Error(err, "Error occurred when deleting namespace", "NS", namespace)
		return err
	}
	oslSubscription := kube.CreateSubscriptionObject(
		serverlessLogicSubscriptionName,
		serverlessLogicOperatorNamespace,
		serverlessLogicSubscriptionChannel,
		serverlessLogicSubscriptionStartingCSV)

	if err := kube.CleanUpSubscriptionAndCSV(ctx, olmClientSet, oslSubscription); err != nil {
		logger.Error(err, "Error occurred when deleting Subscription and CSV", "Subscription", serverlessLogicSubscriptionName)
		return err
	}
	// remove all CRDs, optional (ensure all CRs and namespace have been removed first)
	return nil
}
