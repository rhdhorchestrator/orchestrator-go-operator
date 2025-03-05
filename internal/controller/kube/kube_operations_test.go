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
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"testing"
)

func TestCheckNamespaceExist(t *testing.T) {
	ctx := context.TODO()
	// Create a fake client scheme
	scheme := runtime.NewScheme()
	utilruntime.Must(corev1.AddToScheme(scheme))

	orchestratorNamespaceName := "orchestrator-namespace"

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

	orchestratorNamespaceName := "orchestrator-namespace"
	ns := &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: orchestratorNamespaceName, Labels: AddLabel()}}

	// Test create namespace
	fakeClientWithNS := fake.NewClientBuilder().WithScheme(scheme).Build()
	err := CreateNamespace(ctx, fakeClientWithNS, orchestratorNamespaceName)
	assert.NoError(t, err, "Expected no error")

	// Test create namespace with error
	fakeClientWithoutNS := fake.NewClientBuilder().WithScheme(scheme).WithObjects(ns).Build()
	err = CreateNamespace(ctx, fakeClientWithoutNS, orchestratorNamespaceName)
	assert.Error(t, err, "Expected an error when namespace does not exist")
}
