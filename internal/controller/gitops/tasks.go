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
	"github.com/rhdhorchestrator/orchestrator-operator/internal/controller/kube"
	tektonv1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

const (
	tektonTaskAPIVersion = "tekton.dev/v1"
	tektonKind           = "Task"
	gitCLITask           = "git-cli"
	flattenerTask        = "flattener"
	buildManifestTask    = "build-manifests"
	buildGitOpsTask      = "build-gitops"
	tektonCRDName        = "tasks.tekton.dev"
)

var tektonTaskList = []string{
	gitCLITask,
	flattenerTask,
	buildManifestTask,
	buildGitOpsTask,
}

func HandleTektonTasks(client client.Client, ctx context.Context, gitOpsNamespace string) error {
	taskLogger := log.FromContext(ctx)
	taskLogger.Info("Handling Tekton Tasks...")

	if err := kube.CheckCRDExists(ctx, client, tektonCRDName); err != nil {
		taskLogger.Error(err, "Tekton Task CRD does not exist. Install RedHat Openshift Pipelines Operator")
		return err
	}

	for _, taskName := range tektonTaskList {
		if err := client.Get(ctx, types.NamespacedName{
			Namespace: gitOpsNamespace, Name: taskName}, &tektonv1.Task{}); err != nil {
			if apierrors.IsNotFound(err) {
				tektonTask := getTaskObject(gitOpsNamespace, taskName)
				if tektonTask != nil {
					if err := client.Create(ctx, tektonTask); err != nil {
						taskLogger.Error(err, "Error occurred when creating Tekton Task", "Task", taskName)
						return err
					}
					taskLogger.Info("Successfully created Tekton Task", "Task", taskName)
					continue
				}
			}
			taskLogger.Error(err, "Error occurred when checking task exist", "Task", taskName)
			continue
		}
	}
	return nil
}

func getTaskObject(gitOpsNamespace, taskName string) *tektonv1.Task {
	switch taskName {
	case gitCLITask:
		return createGitCLITaskObject(gitOpsNamespace)
	case flattenerTask:
		return createFlattenerTaskObject(gitOpsNamespace)
	case buildManifestTask:
		return createBuildManifestTaskObject(gitOpsNamespace)
	case buildGitOpsTask:
		return createBuildGitOpsTaskObject(gitOpsNamespace)
	default:
		return nil
	}
}

func createGitCLITaskObject(gitOpsNamespace string) *tektonv1.Task {
	return &tektonv1.Task{
		TypeMeta: metav1.TypeMeta{
			APIVersion: tektonTaskAPIVersion,
			Kind:       tektonKind,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      gitCLITask,
			Namespace: gitOpsNamespace,
			Labels:    kube.GetOrchestratorLabel(),
			Annotations: map[string]string{
				"tekton.dev/pipelines.minVersion": "0.21.0",
				"tekton.dev/categories":           "Git",
				"tekton.dev/tags":                 "git",
				"tekton.dev/displayName":          "git cli",
				"tekton.dev/platforms":            "linux/amd64,linux/s390x,linux/ppc64le",
			},
		},
		Spec: tektonv1.TaskSpec{
			Description: `This task can be used to perform git operations.Git command that needs to be run can be passed as a script to the task. This task needs authentication to git in order to push after the git operation.`,
			Workspaces: []tektonv1.WorkspaceDeclaration{
				{
					Name:        "source",
					Description: "A workspace that contains the fetched git repository.",
				},
				{
					Name:        "input",
					Optional:    true,
					Description: `An optional workspace that contains the files that need to be added to git. You can access the workspace from your script using $(workspaces.input.path).`,
				},
				{
					Name:        "ssh-directory",
					Optional:    true,
					Description: `A .ssh directory with private key, known_hosts, config, etc. Used to authenticate with the git remote.`,
				},
				{
					Name:        "basic-auth",
					Optional:    true,
					Description: `A Workspace containing a .gitconfig and .git-credentials file for authentication.`,
				},
			},
			Params: []tektonv1.ParamSpec{
				{
					Name:        "BASE_IMAGE",
					Description: "The base image for the task.",
					Type:        tektonv1.ParamTypeString,
					Default: &tektonv1.ParamValue{
						Type:      tektonv1.ParamTypeString,
						StringVal: "cgr.dev/chainguard/git:root-2.39@sha256:7759f87050dd8bacabe61354d75ccd7f864d6b6f8ec42697db7159eccd491139"},
				},
				{
					Name:        "GIT_USER_NAME",
					Type:        tektonv1.ParamTypeString,
					Description: "Git user name for performing git operation.",
					Default: &tektonv1.ParamValue{
						Type:      tektonv1.ParamTypeString,
						StringVal: "",
					},
				},
				{
					Name:        "GIT_USER_EMAIL",
					Type:        tektonv1.ParamTypeString,
					Description: "Git user email for performing git operation.",
					Default: &tektonv1.ParamValue{
						Type:      tektonv1.ParamTypeString,
						StringVal: "",
					},
				},
				{
					Name:        "GIT_SCRIPT",
					Type:        tektonv1.ParamTypeString,
					Description: "The git script to run.",
					Default: &tektonv1.ParamValue{
						Type:      tektonv1.ParamTypeString,
						StringVal: "git help",
					},
				},
				{
					Name:        "USER_HOME",
					Type:        tektonv1.ParamTypeString,
					Description: "Absolute path to the user's home directory.",
					Default: &tektonv1.ParamValue{
						Type:      tektonv1.ParamTypeString,
						StringVal: "/root",
					},
				},
				{
					Name:        "VERBOSE",
					Type:        tektonv1.ParamTypeString,
					Description: "Log the commands that are executed during `git-clone`'s operation.",
					Default: &tektonv1.ParamValue{
						Type:      tektonv1.ParamTypeString,
						StringVal: "true",
					},
				},
			},
			Results: []tektonv1.TaskResult{
				{
					Name:        "commit",
					Description: "The precise commit SHA after the git operation.",
				},
			},
			Steps: []tektonv1.Step{
				{
					Name:       "git",
					Image:      "$(params.BASE_IMAGE)",
					WorkingDir: "$(workspaces.source.path)",
					Env: []corev1.EnvVar{
						{Name: "HOME", Value: "$(params.USER_HOME)"},
						{Name: "PARAM_VERBOSE", Value: "$(params.VERBOSE)"},
						{Name: "PARAM_USER_HOME", Value: "$(params.USER_HOME)"},
						{Name: "WORKSPACE_OUTPUT_PATH", Value: "$(workspaces.output.path)"},
						{Name: "WORKSPACE_SSH_DIRECTORY_BOUND", Value: "$(workspaces.ssh-directory.bound)"},
						{Name: "WORKSPACE_SSH_DIRECTORY_PATH", Value: "$(workspaces.ssh-directory.path)"},
						{Name: "WORKSPACE_BASIC_AUTH_DIRECTORY_BOUND", Value: "$(workspaces.basic-auth.bound)"},
						{Name: "WORKSPACE_BASIC_AUTH_DIRECTORY_PATH", Value: "$(workspaces.basic-auth.path)"},
					},
					Script: gitCLITaskScript,
				},
			},
		},
	}
}

func createFlattenerTaskObject(gitOpsNamespace string) *tektonv1.Task {
	return &tektonv1.Task{
		TypeMeta: metav1.TypeMeta{
			APIVersion: tektonTaskAPIVersion,
			Kind:       tektonKind,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      flattenerTask,
			Namespace: gitOpsNamespace,
			Labels:    kube.GetOrchestratorLabel(),
		},
		Spec: tektonv1.TaskSpec{
			Workspaces: []tektonv1.WorkspaceDeclaration{
				{
					Name: "workflow-source",
				},
			},
			Params: []tektonv1.ParamSpec{
				{
					Name:        "workflowId",
					Description: "The workflow ID from the repository",
					Type:        tektonv1.ParamTypeString,
				},
				{
					Name:        "convertToFlat",
					Description: "Whether conversion to flat layout is needed or it's already flattened",
					Type:        tektonv1.ParamTypeString,
					Default: &tektonv1.ParamValue{
						Type:      tektonv1.ParamTypeString,
						StringVal: "true",
					},
				},
			},
			Steps: []tektonv1.Step{
				{
					Name:       "flatten",
					Image:      "registry.access.redhat.com/ubi9-minimal",
					WorkingDir: "$(workspaces.workflow-source.path)",
					Script:     flattenerTaskScript,
				},
			},
		},
	}
}

func createBuildManifestTaskObject(gitOpsNamespace string) *tektonv1.Task {
	return &tektonv1.Task{
		TypeMeta: metav1.TypeMeta{
			APIVersion: tektonTaskAPIVersion,
			Kind:       tektonKind,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      buildManifestTask,
			Namespace: gitOpsNamespace,
			Labels:    kube.GetOrchestratorLabel(),
		},
		Spec: tektonv1.TaskSpec{
			Workspaces: []tektonv1.WorkspaceDeclaration{
				{
					Name: "workflow-source",
				},
			},
			Params: []tektonv1.ParamSpec{
				{
					Name:        "workflowId",
					Description: "The workflow ID from the repository",
					Type:        tektonv1.ParamTypeString,
				},
			},
			Steps: []tektonv1.Step{
				{
					Name:       buildManifestTask,
					Image:      "registry.access.redhat.com/ubi9-minimal",
					WorkingDir: "$(workspaces.workflow-source.path)/flat/$(params.workflowId)",
					Script:     buildManifestTaskScript,
				},
			},
		},
	}
}

func createBuildGitOpsTaskObject(gitOpsNamespace string) *tektonv1.Task {
	return &tektonv1.Task{
		TypeMeta: metav1.TypeMeta{
			APIVersion: tektonTaskAPIVersion,
			Kind:       tektonKind,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      buildGitOpsTask,
			Namespace: gitOpsNamespace,
			Labels:    kube.GetOrchestratorLabel(),
		},
		Spec: tektonv1.TaskSpec{
			Workspaces: []tektonv1.WorkspaceDeclaration{
				{Name: "workflow-source"},
				{Name: "workflow-gitops"},
			},
			Params: []tektonv1.ParamSpec{
				{
					Name:        "workflowId",
					Description: "The workflow ID from the repository",
					Type:        tektonv1.ParamTypeString,
				},
				{
					Name: "imageTag",
					Type: tektonv1.ParamTypeString,
				},
			},
			Steps: []tektonv1.Step{
				{
					Name:       buildGitOpsTask,
					Image:      "registry.access.redhat.com/ubi9-minimal",
					WorkingDir: "$(workspaces.workflow-gitops.path)/workflow-gitops",
					Script:     buildGitOpsTaskScript,
				},
			},
		},
	}
}

func handleTektonTaskCleanUp(client client.Client, ctx context.Context, gitOpsNamespace string) error {
	taskLogger := log.FromContext(ctx)
	taskLogger.Info("Handling Tekton Tasks cleanup...")

	namespaceExist, _ := kube.CheckNamespaceExist(ctx, client, gitOpsNamespace)
	if namespaceExist {
		tektonTaskList, err := listTektonTaskCR(ctx, client, gitOpsNamespace)

		if err != nil || len(tektonTaskList) == 0 {
			taskLogger.Info("Failed to list or have no Tekton Task created by Orchestrator Operator and cannot perform clean up process")
			return nil
		}

		for _, tektonTask := range tektonTaskList {
			taskLogger.Info("Deleting Tekton Task ", "Task", tektonTask.Name)
			// remove tekton task
			err := client.Delete(ctx, &tektonTask)
			if err != nil {
				taskLogger.Error(err, "Error occurred when deleting tekton task", "Tekton Task", tektonTask.Name)
				continue
			}
			taskLogger.Info("Successfully deleted Tekton Task CR created by orchestrator", "Tekton Task", tektonTask.Name)
			continue
		}
	}

	return nil
}

func listTektonTaskCR(ctx context.Context, k8client client.Client, namespace string) ([]tektonv1.Task, error) {
	taskLogger := log.FromContext(ctx)

	crList := &tektonv1.TaskList{}

	listOptions := []client.ListOption{
		client.InNamespace(namespace),
		client.MatchingLabels{kube.CreatedByLabelKey: kube.CreatedByLabelValue},
	}

	// List the CRs
	if err := k8client.List(ctx, crList, listOptions...); err != nil {
		taskLogger.Error(err, "Error occurred when listing Tekton Task CRs")
		return nil, err
	}

	taskLogger.Info("Successfully listed Tekton Task CRs", "Total", len(crList.Items))
	return crList.Items, nil
}
