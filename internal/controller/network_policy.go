package controller

import (
	"context"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/errors"
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

func handleNetworkPolicy(client client.Client, ctx context.Context,
	networkAndServerlessWorkflowNamespace, rhdhNamespace, databaseNamespace string) error {
	npLogger := log.FromContext(ctx)
	// Define the desired NetworkPolicy
	desiredNP := &networkingv1.NetworkPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      NetworkPolicyName,
			Namespace: networkAndServerlessWorkflowNamespace,
		},
		Spec: networkingv1.NetworkPolicySpec{
			PodSelector: metav1.LabelSelector{},
			PolicyTypes: []networkingv1.PolicyType{
				networkingv1.PolicyTypeIngress,
			},
			Ingress: []networkingv1.NetworkPolicyIngressRule{
				{
					From: []networkingv1.NetworkPolicyPeer{
						{
							PodSelector: &metav1.LabelSelector{},
						},
						{
							NamespaceSelector: &metav1.LabelSelector{
								MatchLabels: map[string]string{
									MetaDataNameLabel: knativeEventingNamespacedName,
								},
							},
						},
						{
							NamespaceSelector: &metav1.LabelSelector{
								MatchLabels: map[string]string{
									MetaDataNameLabel: knativeServingNamespacedName,
								},
							},
						},
						{
							NamespaceSelector: &metav1.LabelSelector{
								MatchLabels: map[string]string{
									MetaDataNameLabel: networkAndServerlessWorkflowNamespace,
								},
							},
						},
						{
							NamespaceSelector: &metav1.LabelSelector{
								MatchLabels: map[string]string{
									MetaDataNameLabel: rhdhNamespace,
								},
							},
						},
						{
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
		if errors.IsNotFound(err) {
			// create networkPolicy
			if err := client.Create(ctx, desiredNP); err != nil {
				npLogger.Error(err, "Error occurred when creating NetworkPolicy", "NP", NetworkPolicyName)
				return err
			}
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
