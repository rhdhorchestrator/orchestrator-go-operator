package controller

import (
	"context"
	kubeoperations "github.com/rhdhorchestrator/orchestrator-operator/internal/controller/kube"
	"reflect"

	networkingv1 "k8s.io/api/networking/v1"
	apierrros "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

const (
	MetaDataNameLabel   = "kubernetes.io/metadata.name"
	monitoringNamespace = "openshift-user-workload-monitoring"
)

var (
	NetworkPoliciesList = []string{
		"allow-rhdh-to-sonataflow-and-workflows",
		"allow-intra-namespace",
	}
	allErrors = make(map[string]error)
)

// handleNetworkPolicy performs the retrieval, creation and reconciling of network policy.
// It returns an error if any occurs during retrieval, creation or reconciliation.
func handleNetworkPolicy(client client.Client, ctx context.Context,
	networkAndServerlessWorkflowNamespace, rhdhNamespace, databaseNamespace string, monitoringFlag bool) map[string]error {
	npLogger := log.FromContext(ctx)

	// If monitoring is enabled, an additional NP needs to be applied
	if monitoringFlag {
		NetworkPoliciesList = append(NetworkPoliciesList, "allow-monitoring-to-sonataflow-and-workflows")
	}

	for _, NetworkPolicyName := range NetworkPoliciesList {

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
				Ingress: createIngress(NetworkPolicyName, networkAndServerlessWorkflowNamespace, rhdhNamespace, databaseNamespace),
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
					allErrors[NetworkPolicyName] = err
				}
			} else {
				// Pass along only actual errors
				allErrors[NetworkPolicyName] = err
			}

			continue
		}

		// Compare the current and desired state
		if !reflect.DeepEqual(desiredNP.Spec, existingNP.Spec) {
			existingNP.Spec = desiredNP.Spec
			if err := client.Update(ctx, existingNP); err != nil {
				npLogger.Error(err, "Error occurred when updating NetworkPolicy", "NP", NetworkPolicyName)
				allErrors[NetworkPolicyName] = err
			}
		}
	}

	return allErrors
}

// A switch to create an Ingress for each network policy.
func createIngress(NetworkPolicyName string, networkAndServerlessWorkflowNamespace, rhdhNamespace, databaseNamespace string) []networkingv1.NetworkPolicyIngressRule {

	switch NetworkPolicyName {
	case "allow-rhdh-to-sonataflow-and-workflows":
		return createIngressRHDHSonataflowWorkflows(networkAndServerlessWorkflowNamespace, rhdhNamespace, databaseNamespace)
	case "allow-intra-namespace":
		return createIngressIntraNamespaces()
	case "allow-monitoring-to-sonataflow-and-workflows":
		return createIngressMonitoringSonataflowWorkflows()
	default:
		return []networkingv1.NetworkPolicyIngressRule{}
	}
}
func createIngressRHDHSonataflowWorkflows(networkAndServerlessWorkflowNamespace, rhdhNamespace, databaseNamespace string) []networkingv1.NetworkPolicyIngressRule {
	Ingress := []networkingv1.NetworkPolicyIngressRule{
		{
			From: []networkingv1.NetworkPolicyPeer{
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
	}
	return Ingress
}

func createIngressIntraNamespaces() []networkingv1.NetworkPolicyIngressRule {
	Ingress := []networkingv1.NetworkPolicyIngressRule{
		{
			From: []networkingv1.NetworkPolicyPeer{
				{
					// Allows traffic from all pods within the same namespace as the defined network policy
					PodSelector: &metav1.LabelSelector{},
				},
			},
		},
	}
	return Ingress
}

func createIngressMonitoringSonataflowWorkflows() []networkingv1.NetworkPolicyIngressRule {
	Ingress := []networkingv1.NetworkPolicyIngressRule{
		{
			From: []networkingv1.NetworkPolicyPeer{
				{
					// Allow traffic for all pods in the openshift-user-workload-monitoring namespace.
					NamespaceSelector: &metav1.LabelSelector{
						MatchLabels: map[string]string{
							MetaDataNameLabel: monitoringNamespace,
						},
					},
				},
			},
		},
	}
	return Ingress
}
