package controller

import (
	"context"
	"github.com/operator-framework/api/pkg/operators/v1alpha1"
	"github.com/rhdhorchestrator/orchestrator-operator/internal/controller/kube"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"testing"

	olmclientsetfake "github.com/operator-framework/operator-lifecycle-manager/pkg/api/client/clientset/versioned/fake"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

type MockKubeHelper struct {
	mock.Mock
}

var knativeSubscription = &v1alpha1.Subscription{
	ObjectMeta: metav1.ObjectMeta{Namespace: knativeOperatorNamespace, Name: knativeSubscriptionName},
	Spec: &v1alpha1.SubscriptionSpec{
		Channel:                knativeSubscriptionChannel,
		InstallPlanApproval:    v1alpha1.ApprovalManual,
		CatalogSource:          kube.CatalogSourceName,
		StartingCSV:            knativeSubscriptionStartingCSV,
		CatalogSourceNamespace: kube.CatalogSourceNamespace,
		Package:                knativeSubscriptionName,
	},
}

func (m *MockKubeHelper) CheckNamespaceExist(ctx context.Context, client client.Client, namespace string) (bool, error) {
	args := m.Called(ctx, client, namespace)
	return args.Bool(0), args.Error(1)
}

func (m *MockKubeHelper) CreateNamespace(ctx context.Context, client client.Client, namespace string) error {
	args := m.Called(ctx, client, namespace)
	return args.Error(0)
}

func (m *MockKubeHelper) CheckSubscriptionExists(ctx context.Context, olmclientset olmclientsetfake.Clientset, knativeSubscription v1alpha1.Subscription) (bool, *v1alpha1.Subscription, error) {
	args := m.Called(ctx, olmclientset, knativeSubscription)
	return args.Bool(0), args.Get(1).(*v1alpha1.Subscription), args.Error(2)
}

func (m *MockKubeHelper) InstallSubscriptionAndOperatorGroup(ctx context.Context, client client.Client, olmclientsetfake olmclientsetfake.Clientset, operatorGroupName string, subscription v1alpha1.Subscription) error {
	args := m.Called(ctx, client, olmclientsetfake, operatorGroupName, knativeSubscription)
	return args.Error(0)
}

func TestHandleKNativeOperatorInstallation(t *testing.T) {
	// Create a test context
	ctx := context.TODO()

	// Create a fake client scheme
	scheme := runtime.NewScheme()

	// Create a fake Kubernetes client
	fakeClient := fake.NewClientBuilder().WithScheme(scheme).Build()

	// Create a fake OLM clientset
	fakeOLMClientSet := olmclientsetfake.NewSimpleClientset()

	// Create mock for kube helper functions
	mockKubeHelper := new(MockKubeHelper)
	mockKubeHelper.On("CheckNamespaceExist", ctx, fakeClient, knativeOperatorNamespace).Return(false, nil)
	mockKubeHelper.On("CreateNamespace", ctx, fakeClient, knativeOperatorNamespace).Return(nil)
	mockKubeHelper.On("CheckSubscriptionExists", ctx, fakeOLMClientSet, mock.Anything).Return(false, &v1alpha1.Subscription{}, nil)
	mockKubeHelper.On("InstallSubscriptionAndOperatorGroup", ctx, fakeClient, fakeOLMClientSet, knativeOperatorGroupName, mock.Anything).Return(nil)

	// Call the function under test
	err := handleKNativeOperatorInstallation(ctx, fakeClient, *fakeOLMClientSet)

	// Validate expected results
	assert.NoError(t, err, "Expected no error from handleKNativeOperatorInstallation")

	// Verify function calls
	mockKubeHelper.AssertCalled(t, "CheckNamespaceExist", ctx, fakeClient, knativeOperatorNamespace)
	mockKubeHelper.AssertCalled(t, "CreateNamespace", ctx, fakeClient, knativeOperatorNamespace)
	mockKubeHelper.AssertCalled(t, "CheckSubscriptionExists", ctx, fakeOLMClientSet, mock.Anything)
	mockKubeHelper.AssertCalled(t, "InstallSubscriptionAndOperatorGroup", ctx, fakeClient, fakeOLMClientSet, knativeOperatorGroupName, mock.Anything)
}
