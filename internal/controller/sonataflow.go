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
	orchestratorv1alpha1 "github.com/parodos-dev/orchestrator-operator/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

const (
	SonataFlowAPIVersion             = "sonataflow.org/v1alpha08"
	SonataFlowPlatformCRName         = "sonataflow-platform"
	SonataFlowCRNamespace            = "sonataflow-infra"
	SonataFlowPlatformKind           = "SonataFlowPlatform"
	SonataFlowClusterPlatformKind    = "SonataFlowClusterPlatform"
	SonataFlowClusterPlatformCRName  = "cluster-platform"
	SonataFlowClusterPlatformCRDName = "sonataflowclusterplatforms.sonataflow.org"
)

func getSonataFlowPersistence(orchestrator *orchestratorv1alpha1.Orchestrator) *sonataapi.PersistenceOptionsSpec {
	return &sonataapi.PersistenceOptionsSpec{
		PostgreSQL: &sonataapi.PersistencePostgreSQL{
			SecretRef: sonataapi.PostgreSQLSecretOptions{
				Name:        orchestrator.Spec.PostgresDB.AuthSecret.SecretName,
				UserKey:     orchestrator.Spec.PostgresDB.AuthSecret.UserKey,
				PasswordKey: orchestrator.Spec.PostgresDB.AuthSecret.PasswordKey,
			},
			ServiceRef: &sonataapi.PostgreSQLServiceOptions{
				SQLServiceOptions: &sonataapi.SQLServiceOptions{
					Name:         orchestrator.Spec.PostgresDB.ServiceName,
					Namespace:    orchestrator.Spec.PostgresDB.ServiceNameSpace,
					DatabaseName: orchestrator.Spec.PostgresDB.DatabaseName,
				},
			},
		},
	}
}

func handleSonataFlowClusterCR(ctx context.Context, client client.Client, crName string) error {
	logger := log.FromContext(ctx)
	// check sonataflowlusterplatform CR exists
	sfcCR := &sonataapi.SonataFlowClusterPlatform{}

	err := client.Get(ctx, types.NamespacedName{Name: crName, Namespace: SonataFlowCRNamespace}, sfcCR)
	if err == nil {
		// CR exists; check for CR updates TODO
		logger.Info("CR resource  found.", "CR-Name", crName, "Namespace", SonataFlowCRNamespace)
		return nil
	}

	// Create sonataflow cluster CR object
	sonataFlowClusterCR := &sonataapi.SonataFlowClusterPlatform{
		TypeMeta: metav1.TypeMeta{
			APIVersion: SonataFlowAPIVersion,          // CRD group and version
			Kind:       SonataFlowClusterPlatformKind, // CRD kind
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      crName,             // Name of the CR
			Namespace: "sonataflow-infra", // Namespace of the CR
		},
		Spec: sonataapi.SonataFlowClusterPlatformSpec{
			PlatformRef: sonataapi.SonataFlowPlatformRef{
				Name:      SonataFlowPlatformCRName,
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

func createSonataFlowPlatformCR(
	ctx context.Context,
	client client.Client,
	orchestrator *orchestratorv1alpha1.Orchestrator,
	crName string) error {
	logger := log.FromContext(ctx)

	logger.Info("Starting CR creation for SonataFlowPlatform...")
	logger.Info("printing...", "orchestrator spec postgres db", orchestrator.Spec.PostgresDB)

	sfpCR := &sonataapi.SonataFlowPlatform{}
	err := client.Get(ctx, types.NamespacedName{
		Namespace: SonataFlowCRNamespace,
		Name:      SonataFlowPlatformCRName,
	}, sfpCR)

	if err != nil {
		if apierrors.IsNotFound(err) {
			logger.Info("SonataFlowPlatform not found. Proceed to creating CR...")

			// Create sonataflow platform CR object
			limitResourceMap := make(map[corev1.ResourceName]resource.Quantity)

			cpuQuantity, _ := resource.ParseQuantity(orchestrator.Spec.OrchestratorPlatform.SonataFlowPlatform.Resources.Limits.Cpu)
			memoryQuantity, _ := resource.ParseQuantity(orchestrator.Spec.OrchestratorPlatform.SonataFlowPlatform.Resources.Limits.Memory)
			limitResourceMap[corev1.ResourceCPU] = cpuQuantity
			limitResourceMap[corev1.ResourceMemory] = memoryQuantity
			//logger.Info("Limit Map", "Map", limitResourceMap)

			requestResourceMap := make(map[corev1.ResourceName]resource.Quantity)
			requestCpuQuantity, _ := resource.ParseQuantity(orchestrator.Spec.OrchestratorPlatform.SonataFlowPlatform.Resources.Requests.Cpu)
			requestMemoryQuantity, _ := resource.ParseQuantity(orchestrator.Spec.OrchestratorPlatform.SonataFlowPlatform.Resources.Requests.Memory)
			requestResourceMap[corev1.ResourceCPU] = requestCpuQuantity
			requestResourceMap[corev1.ResourceMemory] = requestMemoryQuantity
			//logger.Info("Request Map", "Map", requestResourceMap)

			var enabled = true
			sonataFlowPlatformCR := &sonataapi.SonataFlowPlatform{
				TypeMeta: metav1.TypeMeta{
					APIVersion: SonataFlowAPIVersion,   // Replace with your CRD group and version
					Kind:       SonataFlowPlatformKind, // Replace with your CRD kind
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      SonataFlowPlatformCRName, // Name of the CR
					Namespace: SonataFlowCRNamespace,    // Namespace of the CR
				},
				Spec: sonataapi.SonataFlowPlatformSpec{
					Build: sonataapi.BuildPlatformSpec{
						Template: sonataapi.BuildTemplate{
							Resources: corev1.ResourceRequirements{
								Limits:   corev1.ResourceList(limitResourceMap),
								Requests: corev1.ResourceList(requestResourceMap),
							},
						}},
					Services: &sonataapi.ServicesPlatformSpec{
						DataIndex: &sonataapi.ServiceSpec{
							Enabled:     &enabled,
							Persistence: getSonataFlowPersistence(orchestrator),
							//PodTemplate: sonataapi.PodTemplateSpec{},
						},
						JobService: &sonataapi.ServiceSpec{
							Enabled:     &enabled,
							Persistence: getSonataFlowPersistence(orchestrator),
							//PodTemplate: sonataapi.PodTemplateSpec{},
						},
					},
				},
			}
			logger.Info("Persistence function", "Persistent", getSonataFlowPersistence(orchestrator))
			// Create sonataflow platform CR
			if err := client.Create(ctx, sonataFlowPlatformCR); err != nil {
				logger.Error(err, "Failed to create Custom Resource", "CR-Name", crName)
				return err
			}
			logger.Info("Successfully created CR", "CR-Name", sonataFlowPlatformCR.Name)
		}
	}
	return nil
}

func checkSonataFlowCRExists() {

}
