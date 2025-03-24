package controller

import (
	"context"
	"testing"

	kubeoperations "github.com/rhdhorchestrator/orchestrator-operator/internal/controller/kube"
	"github.com/stretchr/testify/assert"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

const (
	testNamespace         = "test-namespace"
	testRHDHNamespace     = "test-rhdh-namespace"
	testDatabaseNamespace = "test-db-namespace"
)

func TestHandleNetworkPolicy(t *testing.T) {

	ctx := context.TODO()
	// Create a fake client scheme
	scheme := runtime.NewScheme()
	utilruntime.Must(networkingv1.AddToScheme(scheme))
	utilruntime.Must(metav1.AddMetaToScheme(scheme))
	testCases := []struct {
		name             string
		existingPolicies []*networkingv1.NetworkPolicy
		monitoringFlag   bool
		expectCreate     bool
		expectUpdate     bool
		errorMap         map[string]error
	}{
		{
			name:             "Creates new policies when they don't exist",
			existingPolicies: []*networkingv1.NetworkPolicy{},
			monitoringFlag:   false,
			expectCreate:     true,
			expectUpdate:     false,
			errorMap:         map[string]error{},
		},
		{
			name:             "Creates new policies when they don't exist, with monitoring",
			existingPolicies: []*networkingv1.NetworkPolicy{},
			monitoringFlag:   true,
			expectCreate:     true,
			expectUpdate:     false,
			errorMap:         map[string]error{},
		},
		{
			name: "Updates existing policies",
			existingPolicies: []*networkingv1.NetworkPolicy{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      allowRHDHToSonataflowWorkflows,
						Namespace: testNamespace,
						Labels:    kubeoperations.AddLabel(),
					},
					Spec: networkingv1.NetworkPolicySpec{
						// This policy applies to all pods within the namespace where the policy is defined
						PodSelector: metav1.LabelSelector{},
						PolicyTypes: []networkingv1.PolicyType{
							// This policy concerns traffic coming into the pods
							networkingv1.PolicyTypeIngress,
						},
						Ingress: []networkingv1.NetworkPolicyIngressRule{},
					},
				},
			},
			monitoringFlag: false,
			expectCreate:   false,
			expectUpdate:   true,
			errorMap:       map[string]error{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			var objects []client.Object
			for _, policy := range tc.existingPolicies {
				objects = append(objects, policy)
			}

			fakeClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(objects...).Build()

			existingNP := &networkingv1.NetworkPolicy{}
			if tc.expectCreate && tc.expectUpdate {

			} else if tc.expectCreate {
				err := fakeClient.Get(ctx, types.NamespacedName{Name: allowRHDHToSonataflowWorkflows, Namespace: testNamespace}, existingNP)
				assert.Error(t, err)
				err = fakeClient.Get(ctx, types.NamespacedName{Name: allowIntraNamespace, Namespace: testNamespace}, existingNP)
				assert.Error(t, err)

				if tc.monitoringFlag {
					err = fakeClient.Get(ctx, types.NamespacedName{Name: allowMonitoringToSonataflowWorkflows, Namespace: testNamespace}, existingNP)
					assert.Error(t, err)
				}

				errors := handleNetworkPolicy(fakeClient, ctx, testNamespace, testRHDHNamespace, testDatabaseNamespace, tc.monitoringFlag)

				err = fakeClient.Get(ctx, types.NamespacedName{Name: allowRHDHToSonataflowWorkflows, Namespace: testNamespace}, existingNP)
				assert.Nil(t, err)
				err = fakeClient.Get(ctx, types.NamespacedName{Name: allowIntraNamespace, Namespace: testNamespace}, existingNP)
				assert.Nil(t, err)

				if tc.monitoringFlag {
					err = fakeClient.Get(ctx, types.NamespacedName{Name: allowMonitoringToSonataflowWorkflows, Namespace: testNamespace}, existingNP)
					assert.Nil(t, err)
				}
				assert.Equal(t, tc.errorMap, errors)

			} else if tc.expectUpdate {

				err := fakeClient.Get(ctx, types.NamespacedName{Name: allowRHDHToSonataflowWorkflows, Namespace: testNamespace}, existingNP)
				assert.Nil(t, err)

				// Call handler to update the Ingress
				errors := handleNetworkPolicy(fakeClient, ctx, testNamespace, testRHDHNamespace, testDatabaseNamespace, tc.monitoringFlag)
				assert.Equal(t, tc.errorMap, errors)
				err = fakeClient.Get(ctx, types.NamespacedName{Name: allowRHDHToSonataflowWorkflows, Namespace: testNamespace}, existingNP)
				assert.Nil(t, err)
				assert.NotEqual(t, tc.existingPolicies[len(tc.existingPolicies)-1].Spec.Ingress, existingNP)

			}

		})
	}
}

func TestCreateIngressSwitch(t *testing.T) {
	// Create a fake client scheme
	scheme := runtime.NewScheme()
	utilruntime.Must(networkingv1.AddToScheme(scheme))
	utilruntime.Must(metav1.AddMetaToScheme(scheme))
	testCases := []struct {
		name            string
		npName          string
		expectedIngress []networkingv1.NetworkPolicyIngressRule
	}{
		{
			name:            "Create RHDH Ingress",
			npName:          allowRHDHToSonataflowWorkflows,
			expectedIngress: createIngressRHDHSonataflowWorkflows(testNamespace, testRHDHNamespace, testDatabaseNamespace),
		},
		{
			name:            "Create Intra Ingress",
			npName:          allowIntraNamespace,
			expectedIngress: createIngressIntraNamespaces(),
		},
		{
			name:            "Create Monitoring Ingress",
			npName:          allowMonitoringToSonataflowWorkflows,
			expectedIngress: createIngressMonitoringSonataflowWorkflows(),
		},
		{
			name:            "Create Non existent Ingress",
			npName:          "test",
			expectedIngress: []networkingv1.NetworkPolicyIngressRule{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ingress := createIngress(tc.npName, testNamespace, testRHDHNamespace, testDatabaseNamespace)
			assert.Equal(t, tc.expectedIngress, ingress)
		})
	}

}
