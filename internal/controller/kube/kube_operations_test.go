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

package kube

import (
	"context"
	operatorsv1 "github.com/operator-framework/api/pkg/operators/v1"
	"github.com/operator-framework/api/pkg/operators/v1alpha1"
	olmclientsetfake "github.com/operator-framework/operator-lifecycle-manager/pkg/api/client/clientset/versioned/fake"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"testing"
)

var (
	orchestratorNamespace     = "orchestrator-namespace"
	orchestratorOperatorGroup = "orchestrator-operator-group"
	subscriptionName          = "orchestrator-subscription"
	subscription              = &v1alpha1.Subscription{
		ObjectMeta: metav1.ObjectMeta{
			Name:      subscriptionName,
			Namespace: orchestratorNamespace,
			Labels:    AddLabel(),
		},
		Spec: &v1alpha1.SubscriptionSpec{
			Channel:                "channel",
			StartingCSV:            "starting-csv",
			InstallPlanApproval:    v1alpha1.ApprovalManual,
			CatalogSource:          CatalogSourceName,
			CatalogSourceNamespace: CatalogSourceNamespace,
			Package:                subscriptionName,
		},
	}
	existingLabelMap = map[string]string{
		CreatedByLabelKey: CreatedByLabelValue,
	}
)

func TestCheckNamespaceExist(t *testing.T) {
	ctx := context.TODO()
	// Create a fake client scheme
	scheme := runtime.NewScheme()
	utilruntime.Must(corev1.AddToScheme(scheme))

	testCases := []struct {
		name           string
		namespace      string
		namespaceObj   *corev1.Namespace
		expectedExists bool
	}{
		{
			name:           "Namespace exists",
			namespace:      orchestratorNamespace,
			namespaceObj:   &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: orchestratorNamespace}},
			expectedExists: true,
		},
		{
			name:           "Namespace does not exist",
			namespace:      "fake-namespace",
			namespaceObj:   &corev1.Namespace{},
			expectedExists: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			fakeClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(tc.namespaceObj).Build()
			exists, _ := CheckNamespaceExist(ctx, fakeClient, tc.namespace)
			assert.Equal(t, tc.expectedExists, exists)
		})
	}
}

func TestCreateNamespace(t *testing.T) {
	ctx := context.TODO()
	scheme := runtime.NewScheme()
	utilruntime.Must(corev1.AddToScheme(scheme))

	testCases := []struct {
		name          string
		namespace     string
		namespaceObj  *corev1.Namespace
		expectedError error
	}{
		{
			name:          "Create namespace",
			namespace:     orchestratorNamespace,
			namespaceObj:  &corev1.Namespace{},
			expectedError: nil,
		},
		{
			name:          "Create namespace with error",
			namespace:     "fake-namespace",
			namespaceObj:  &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "fake-namespace", Labels: AddLabel()}},
			expectedError: apierrors.NewAlreadyExists(schema.GroupResource{}, "fake-namespace"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			fakeClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(tc.namespaceObj).Build()
			err := CreateNamespace(ctx, fakeClient, tc.namespace)
			assert.Equal(t, apierrors.IsAlreadyExists(tc.expectedError), apierrors.IsAlreadyExists(err))
		})
	}
}

func TestInstallSubscriptionAndOperatorGroup(t *testing.T) {
	ctx := context.TODO()
	scheme := runtime.NewScheme()
	utilruntime.Must(v1alpha1.AddToScheme(scheme))
	utilruntime.Must(operatorsv1.AddToScheme(scheme))

	expectedError := apierrors.NewAlreadyExists(schema.GroupResource{}, "fake-subscription")

	t.Run("Install subscription and operator group no error", func(t *testing.T) {
		fakeClient := fake.NewClientBuilder().WithScheme(scheme).Build()
		fakeOLMClientSet := olmclientsetfake.NewSimpleClientset()
		err := InstallSubscriptionAndOperatorGroup(
			ctx, fakeClient,
			fakeOLMClientSet,
			orchestratorOperatorGroup,
			subscription)
		assert.Equal(t, nil, err)
	})

	t.Run("Install subscription and operator group with error", func(t *testing.T) {
		fakeClient := fake.NewClientBuilder().WithScheme(scheme).Build()
		fakeOLMClientSetWithSubscription := olmclientsetfake.NewSimpleClientset()
		fakeOLMClientSetWithSubscription.OperatorsV1alpha1().
			Subscriptions(orchestratorNamespace).
			Create(ctx, subscription, metav1.CreateOptions{})
		err := InstallSubscriptionAndOperatorGroup(
			ctx, fakeClient,
			fakeOLMClientSetWithSubscription,
			orchestratorOperatorGroup,
			subscription)
		assert.Error(t, err, "Expected error when subscription already exists")
		assert.Equal(t, apierrors.IsAlreadyExists(expectedError), apierrors.IsAlreadyExists(err))
	})
}

func TestApproveInstallPlan(t *testing.T) {
	ctx := context.TODO()
	scheme := runtime.NewScheme()
	utilruntime.Must(v1alpha1.AddToScheme(scheme))

	installPlan := &v1alpha1.InstallPlan{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "install-plan",
			Namespace: orchestratorNamespace,
		},
		Spec: v1alpha1.InstallPlanSpec{
			Approval: v1alpha1.ApprovalManual,
			Approved: false,
		},
	}

	// Test with approve InstallPlan with no errors
	t.Run("Approve install plan", func(t *testing.T) {
		fakeClientWithInstallPlan := fake.NewClientBuilder().WithScheme(scheme).WithObjects(installPlan).Build()
		err := ApproveInstallPlan(fakeClientWithInstallPlan, ctx, installPlan.Name, orchestratorNamespace)
		assert.NoError(t, err, "Expected no error")

		// Verify InstallPlan is approved
		updatedInstallPlan := &v1alpha1.InstallPlan{}
		_ = fakeClientWithInstallPlan.Get(ctx, types.NamespacedName{Name: installPlan.Name, Namespace: installPlan.Namespace}, updatedInstallPlan)
		assert.Equal(t, true, updatedInstallPlan.Spec.Approved)
	})

	// Test approve InstallPlan with error
	t.Run("Approve install plan with error", func(t *testing.T) {
		fakeClientWithoutInstallPlan := fake.NewClientBuilder().WithScheme(scheme).Build()
		err := ApproveInstallPlan(fakeClientWithoutInstallPlan, ctx, installPlan.Name, orchestratorNamespace)
		assert.Error(t, err, "Expected error")
		assert.True(t, apierrors.IsNotFound(err))
	})
}

func TestGetOperatorGroup(t *testing.T) {
	ctx := context.TODO()
	scheme := runtime.NewScheme()
	utilruntime.Must(operatorsv1.AddToScheme(scheme))

	operatorGroup := &operatorsv1.OperatorGroup{
		ObjectMeta: metav1.ObjectMeta{
			Name:      orchestratorOperatorGroup,
			Namespace: orchestratorNamespace,
		},
	}

	// Test get operator group
	t.Run("Get operator group", func(t *testing.T) {
		fakeClientWithOperatorGroup := fake.NewClientBuilder().WithScheme(scheme).WithObjects(operatorGroup).Build()
		err := getOperatorGroup(ctx, fakeClientWithOperatorGroup, orchestratorNamespace, orchestratorOperatorGroup)
		assert.NoError(t, err, "Expected no error")
	})

	// Test create operator group
	t.Run("Create operator group", func(t *testing.T) {
		fakeClientWithoutOperatorGroup := fake.NewClientBuilder().WithScheme(scheme).Build()
		err := getOperatorGroup(ctx, fakeClientWithoutOperatorGroup, orchestratorNamespace, orchestratorOperatorGroup)
		assert.NoError(t, err, "Expected no error")

		createdOperatorGroup := &operatorsv1.OperatorGroup{}
		_ = fakeClientWithoutOperatorGroup.Get(
			ctx,
			types.NamespacedName{Name: orchestratorOperatorGroup, Namespace: orchestratorNamespace},
			createdOperatorGroup)
		assert.Equal(t, createdOperatorGroup.Namespace, operatorGroup.Namespace, "OperatorGroup namespace should match")
		assert.Equal(t, createdOperatorGroup.Name, operatorGroup.Name, "OperatorGroup namespace should match")
	})
}

func TestCreateSubscriptionObject(t *testing.T) {
	actualSubscription := CreateSubscriptionObject(
		subscriptionName,
		orchestratorNamespace,
		subscription.Spec.Channel,
		subscription.Spec.StartingCSV,
	)
	assert.Equal(t, subscription, actualSubscription)
}

func TestCheckSubscriptionExists(t *testing.T) {
	ctx := context.TODO()
	scheme := runtime.NewScheme()
	utilruntime.Must(v1alpha1.AddToScheme(scheme))

	t.Run("Subscription exist", func(t *testing.T) {
		fakeOLMClientSetWithoutSubscription := olmclientsetfake.NewSimpleClientset(subscription)
		subscriptionExist, _, err := CheckSubscriptionExists(ctx, fakeOLMClientSetWithoutSubscription, subscription)
		assert.NoError(t, err, "Expected no error")
		assert.Equal(t, true, subscriptionExist)
	})
	t.Run("Subscription does not exist", func(t *testing.T) {
		fakeOLMClientSetWithSubscription := olmclientsetfake.NewSimpleClientset()
		subscriptionExist, _, err := CheckSubscriptionExists(ctx, fakeOLMClientSetWithSubscription, subscription)
		assert.NoError(t, err, "Expected no error")
		assert.Equal(t, false, subscriptionExist)
	})
}

func TestCheckCRDExists(t *testing.T) {
	ctx := context.TODO()
	scheme := runtime.NewScheme()
	utilruntime.Must(apiextensionsv1.AddToScheme(scheme))

	t.Run("Check CRD exist with error", func(t *testing.T) {
		fakeClientWithoutCRD := fake.NewClientBuilder().WithScheme(scheme).Build()
		err := CheckCRDExists(ctx, fakeClientWithoutCRD, "sonataflowclusterplatforms")
		assert.Error(t, err, "Expected error")
	})

	t.Run("Check CRD exist without error", func(t *testing.T) {
		crdName := "sonataflowclusterplatforms"
		crd := &apiextensionsv1.CustomResourceDefinition{
			ObjectMeta: metav1.ObjectMeta{
				Name: crdName,
			},
		}
		fakeClientWithCRD := fake.NewClientBuilder().WithScheme(scheme).WithObjects(crd).Build()
		err := CheckCRDExists(ctx, fakeClientWithCRD, crdName)
		assert.NoError(t, err, "Expected no error")
	})
}

func TestCleanUpNamespace(t *testing.T) {
	ctx := context.TODO()
	scheme := runtime.NewScheme()
	utilruntime.Must(corev1.AddToScheme(scheme))

	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{Name: orchestratorNamespace, Labels: AddLabel()},
	}
	t.Run("Clean up namespace with no error", func(t *testing.T) {
		fakeClientWithoutNS := fake.NewClientBuilder().WithScheme(scheme).WithObjects(ns).Build()
		err := CleanUpNamespace(ctx, orchestratorNamespace, fakeClientWithoutNS)
		assert.NoError(t, err, "Expected no error")
	})
}

func TestAddLabel(t *testing.T) {
	expectedLabels := map[string]string{
		CreatedByLabelKey: CreatedByLabelValue,
	}

	t.Run("Add label", func(t *testing.T) {
		labelMap := AddLabel()
		assert.NotNil(t, labelMap, "Expected labelMap to not be nil")
		assert.Equal(t, expectedLabels, labelMap, "Expected labelMap to match expectedLabels")
		assert.Equal(t, len(expectedLabels), len(labelMap), "Expected labelMap to have the same length")
	})

}

func TestCheckLabelExists(t *testing.T) {
	testCases := []struct {
		name          string
		labelMap      map[string]string
		expectedExist bool
	}{
		{
			name:          "Check label exists",
			labelMap:      existingLabelMap,
			expectedExist: true,
		},
		{
			name: "Check label doesn't exist",
			labelMap: map[string]string{
				"app.kubernetes.io/created-by": "kustomize",
			},
			expectedExist: false,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			labelExist := CheckLabelExist(tc.labelMap)
			assert.Equal(t, tc.expectedExist, labelExist)
		})
	}
}
