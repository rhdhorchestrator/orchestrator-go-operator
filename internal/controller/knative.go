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

package controller

import (
	"context"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	knative "knative.dev/operator/pkg/apis/operator/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

const (
	KnativeAPIVersion             = "operator.knative.dev/v1beta1"
	KnativeServingKind            = "KnativeServing"
	KnativeServingNamespacedName  = "knative-serving"
	KnativeEventingKind           = "KnativeEventing"
	KnativeEventingNamespacedName = "knative-eventing"
)

func handleKnativeEventingCR(ctx context.Context, client client.Client) error {
	logger := log.FromContext(ctx)
	// check CR exists
	knativeEventingCR := &knative.KnativeEventing{}
	err := client.Get(ctx, types.NamespacedName{Name: KnativeEventingNamespacedName, Namespace: KnativeEventingNamespacedName}, knativeEventingCR)
	if err == nil {
		// update CR TODO
		return nil
	} else {
		if apierrors.IsNotFound(err) {
			knEventing := &knative.KnativeEventing{
				TypeMeta: metav1.TypeMeta{
					APIVersion: KnativeAPIVersion,
					Kind:       KnativeEventingKind,
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      KnativeEventingNamespacedName,
					Namespace: KnativeEventingNamespacedName,
				},
				Spec: knative.KnativeEventingSpec{},
				//Status: knative.KnativeEventingStatus{},
			}
			if err = client.Create(ctx, knEventing); err != nil {
				logger.Error(err, "Error occurred when creating CR resource", "CR-Name", knativeEventingCR.Name)
			}
			logger.Info("Successfully created Knative Eventing resource", "CR-Name", knativeEventingCR.Name)
		}
	}
	return err
}

func handleKnativeServingCR(ctx context.Context, client client.Client) error {
	logger := log.FromContext(ctx)
	// check CR exists
	knativeServingCR := &knative.KnativeServing{}
	err := client.Get(ctx, types.NamespacedName{Name: KnativeServingNamespacedName, Namespace: KnativeServingNamespacedName}, knativeServingCR)
	if err == nil {
		// update CR TODO
		return nil
	} else {
		if apierrors.IsNotFound(err) {
			knEventing := &knative.KnativeEventing{
				TypeMeta: metav1.TypeMeta{
					APIVersion: KnativeAPIVersion,
					Kind:       KnativeServingKind,
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      KnativeServingNamespacedName,
					Namespace: KnativeServingNamespacedName,
				},
				Spec: knative.KnativeEventingSpec{},
				//Status: knative.KnativeEventingStatus{},
			}
			if err = client.Create(ctx, knEventing); err != nil {
				logger.Error(err, "Error occurred when creating CR resource", "CR-Name", knativeServingCR.Name)
			}
			logger.Info("Successfully created Knative Serving resource", "CR-Name", knativeServingCR.Name)
		}
	}
	return err
}
