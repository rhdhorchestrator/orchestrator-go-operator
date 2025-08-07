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

var objects []client.Object

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
						Labels:    kubeoperations.GetOrchestratorLabel(),
					},
					Spec: networkingv1.NetworkPolicySpec{
						PodSelector: metav1.LabelSelector{},
						PolicyTypes: []networkingv1.PolicyType{
							networkingv1.PolicyTypeIngress,
						},
						Ingress: []networkingv1.NetworkPolicyIngressRule{},
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      allowServerlessLogicToSonataFlowWorkflows,
						Namespace: testNamespace,
						Labels:    kubeoperations.GetOrchestratorLabel(),
					},
					Spec: networkingv1.NetworkPolicySpec{
						PodSelector: metav1.LabelSelector{},
						PolicyTypes: []networkingv1.PolicyType{
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

			existingNP := &networkingv1.NetworkPolicy{}
			// Flow for test cases that expect creating Policies
			if tc.expectCreate {

				// Create fake client with no existing Network Policies
				fakeClient := fake.NewClientBuilder().WithScheme(scheme).Build()

				// Call handler to Create the Network Policies
				errors := handleNetworkPolicy(fakeClient, ctx, testNamespace, testRHDHNamespace, testDatabaseNamespace, tc.monitoringFlag)

				// Verify that the fake client is populated with policies after calling the handler
				err := fakeClient.Get(ctx, types.NamespacedName{Name: allowRHDHToSonataflowWorkflows, Namespace: testNamespace}, existingNP)
				assert.NoError(t, err)
				err = fakeClient.Get(ctx, types.NamespacedName{Name: allowIntraNamespace, Namespace: testNamespace}, existingNP)
				assert.NoError(t, err)
				err = fakeClient.Get(ctx, types.NamespacedName{Name: allowServerlessLogicToSonataFlowWorkflows, Namespace: testNamespace}, existingNP)
				assert.NoError(t, err)

				if tc.monitoringFlag {
					err = fakeClient.Get(ctx, types.NamespacedName{Name: allowMonitoringToSonataflowWorkflows, Namespace: testNamespace}, existingNP)
					assert.NoError(t, err)
				}
				assert.Equal(t, tc.errorMap, errors)

				// Flow for test cases that expect updating existing Policies
			} else if tc.expectUpdate {

				// This conversion will allow passing multiple Network Policy objects to the fake client
				for _, policy := range tc.existingPolicies {
					objects = append(objects, policy)
				}

				// Create the fake client with existing Network Policies
				fakeClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(objects...).Build()

				// Verify that the fake client is populated with a policy
				err := fakeClient.Get(ctx, types.NamespacedName{Name: allowRHDHToSonataflowWorkflows, Namespace: testNamespace}, existingNP)
				assert.NoError(t, err)

				// Call handler to update the Ingress
				errors := handleNetworkPolicy(fakeClient, ctx, testNamespace, testRHDHNamespace, testDatabaseNamespace, tc.monitoringFlag)
				assert.Equal(t, tc.errorMap, errors)
				err = fakeClient.Get(ctx, types.NamespacedName{Name: allowRHDHToSonataflowWorkflows, Namespace: testNamespace}, existingNP)
				assert.NoError(t, err)
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
			name:            "Create Serverless Operator Ingress",
			npName:          allowServerlessLogicToSonataFlowWorkflows,
			expectedIngress: createIngressServerlessLogicSonataFlowWorkflows(),
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
