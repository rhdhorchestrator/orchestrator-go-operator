package controller

import (
	"context"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func handleNetworkPolicy(client client.Client, ctx context.Context, networkNamespace string) error {
	// Define the desired NetworkPolicy
	np := &networkingv1.NetworkPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "allow-intra-namespace",
			Namespace: networkNamespace,
		},
		Spec: networkingv1.NetworkPolicySpec{
			PodSelector: metav1.LabelSelector{
				MatchLabels: map[string]string{"kubernetes.io/metadata.name": "example"},
			},
			PolicyTypes: []networkingv1.PolicyType{
				networkingv1.PolicyTypeIngress,
				networkingv1.PolicyTypeEgress,
			},
			Ingress: []networkingv1.NetworkPolicyIngressRule{
				{
					From: []networkingv1.NetworkPolicyPeer{
						{
							PodSelector: &metav1.LabelSelector{
								MatchLabels: map[string]string{"role": "frontend"},
							},
						},
					},
				},
			},
		},
	}

	// Create or Update the NetworkPolicy
	existingNP := &networkingv1.NetworkPolicy{}
	err := client.Get(ctx, types.NamespacedName{Name: np.Name, Namespace: np.Namespace}, existingNP)
	if err != nil {
		if errors.IsNotFound(err) {
			// If the NetworkPolicy doesn't exist, create it
			if err := client.Create(ctx, np); err != nil {

			}
		}
		return err
	}

	// If the NetworkPolicy exists, update it if necessary
	if !reflect.DeepEqual(np.Spec, existingNP.Spec) {
		existingNP.Spec = np.Spec
		return client.Update(ctx, existingNP)
	}
	return nil
}
