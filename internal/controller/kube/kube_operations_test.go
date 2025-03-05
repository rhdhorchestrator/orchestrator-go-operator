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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"testing"
)

var (
	orchestratorNamespaceName = "orchestrator-namespace"
	orchestratorOperatorGroup = "orchestrator-operator-group"
)

func TestCheckNamespaceExist(t *testing.T) {
	ctx := context.TODO()
	// Create a fake client scheme
	scheme := runtime.NewScheme()
	utilruntime.Must(corev1.AddToScheme(scheme))

	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{Name: orchestratorNamespaceName},
	}

	// Test Namespace exists
	fakeClientWithNS := fake.NewClientBuilder().WithScheme(scheme).WithObjects(ns).Build()
	exists, err := CheckNamespaceExist(ctx, fakeClientWithNS, orchestratorNamespaceName)
	assert.NoError(t, err, "Expected no error when namespace exists")
	assert.True(t, exists, "Expected namespace to exist")

	// Test Namespace does not exist
	fakeClientWithoutNS := fake.NewClientBuilder().WithScheme(scheme).Build()
	exists, err = CheckNamespaceExist(ctx, fakeClientWithoutNS, orchestratorNamespaceName)
	assert.Error(t, err, "Expected an error when namespace does not exist")
	assert.False(t, exists, "Expected namespace to not exist")
}

func TestCreateNamespace(t *testing.T) {
	ctx := context.TODO()
	scheme := runtime.NewScheme()
	utilruntime.Must(corev1.AddToScheme(scheme))

	ns := &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: orchestratorNamespaceName, Labels: AddLabel()}}

	// Test create namespace
	fakeClientWithoutNS := fake.NewClientBuilder().WithScheme(scheme).Build()
	err := CreateNamespace(ctx, fakeClientWithoutNS, orchestratorNamespaceName)
	assert.NoError(t, err, "Expected no error")

	// Test create namespace with error
	fakeClientWithNS := fake.NewClientBuilder().WithScheme(scheme).WithObjects(ns).Build()
	err = CreateNamespace(ctx, fakeClientWithNS, orchestratorNamespaceName)
	assert.Error(t, err, "Expected an error when namespace does not exist")
}

func TestInstallSubscriptionAndOperatorGroup(t *testing.T) {
	ctx := context.TODO()
	scheme := runtime.NewScheme()
	utilruntime.Must(v1alpha1.AddToScheme(scheme))
	utilruntime.Must(operatorsv1.AddToScheme(scheme))

	subscriptionName := "orchestrator-subscription"
	subscription := &v1alpha1.Subscription{
		ObjectMeta: metav1.ObjectMeta{
			Name:      subscriptionName,
			Namespace: orchestratorNamespaceName,
		},
		Spec: &v1alpha1.SubscriptionSpec{
			Channel:                "channel",
			StartingCSV:            "starting-csv",
			InstallPlanApproval:    v1alpha1.ApprovalManual,
			CatalogSource:          CatalogSourceName,
			CatalogSourceNamespace: CatalogSourceNamespace,
		},
	}

	fakeClient := fake.NewClientBuilder().WithScheme(scheme).Build()

	// Test subscription is created
	fakeOLMClientSetWithoutSubscription := olmclientsetfake.NewSimpleClientset()
	err := InstallSubscriptionAndOperatorGroup(
		ctx, fakeClient,
		fakeOLMClientSetWithoutSubscription,
		orchestratorOperatorGroup,
		subscription)
	assert.NoError(t, err, "Expected no error")

	// Test create subscription with error
	fakeOLMClientSetWithSubscription := olmclientsetfake.NewSimpleClientset()
	fakeOLMClientSetWithSubscription.OperatorsV1alpha1().
		Subscriptions(orchestratorNamespaceName).
		Create(ctx, subscription, metav1.CreateOptions{})

	err = InstallSubscriptionAndOperatorGroup(
		ctx,
		fakeClient,
		fakeOLMClientSetWithSubscription,
		orchestratorOperatorGroup,
		subscription)
	assert.Error(t, err, "Expected error")
}
