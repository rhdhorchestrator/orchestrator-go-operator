/*
Copyright 2024 Red Hat, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
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

package gitops

import (
	"context"
	argocdv1alpha1 "github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	"github.com/rhdhorchestrator/orchestrator-operator/internal/controller/kube"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

const (
	argoCDCRName     = "orchestrator-gitops"
	argoCDCRDName    = "appprojects.argoproj.io"
	argoCDAPIVersion = "argoproj.io/v1alpha1"
	argoCDKind       = "AppProject"
)

func handleArgoCDProject(gitOpsNamespace string, client client.Client, ctx context.Context) error {
	argoLogger := log.FromContext(ctx)
	argoLogger.Info("Handling ArgoCD Project...")

	if err := kube.CheckCRDExists(ctx, client, argoCDCRDName); err != nil {
		argoLogger.Error(err, "ArgoCD CRD does not exist. Install ArgoCD Operator")
		return err
	}

	desiredAppProject := &argocdv1alpha1.AppProject{
		TypeMeta: metav1.TypeMeta{
			APIVersion: argoCDAPIVersion,
			Kind:       argoCDKind,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      argoCDCRName,
			Namespace: gitOpsNamespace,
			Labels:    kube.GetOrchestratorLabel(),
		},
		Spec: argocdv1alpha1.AppProjectSpec{
			Destinations: []argocdv1alpha1.ApplicationDestination{
				{
					Name:      "*",
					Namespace: "*",
					Server:    "*",
				},
			},
			SourceRepos: []string{"*"},
		},
	}
	existingAppProject := &argocdv1alpha1.AppProject{}

	err := client.Get(ctx, types.NamespacedName{
		Namespace: gitOpsNamespace,
		Name:      argoCDCRName,
	}, existingAppProject)

	if err != nil {
		if errors.IsNotFound(err) {
			argoLogger.Info("Creating ArgoCD project...")
			if err := client.Create(ctx, desiredAppProject); err != nil {
				argoLogger.Error(err, "Error occurred when creating ArgoCD AppProject", "CR", argoCDCRName)
				return err
			}
			argoLogger.Info("Successfully created ArgoCD AppProject")
			return nil
		}
		argoLogger.Error(err, "Error occurred when retrieving ArgoCD AppProject", "CR", argoCDCRName)
		return err
	} else {
		// Compare the current and desired state
		if !reflect.DeepEqual(desiredAppProject.Spec, existingAppProject.Spec) {
			existingAppProject.Spec = desiredAppProject.Spec
			if err := client.Update(ctx, existingAppProject); err != nil {
				argoLogger.Error(err, "Error occurred when updating GitOps", "ArgoCD", argoCDCRName)
				return err
			}
		}
	}
	return nil
}

func handleArgoCDProjectCleanUp(gitOpsNamespace string, client client.Client, ctx context.Context) error {
	argoLogger := log.FromContext(ctx)

	argoLogger.Info("Handling ArgoCD ProjectCleanUp...")

	namespaceExist, _ := kube.CheckNamespaceExist(ctx, client, gitOpsNamespace)
	if namespaceExist {
		argoCDProjectCRList, err := listArgoCDProjectCR(ctx, client, gitOpsNamespace)

		if err != nil || len(argoCDProjectCRList) == 0 {
			argoLogger.Info("Failed to list or have no ArgoCD Project CRs created by Orchestrator Operator and cannot perform clean up process")
			return nil
		}

		if len(argoCDProjectCRList) == 1 {
			// remove ArgoCD Project CR
			err := client.Delete(ctx, &argoCDProjectCRList[0])
			if err != nil {
				argoLogger.Error(err, "Error occurred when deleting ArgoCD Project App", "ArgoCD", argoCDCRName)
				return err

			}
			argoLogger.Info("Successfully deleted ArgoCD Project CR created by orchestrator", "ArgoCD Project", argoCDCRName)
			return nil
		}
	}
	return nil
}

func listArgoCDProjectCR(ctx context.Context, k8client client.Client, namespace string) ([]argocdv1alpha1.AppProject, error) {
	argoLogger := log.FromContext(ctx)

	crList := &argocdv1alpha1.AppProjectList{}

	listOptions := []client.ListOption{
		client.InNamespace(namespace),
		client.MatchingLabels{kube.CreatedByLabelKey: kube.CreatedByLabelValue},
	}

	// List the CRs
	if err := k8client.List(ctx, crList, listOptions...); err != nil {
		argoLogger.Error(err, "Error occurred when listing ArgoCD Project CRs", "CR", argoCDCRName)
		return nil, err
	}

	argoLogger.Info("Successfully listed ArgoCD Project CRs", "Total", len(crList.Items))
	return crList.Items, nil
}
