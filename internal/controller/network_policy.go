package controller

import (
	"context"
	"reflect"

	kubeoperations "github.com/rhdhorchestrator/orchestrator-operator/internal/controller/kube"
	knative "github.com/rhdhorchestrator/orchestrator-operator/internal/controller/knative"

	networkingv1 "k8s.io/api/networking/v1"
	apierrros "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

const (
	metaDataNameLabel                    = "kubernetes.io/metadata.name"
	monitoringNamespace                  = "openshift-user-workload-monitoring"
	allowRHDHToSonataflowWorkflows       = "allow-rhdh-to-sonataflow-and-workflows"
	allowIntraNamespace                  = "allow-intra-namespace"
	allowMonitoringToSonataflowWorkflows = "allow-monitoring-to-sonataflow-and-workflows"
)

var (
	NetworkPoliciesList = []string{
		allowRHDHToSonataflowWorkflows,
		allowIntraNamespace,
		allowMonitoringToSonataflowWorkflows,
	}
	allErrors = make(map[string]error)
)

// handleNetworkPolicy performs the retrieval, creation and reconciling of network policy.
// It returns an error if any occurs during retrieval, creation or reconciliation.
func handleNetworkPolicy(client client.Client, ctx context.Context,
	networkAndServerlessWorkflowNamespace, rhdhNamespace, databaseNamespace string, monitoringFlag bool) map[string]error {
	npLogger := log.FromContext(ctx)

	for _, NetworkPolicyName := range NetworkPoliciesList {

		if !monitoringFlag && (NetworkPolicyName == allowMonitoringToSonataflowWorkflows) {
			continue
		}

		// Define the desired NetworkPolicy
		desiredNP := &networkingv1.NetworkPolicy{
			ObjectMeta: metav1.ObjectMeta{
				Name:      NetworkPolicyName,
				Namespace: networkAndServerlessWorkflowNamespace,
				Labels:    kubeoperations.GetOrchestratorLabel(),
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
func createIngress(networkPolicyName string, networkAndServerlessWorkflowNamespace, rhdhNamespace, databaseNamespace string) []networkingv1.NetworkPolicyIngressRule {

	switch networkPolicyName {
	case allowRHDHToSonataflowWorkflows:
		return createIngressRHDHSonataflowWorkflows(networkAndServerlessWorkflowNamespace, rhdhNamespace, databaseNamespace)
	case allowIntraNamespace:
		return createIngressIntraNamespaces()
	case allowMonitoringToSonataflowWorkflows:
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
							metaDataNameLabel: knative.KnativeEventingNamespacedName,
						},
					},
				},
				{
					// Allows traffic from pods in the K-Native Serving namespace
					NamespaceSelector: &metav1.LabelSelector{
						MatchLabels: map[string]string{
							metaDataNameLabel: knative.KnativeServingNamespacedName,
						},
					},
				},
				{
					// Allows traffic from pods in the Workflow namespace
					NamespaceSelector: &metav1.LabelSelector{
						MatchLabels: map[string]string{
							metaDataNameLabel: networkAndServerlessWorkflowNamespace,
						},
					},
				},
				{
					// Allows traffic from pods in the RHDH namespace
					NamespaceSelector: &metav1.LabelSelector{
						MatchLabels: map[string]string{
							metaDataNameLabel: rhdhNamespace,
						},
					},
				},
				{
					// Allows traffic from pods in the Database namespace
					NamespaceSelector: &metav1.LabelSelector{
						MatchLabels: map[string]string{
							metaDataNameLabel: databaseNamespace,
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
							metaDataNameLabel: monitoringNamespace,
						},
					},
				},
			},
		},
	}
	return Ingress
}
