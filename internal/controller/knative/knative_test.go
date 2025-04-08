package knative

import (
	"context"
	"testing"

	operatorsv1 "github.com/operator-framework/api/pkg/operators/v1"
	"github.com/operator-framework/api/pkg/operators/v1alpha1"
	olmclientsetfake "github.com/operator-framework/operator-lifecycle-manager/pkg/api/client/clientset/versioned/fake"
	"github.com/rhdhorchestrator/orchestrator-operator/internal/controller/kube"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apierrros "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	Knative "knative.dev/operator/pkg/apis/operator/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

const (
	testEventingName          = "test-name"
	testServingName           = "test-name"
	testChannel               = "test-channel"
	orchestratorNamespace     = "orchestrator-namespace"
	orchestratorOperatorGroup = "orchestrator-operator-group"
	subscriptionName          = "orchestrator-subscription"
)

var objects []client.Object

func TestHandleKNativeOperatorInstallation(t *testing.T) {
	ctx := context.TODO()
	// Create a fake client scheme
	scheme := runtime.NewScheme()
	utilruntime.Must(Knative.AddToScheme(scheme))
	utilruntime.Must(corev1.AddToScheme(scheme))
	utilruntime.Must(v1alpha1.AddToScheme(scheme))
	utilruntime.Must(apiextensionsv1.AddToScheme(scheme))
	utilruntime.Must(metav1.AddMetaToScheme(scheme))
	utilruntime.Must(operatorsv1.AddToScheme(scheme))

	testCases := []struct {
		name                 string
		subExists            bool
		expectedErrorMessage string
	}{
		{
			name:                 "Subscription exists",
			subExists:            true,
			expectedErrorMessage: "",
		},
		{
			name:                 "Subscription does not exist",
			subExists:            false,
			expectedErrorMessage: "subscriptions.operators.coreos.com \"serverless-operator\" not found",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			desiredSubscription := kube.CreateSubscriptionObject(
				KnativeSubscriptionName,
				KnativeOperatorNamespace,
				testChannel,
				KnativeSubscriptionStartingCSV)

			desiredSubscription.Status = v1alpha1.SubscriptionStatus{
				InstallPlanRef: &corev1.ObjectReference{Name: "test-plan"},
				CurrentCSV:     "test-csv",
			}

			builder := fake.NewClientBuilder().WithScheme(scheme)
			if tc.subExists {
				builder.WithObjects(desiredSubscription)

			}
			fakeClient := builder.Build()
			fakeOLMClientSet := olmclientsetfake.NewSimpleClientset()
			err := kube.InstallSubscriptionAndOperatorGroup(
				ctx, fakeClient,
				fakeOLMClientSet,
				orchestratorOperatorGroup,
				desiredSubscription)
			assert.Equal(t, nil, err)

			err = HandleKNativeOperatorInstallation(ctx, fakeClient, fakeOLMClientSet)
			if tc.subExists {
				assert.NoError(t, err)
			} else {
				assert.Equal(t, err.Error(), tc.expectedErrorMessage)
			}

			namespace := desiredSubscription.Namespace
			subscriptionName := desiredSubscription.Name

			subscription, err := fakeOLMClientSet.OperatorsV1alpha1().Subscriptions(namespace).Get(ctx, subscriptionName, metav1.GetOptions{})
			if tc.subExists {
				if subscription != nil {
					assert.Equal(t, subscription.Spec.Channel, desiredSubscription.Spec.Channel)
				}
			}
			assert.NoError(t, err)
		})
	}
}

func TestHandleKnativeCR(t *testing.T) {
	ctx := context.TODO()
	// Create a fake client scheme
	scheme := runtime.NewScheme()
	utilruntime.Must(Knative.AddToScheme(scheme))
	utilruntime.Must(corev1.AddToScheme(scheme))
	utilruntime.Must(apiextensionsv1.AddToScheme(scheme))
	utilruntime.Must(metav1.AddMetaToScheme(scheme))

	testCases := []struct {
		name                 string
		eventingCRDExists    bool
		servingCRDExists     bool
		expectedErrorMessage string
	}{
		{
			name:                 "Both CRD exists",
			eventingCRDExists:    true,
			servingCRDExists:     true,
			expectedErrorMessage: "",
		},
		{
			name:                 "Only eventing exists",
			eventingCRDExists:    true,
			servingCRDExists:     false,
			expectedErrorMessage: "customresourcedefinitions.apiextensions.k8s.io \"knativeservings.operator.knative.dev\" not found",
		},
		{
			name:                 "Only serving CRD exists",
			eventingCRDExists:    false,
			servingCRDExists:     true,
			expectedErrorMessage: "customresourcedefinitions.apiextensions.k8s.io \"knativeeventings.operator.knative.dev\" not found",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			eventingCRD := &apiextensionsv1.CustomResourceDefinition{
				TypeMeta: metav1.TypeMeta{},
				ObjectMeta: metav1.ObjectMeta{
					Name: KnativeEventingCRDName,
				},
				Spec: apiextensionsv1.CustomResourceDefinitionSpec{},
			}
			servingCRD := &apiextensionsv1.CustomResourceDefinition{
				TypeMeta: metav1.TypeMeta{},
				ObjectMeta: metav1.ObjectMeta{
					Name: KnativeServingCRDName,
				},
				Spec: apiextensionsv1.CustomResourceDefinitionSpec{},
			}
			builder := fake.NewClientBuilder().WithScheme(scheme)
			if tc.eventingCRDExists {
				builder.WithObjects(eventingCRD)
			}
			if tc.servingCRDExists {
				builder.WithObjects(servingCRD)
			}
			fakeClient := builder.Build()

			err := HandleKnativeCR(ctx, fakeClient)
			if err != nil {
				assert.Equal(t, tc.expectedErrorMessage, err.Error())
			}

		})
	}
}

func TestHandleKnativeEventingCR(t *testing.T) {
	ctx := context.TODO()
	// Create a fake client scheme
	scheme := runtime.NewScheme()
	utilruntime.Must(Knative.AddToScheme(scheme))
	utilruntime.Must(corev1.AddToScheme(scheme))
	utilruntime.Must(metav1.AddMetaToScheme(scheme))

	testCases := []struct {
		name           string
		eventingExists bool
		expectedError  error
		eventingObject *Knative.KnativeEventing
	}{
		{
			name:           "Update existing Knative Eventing",
			eventingExists: true,
			expectedError:  nil,
			eventingObject: &Knative.KnativeEventing{
				TypeMeta: metav1.TypeMeta{
					APIVersion: KnativeAPIVersion,
					Kind:       KnativeEventingKind,
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      testEventingName, // This is the field that indicates that an update has been performed
					Namespace: KnativeEventingNamespacedName,
				},
				Spec: Knative.KnativeEventingSpec{},
			},
		},
		{
			name:           "Create Knative Eventing",
			eventingExists: false,
			expectedError:  nil,
			eventingObject: &Knative.KnativeEventing{},
		},
	}

	for _, tc := range testCases {
		existingEventing := &Knative.KnativeEventing{}

		if !tc.eventingExists {
			t.Run(tc.name, func(t *testing.T) {
				fakeClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects().Build()
				err := HandleKnativeEventingCR(ctx, fakeClient)
				assert.Equal(t, tc.expectedError, err)
				err = fakeClient.Get(ctx, types.NamespacedName{Name: KnativeEventingNamespacedName, Namespace: KnativeEventingNamespacedName}, existingEventing)
				assert.NoError(t, err)
			})

		} else {

			t.Run(tc.name, func(t *testing.T) {
				fakeClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(tc.eventingObject).Build()

				err := HandleKnativeEventingCR(ctx, fakeClient)
				assert.Equal(t, tc.expectedError, err)
				err = fakeClient.Get(ctx, types.NamespacedName{Name: testEventingName, Namespace: KnativeEventingNamespacedName}, existingEventing)
				assert.NoError(t, err)
				assert.Equal(t, existingEventing, tc.eventingObject)
			})
		}
	}
}

func TestHandleKnativeServingCR(t *testing.T) {
	ctx := context.TODO()
	// Create a fake client scheme
	scheme := runtime.NewScheme()
	utilruntime.Must(Knative.AddToScheme(scheme))
	utilruntime.Must(corev1.AddToScheme(scheme))
	utilruntime.Must(metav1.AddMetaToScheme(scheme))

	testCases := []struct {
		name          string
		servingExists bool
		expectedError error
		servingObject *Knative.KnativeServing
	}{
		{
			name:          "Update existing Serving Eventing",
			servingExists: true,
			expectedError: nil,
			servingObject: &Knative.KnativeServing{
				TypeMeta: metav1.TypeMeta{
					APIVersion: KnativeAPIVersion,
					Kind:       KnativeServingKind,
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      testServingName, // This is the field that indicates that an update has been performed
					Namespace: KnativeServingNamespacedName,
				},
				Spec: Knative.KnativeServingSpec{},
			},
		},
		{
			name:          "Create Knative Serving",
			servingExists: false,
			expectedError: nil,
			servingObject: &Knative.KnativeServing{},
		},
	}

	for _, tc := range testCases {
		existingServing := &Knative.KnativeServing{}

		if !tc.servingExists {
			t.Run(tc.name, func(t *testing.T) {

				fakeClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects().Build()
				err := HandleKnativeServingCR(ctx, fakeClient)
				assert.Equal(t, tc.expectedError, err)

				err = fakeClient.Get(ctx, types.NamespacedName{Name: KnativeServingNamespacedName, Namespace: KnativeServingNamespacedName}, existingServing)
				assert.NoError(t, err)
			})

		} else {
			t.Run(tc.name, func(t *testing.T) {
				fakeClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(tc.servingObject).Build()
				err := HandleKnativeServingCR(ctx, fakeClient)
				assert.Equal(t, tc.expectedError, err)

				err = fakeClient.Get(ctx, types.NamespacedName{Name: testServingName, Namespace: KnativeServingNamespacedName}, existingServing)
				assert.NoError(t, err)
				assert.Equal(t, existingServing, tc.servingObject)
			})
		}
	}

}

func TestHandleKnativeCleanUp(t *testing.T) {
	ctx := context.TODO()
	// Create a fake client scheme
	scheme := runtime.NewScheme()
	utilruntime.Must(Knative.AddToScheme(scheme))
	utilruntime.Must(corev1.AddToScheme(scheme))
	utilruntime.Must(metav1.AddMetaToScheme(scheme))

	testCases := []struct {
		name       string
		namespaces []*corev1.Namespace
	}{
		{
			name: "Validate namespaces were deleted",
			namespaces: []*corev1.Namespace{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      KnativeEventingNamespacedName,
						Namespace: KnativeEventingNamespacedName,
					},
					Spec: corev1.NamespaceSpec{},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      KnativeOperatorNamespace,
						Namespace: KnativeOperatorNamespace,
					},
					Spec: corev1.NamespaceSpec{},
				}, {
					ObjectMeta: metav1.ObjectMeta{
						Name:      KnativeServingNamespacedName,
						Namespace: KnativeServingNamespacedName,
					},
					Spec: corev1.NamespaceSpec{},
				},
			},
		},
	}
	tc := testCases[0]
	t.Run(tc.name, func(t *testing.T) {
		for _, ns := range tc.namespaces {
			objects = append(objects, ns)
		}
		fakeClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(objects...).Build()

		for _, ns := range tc.namespaces {
			namespace := &corev1.Namespace{}
			err := fakeClient.Get(ctx, types.NamespacedName{Name: ns.Name}, namespace)
			assert.True(t, apierrros.IsNotFound(err))
		}

		err := HandleKnativeCleanUp(ctx, fakeClient)
		assert.NoError(t, err)

		for _, ns := range tc.namespaces {
			namespace := &corev1.Namespace{}
			err = fakeClient.Get(ctx, types.NamespacedName{Name: ns.Name}, namespace)
			assert.True(t, apierrros.IsNotFound(err))
		}
	})
}
