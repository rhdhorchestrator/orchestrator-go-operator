package gitops

import (
	"context"
	"github.com/rhdhorchestrator/orchestrator-operator/internal/controller/kube"
	tektonv1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

const (
	tektonAPIVersion                = "tekton.dev/v1"
	pipelineName                    = "workflow-deployment"
	fetchWorkflowPipelineTask       = "fetch-workflow"
	fetchWorkflowGitOpsPipelineTask = "fetch-workflow-gitops"
	flattenWorkflowPipelineTask     = "flatten-workflow"
	buildManifestsPipelineTask      = "build-manifests"
	buildGitOpsPipelineTask         = "build-gitops"
	buildAndPushImagePipelineTask   = "build-and-push-image"
	pushWorkflowGitOpsPipelineTask  = "push-workflow-gitops"
	pipelineCRDName                 = "pipelines.tekton.dev"
)

func HandleTektonPipeline(client client.Client, ctx context.Context, gitOpsNamespace string) error {
	logger := log.FromContext(ctx)
	logger.Info("Handling tekton pipeline resources")

	if err := kube.CheckCRDExists(ctx, client, pipelineCRDName); err != nil {
		logger.Error(err, "Tekton Pipeline CRD does not exist. Install RedHat Openshift Pipelines Operator")
		return err
	}

	// pipeline definition
	desiredPipeline := &tektonv1.Pipeline{
		TypeMeta: metav1.TypeMeta{
			APIVersion: tektonAPIVersion,
			Kind:       "Pipeline",
		},
		ObjectMeta: ctrl.ObjectMeta{
			Name:      pipelineName,
			Namespace: gitOpsNamespace,
			Labels:    kube.AddLabel(),
		},
		Spec: tektonv1.PipelineSpec{
			Description: "This pipeline clones a git repo, builds a Docker image with Kaniko, and pushes it to a registry",
			Params: []tektonv1.ParamSpec{
				{
					Name:        "gitUrl",
					Description: "The SSH URL of the repository to clone",
					Type:        tektonv1.ParamTypeString,
				},
				{
					Name:        "gitOpsUrl",
					Description: "The SSH URL of the config repository for pushing the changes",
					Type:        tektonv1.ParamTypeString,
				},
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
				{
					Name:        "quayOrgName",
					Description: "The Quay Organization Name of the published workflow",
					Type:        tektonv1.ParamTypeString,
				},
				{
					Name:        "quayRepoName",
					Description: "The Quay Repository Name of the published workflow",
					Type:        tektonv1.ParamTypeString,
				},
			},
			Workspaces: []tektonv1.PipelineWorkspaceDeclaration{
				{Name: "workflow-source"},
				{Name: "workflow-gitops"},
				{Name: "ssh-creds"},
				{Name: "docker-credentials"},
			},
			Tasks: []tektonv1.PipelineTask{
				{
					Name:    fetchWorkflowPipelineTask,
					TaskRef: &tektonv1.TaskRef{Name: gitCLITask},
					Workspaces: []tektonv1.WorkspacePipelineTaskBinding{
						{Name: "source", Workspace: "workflow-source"},
						{Name: "ssh-directory", Workspace: "ssh-creds"},
					},
					Params: []tektonv1.Param{
						{Name: "GIT_USER_NAME", Value: tektonv1.ParamValue{Type: tektonv1.ParamTypeString, StringVal: "The Orchestrator Tekton Pipeline"}},
						{Name: "GIT_USER_EMAIL", Value: tektonv1.ParamValue{Type: tektonv1.ParamTypeString, StringVal: "rhdhorchestrator@redhat.com"}},
						{Name: "USER_HOME", Value: tektonv1.ParamValue{Type: tektonv1.ParamTypeString, StringVal: "/home/git"}},
						{Name: "GIT_SCRIPT", Value: tektonv1.ParamValue{Type: tektonv1.ParamTypeString, StringVal: gitCloneScript}},
					},
				},
				{
					Name:    fetchWorkflowGitOpsPipelineTask,
					TaskRef: &tektonv1.TaskRef{Name: gitCLITask},
					Workspaces: []tektonv1.WorkspacePipelineTaskBinding{
						{Name: "source", Workspace: "workflow-gitops"},
						{Name: "ssh-directory", Workspace: "ssh-creds"},
					},
					Params: []tektonv1.Param{
						{Name: "GIT_USER_NAME", Value: tektonv1.ParamValue{Type: tektonv1.ParamTypeString, StringVal: "The Orchestrator Tekton Pipeline"}},
						{Name: "GIT_USER_EMAIL", Value: tektonv1.ParamValue{Type: tektonv1.ParamTypeString, StringVal: "rhdhorchestrator@redhat.com"}},
						{Name: "USER_HOME", Value: tektonv1.ParamValue{Type: tektonv1.ParamTypeString, StringVal: "/home/git"}},
						{Name: "GIT_SCRIPT", Value: tektonv1.ParamValue{Type: tektonv1.ParamTypeString, StringVal: gitCloneGitOpsScript}},
					},
				},
				{
					Name:     flattenWorkflowPipelineTask,
					RunAfter: []string{fetchWorkflowPipelineTask},
					TaskRef:  &tektonv1.TaskRef{Name: flattenerTask},
					Workspaces: []tektonv1.WorkspacePipelineTaskBinding{
						{Name: "workflow-source", Workspace: "workflow-source"}},
					Params: []tektonv1.Param{
						{Name: "workflowId", Value: tektonv1.ParamValue{Type: tektonv1.ParamTypeString, StringVal: "$(params.workflowId)"}},
						{Name: "convertToFlat", Value: tektonv1.ParamValue{Type: tektonv1.ParamTypeString, StringVal: "$(params.convertToFlat)"}},
					},
				},
				{
					Name:     buildManifestsPipelineTask,
					RunAfter: []string{flattenWorkflowPipelineTask},
					TaskRef:  &tektonv1.TaskRef{Name: buildManifestTask},
					Workspaces: []tektonv1.WorkspacePipelineTaskBinding{
						{Name: "workflow-source", Workspace: "workflow-source"}},
					Params: []tektonv1.Param{
						{Name: "workflowId", Value: tektonv1.ParamValue{Type: tektonv1.ParamTypeString, StringVal: "$(params.workflowId)"}}},
				},
				{
					Name:     buildGitOpsPipelineTask,
					RunAfter: []string{buildManifestsPipelineTask, fetchWorkflowGitOpsPipelineTask},
					TaskRef:  &tektonv1.TaskRef{Name: buildGitOpsTask},
					Workspaces: []tektonv1.WorkspacePipelineTaskBinding{
						{Name: "workflow-source", Workspace: "workflow-source"},
						{Name: "workflow-gitops", Workspace: "workflow-gitops"},
					},
					Params: []tektonv1.Param{
						{Name: "workflowId", Value: tektonv1.ParamValue{Type: tektonv1.ParamTypeString, StringVal: "$(params.workflowId)"}},
						{Name: "imageTag", Value: tektonv1.ParamValue{Type: tektonv1.ParamTypeString, StringVal: "$(tasks.fetch-workflow.results.commit)"}},
					},
				},
				{
					Name:     buildAndPushImagePipelineTask,
					RunAfter: []string{flattenWorkflowPipelineTask},
					TaskRef: &tektonv1.TaskRef{
						ResolverRef: tektonv1.ResolverRef{
							Resolver: "cluster",
							Params: []tektonv1.Param{
								{Name: "kind", Value: tektonv1.ParamValue{Type: tektonv1.ParamTypeString, StringVal: "task"}},
								{Name: "name", Value: tektonv1.ParamValue{Type: tektonv1.ParamTypeString, StringVal: "buildah"}},
								{Name: "namespace", Value: tektonv1.ParamValue{Type: tektonv1.ParamTypeString, StringVal: "openshift-pipelines"}},
							},
						},
					},
					Workspaces: []tektonv1.WorkspacePipelineTaskBinding{
						{Name: "source", Workspace: "workflow-source"},
						{Name: "dockerconfig", Workspace: "docker-credentials"},
					},
					Params: []tektonv1.Param{
						{Name: "IMAGE", Value: tektonv1.ParamValue{
							Type:      tektonv1.ParamTypeString,
							StringVal: "quay.io/$(params.quayOrgName)/$(params.quayRepoName):$(tasks.fetch-workflow.results.commit)",
						}},
						{Name: "DOCKERFILE", Value: tektonv1.ParamValue{Type: tektonv1.ParamTypeString, StringVal: "flat/workflow-builder.Dockerfile"}},
						{Name: "CONTEXT", Value: tektonv1.ParamValue{Type: tektonv1.ParamTypeString, StringVal: "flat/$(params.workflowId)"}},
						{Name: "BUILD_EXTRA_ARGS", Value: tektonv1.ParamValue{
							Type:      tektonv1.ParamTypeString,
							StringVal: "--authfile=/workspace/dockerconfig/.dockerconfigjson --ulimit nofile=4096:4096 --build-arg WF_RESOURCES=.",
						}},
					},
				},
				{
					Name:     pushWorkflowGitOpsPipelineTask,
					RunAfter: []string{buildGitOpsPipelineTask, buildAndPushImagePipelineTask},
					TaskRef: &tektonv1.TaskRef{
						Name: gitCLITask,
					},
					Workspaces: []tektonv1.WorkspacePipelineTaskBinding{
						{Name: "source", Workspace: "workflow-gitops"},
						{Name: "ssh-directory", Workspace: "ssh-creds"},
					},
					Params: []tektonv1.Param{
						{Name: "GIT_USER_NAME", Value: tektonv1.ParamValue{
							Type:      tektonv1.ParamTypeString,
							StringVal: "The Orchestrator Tekton Pipeline",
						}},
						{Name: "GIT_USER_EMAIL", Value: tektonv1.ParamValue{
							Type:      tektonv1.ParamTypeString,
							StringVal: "rhdhorchestrator@redhat.com",
						}},
						{Name: "USER_HOME", Value: tektonv1.ParamValue{
							Type:      tektonv1.ParamTypeString,
							StringVal: "/home/git",
						}},
						{Name: "GIT_SCRIPT", Value: tektonv1.ParamValue{
							Type:      tektonv1.ParamTypeString,
							StringVal: gitScript,
						}},
					},
				},
			},
		},
	}
	existingPipeline := &tektonv1.Pipeline{}

	if err := client.Get(ctx, types.NamespacedName{
		Namespace: gitOpsNamespace,
		Name:      pipelineName,
	}, existingPipeline); err != nil {
		if errors.IsNotFound(err) {
			if err := client.Create(ctx, desiredPipeline); err != nil {
				logger.Error(err, "Error occurred when creating Tekton Pipeline", "Pipeline", pipelineName)
				return err
			}
			logger.Info("Successfully created Tekton Pipeline", "Pipeline", pipelineName)
			return err
		}
		return err
	}
	return nil
}

func handleTektonPipelineCleanUp(client client.Client, ctx context.Context, gitOpsNamespace string) error {
	pipelineLogger := log.FromContext(ctx)

	pipelineLogger.Info("Handling Tekton Pipeline CleanUp...")

	namespaceExist, _ := kube.CheckNamespaceExist(ctx, client, gitOpsNamespace)
	if namespaceExist {
		tektonPipelineCRList, err := listTektonPipelineCR(ctx, client, gitOpsNamespace)

		if err != nil || len(tektonPipelineCRList) == 0 {
			pipelineLogger.Info("Failed to list or have no Tekton Pipeline CRs created by Orchestrator Operator and cannot perform clean up process")
			return nil
		}

		if len(tektonPipelineCRList) == 1 {
			// remove Tekton Pipeline CR
			err := client.Delete(ctx, &tektonPipelineCRList[0])
			if err != nil {
				pipelineLogger.Error(err, "Error occurred when deleting Tekton Pipeline", "Tekton Pipeline", pipelineName)
				return err

			}
			pipelineLogger.Info("Successfully deleted Tekton Pipeline CR created by orchestrator", "Tekton Pipeline", pipelineName)
			return nil
		}
	}
	return nil
}

func listTektonPipelineCR(ctx context.Context, k8client client.Client, namespace string) ([]tektonv1.Pipeline, error) {
	pipelineLogger := log.FromContext(ctx)

	crList := &tektonv1.PipelineList{}

	listOptions := []client.ListOption{
		client.InNamespace(namespace),
		client.MatchingLabels{kube.CreatedByLabelKey: kube.CreatedByLabelValue},
	}

	// List the CRs
	if err := k8client.List(ctx, crList, listOptions...); err != nil {
		pipelineLogger.Error(err, "Error occurred when listing Tekton Pipeline CRs")
		return nil, err
	}

	pipelineLogger.Info("Successfully listed Tekton Pipeline CRs", "Total", len(crList.Items))
	return crList.Items, nil
}
