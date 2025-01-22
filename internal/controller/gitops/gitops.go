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
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// HandleGitOps performs the retrieval, creation and reconciling of Tekton and GitOps policy.
// It returns an error if any occurs during retrieval, creation or reconciliation.
func HandleGitOps(client client.Client, ctx context.Context, gitOpsNamespace string) error {
	logger := log.FromContext(ctx)
	logger.Info("Handling GitOps resource")

	if err := handleArgoCDProject(gitOpsNamespace, client, ctx); err != nil {
		return err
	}

	if err := handleTektonPipelineTasks(client, ctx, gitOpsNamespace); err != nil {
		return err
	}

	return nil
}

func handleTektonPipelineTasks(client client.Client, ctx context.Context, gitOpsNamespace string) error {
	logger := log.FromContext(ctx)
	logger.Info("Handling Tekton resource")

	// handle tekton task
	if err := HandleTektonTasks(client, ctx, gitOpsNamespace); err != nil {
		return err
	}

	// handle tekton pipeline
	if err := HandleTektonPipeline(client, ctx, gitOpsNamespace); err != nil {
		return err
	}
	return nil
}

func HandleGitOpsCleanUp(client client.Client, ctx context.Context, gitOpsNamespace string) error {
	logger := log.FromContext(ctx)
	logger.Info("Handling GitOps resource clean up")

	// handle argocd clean up
	if err := handleArgoCDProjectCleanUp(gitOpsNamespace, client, ctx); err != nil {
		return err
	}

	// handle tekton pipeline clean up
	if err := handleTektonPipelineCleanUp(client, ctx, gitOpsNamespace); err != nil {
		return err
	}

	// handle tekton clean up
	if err := handleTektonTaskCleanUp(client, ctx, gitOpsNamespace); err != nil {
		return err
	}
	return nil
}
