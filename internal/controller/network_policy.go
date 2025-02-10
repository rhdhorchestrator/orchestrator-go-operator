package controller

import (
	"context"
	kubeoperations "github.com/rhdhorchestrator/orchestrator-operator/internal/controller/kube"
	networkingv1 "k8s.io/api/networking/v1"
	apierrros "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

const (
	MetaDataNameLabel = "kubernetes.io/metadata.name"
	NetworkPolicyName = "allow-rhdh-to-sonataflow-and-workflows"
)

// handleNetworkPolicy performs the retrieval, creation and reconciling of network policy.
// It returns an error if any occurs during retrieval, creation or reconciliation.
func handleNetworkPolicy(client client.Client, ctx context.Context,
	networkAndServerlessWorkflowNamespace, rhdhNamespace, databaseNamespace string) error {
	npLogger := log.FromContext(ctx)
	// Define the desired NetworkPolicy
	desiredNP := &networkingv1.NetworkPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      NetworkPolicyName,
			Namespace: networkAndServerlessWorkflowNamespace,
			Labels:    kubeoperations.AddLabel(),
		},
		Spec: networkingv1.NetworkPolicySpec{
			// This policy applies to all pods within the namespace where the policy is defined
			PodSelector: metav1.LabelSelector{},
			PolicyTypes: []networkingv1.PolicyType{
				// This policy concerns traffic coming into the pods
				networkingv1.PolicyTypeIngress,
			},
			Ingress: []networkingv1.NetworkPolicyIngressRule{
				{
					From: []networkingv1.NetworkPolicyPeer{
						{
							// Allows traffic from all pods within the same namespace as the defined network policy
							PodSelector: &metav1.LabelSelector{},
						},
						{
							// Allows traffic from pods in the K-Native Eventing namespace
							NamespaceSelector: &metav1.LabelSelector{
								MatchLabels: map[string]string{
									MetaDataNameLabel: knativeEventingNamespacedName,
								},
							},
						},
						{
							// Allows traffic from pods in the K-Native Serving namespace
							NamespaceSelector: &metav1.LabelSelector{
								MatchLabels: map[string]string{
									MetaDataNameLabel: knativeServingNamespacedName,
								},
							},
						},
						{
							// Allows traffic from pods in the Workflow namespace
							NamespaceSelector: &metav1.LabelSelector{
								MatchLabels: map[string]string{
									MetaDataNameLabel: networkAndServerlessWorkflowNamespace,
								},
							},
						},
						{
							// Allows traffic from pods in the RHDH namespace
							NamespaceSelector: &metav1.LabelSelector{
								MatchLabels: map[string]string{
									MetaDataNameLabel: rhdhNamespace,
								},
							},
						},
						{
							// Allows traffic from pods in the Database namespace
							NamespaceSelector: &metav1.LabelSelector{
								MatchLabels: map[string]string{
									MetaDataNameLabel: databaseNamespace,
								},
							},
						},
					},
				},
			},
		},
	}

	existingNP := &networkingv1.NetworkPolicy{}
	// get existing the networkPolicy
	err := client.Get(ctx, types.NamespacedName{Name: desiredNP.Name, Namespace: desiredNP.Namespace}, existingNP)
	if err != nil {
		if apierrros.IsNotFound(err) {
			// create network policy
			if err := client.Create(ctx, desiredNP); err != nil {
				npLogger.Error(err, "Error occurred when creating NetworkPolicy", "NP", NetworkPolicyName)
				return err
			}
			return nil
		}
		return err
	}

	// Compare the current and desired state
	if !reflect.DeepEqual(desiredNP.Spec, existingNP.Spec) {
		existingNP.Spec = desiredNP.Spec
		if err := client.Update(ctx, existingNP); err != nil {
			npLogger.Error(err, "Error occurred when updating NetworkPolicy", "NP", NetworkPolicyName)
			return err
		}
	}
	return nil
}
