package gitops

import (
	"context"
	argocdv1alpha1 "github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	"github.com/parodos-dev/orchestrator-operator/internal/controller/kube"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

const (
	argoCDCRName     = "orchestrator-gitops"
	argoCDAPIVersion = "argoproj.io/v1alpha1"
	argoCDKind       = "AppProject"
)

func handleArgoCDProject(gitOpsNamespace string, client client.Client, ctx context.Context) {
	argoLogger := log.FromContext(ctx)
	argoLogger.Info("Handling ArgoCD Project...")

	appProject := &argocdv1alpha1.AppProject{
		TypeMeta: metav1.TypeMeta{
			APIVersion: argoCDAPIVersion,
			Kind:       argoCDKind,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      argoCDCRName,
			Namespace: gitOpsNamespace,
			Labels:    kube.AddLabel(),
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

	if err := client.Create(ctx, appProject); err != nil {
		argoLogger.Error(err, "Error occurred when creating ArgoCD AppProject", "CR", argoCDCRName)
	}

	argoLogger.Info("Successfully created ArgoCD AppProject")
}
