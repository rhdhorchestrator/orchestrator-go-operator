package rhdh

import (
	"context"
	"fmt"
	olmclientset "github.com/operator-framework/operator-lifecycle-manager/pkg/api/client/clientset/versioned"
	orchestratorv1alpha2 "github.com/rhdhorchestrator/orchestrator-operator/api/v1alpha3"
	kubeoperations "github.com/rhdhorchestrator/orchestrator-operator/internal/controller/kube"
	"github.com/rhdhorchestrator/orchestrator-operator/internal/controller/util"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	rhdhv1alpha3 "redhat-developer/red-hat-developer-hub-operator/api/v1alpha3"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

const (
	rhdhOperatorGroup                 = "rhdh-operator-group"
	rhdhAPIVersion                    = "rhdh.redhat.com/v1alpha2"
	rhdhKind                          = "Backstage"
	rhdhCRDName                       = "backstages.rhdh.redhat.com"
	rhdhReplica                 int32 = 1
	rhdhSubscriptionName              = "rhdh"
	rhdhSubscriptionChannel           = "fast-1.4"
	rhdhOperatorNamespace             = "rhdh-operator"
	rhdhSubscriptionStartingCSV       = "rhdh-operator.v1.4.1"
)

var ConfigMapNameAndConfigDataKey = map[string]string{
	AppConfigRHDHName:              "app-config-rhdh.yaml",
	AppConfigRHDHAuthName:          "app-config-auth.gh.yaml",
	AppConfigRHDHCatalogName:       "app-config-catalog.yaml",
	AppConfigRHDHDynamicPluginName: "dynamic-plugins.yaml",
}

func HandleRHDHOperatorInstallation(ctx context.Context, client client.Client, olmClientSet olmclientset.Clientset) error {
	rhdhLogger := log.FromContext(ctx)

	if _, err := kubeoperations.CheckNamespaceExist(ctx, client, rhdhOperatorNamespace); err != nil {
		if apierrors.IsNotFound(err) {
			if err := kubeoperations.CreateNamespace(ctx, client, rhdhOperatorNamespace); err != nil {
				rhdhLogger.Error(err, "Error occurred when creating namespace for RHDH operator", "NS", rhdhOperatorNamespace)
				return nil
			}
		}
		rhdhLogger.Error(err, "Error occurred when checking namespace exists for RHDH operator", "NS", rhdhOperatorNamespace)
		return err
	}

	// check if subscription exist
	rhdhSubscription := kubeoperations.CreateSubscriptionObject(
		rhdhSubscriptionName,
		rhdhOperatorNamespace,
		rhdhSubscriptionChannel,
		rhdhSubscriptionStartingCSV)

	// check if subscription exists
	subscriptionExists, existingSubscription, err := kubeoperations.CheckSubscriptionExists(ctx, olmClientSet, rhdhSubscription)
	if err != nil {
		rhdhLogger.Error(err, "Error occurred when checking subscription exists", "SubscriptionName", rhdhSubscriptionName)
		return err
	}
	if !subscriptionExists {
		if err := kubeoperations.InstallSubscriptionAndOperatorGroup(
			ctx, client, olmClientSet,
			rhdhOperatorGroup, rhdhSubscription); err != nil {
			rhdhLogger.Error(err, "Error occurred when installing operator", "SubscriptionName", rhdhSubscriptionName)
			return err
		}
		rhdhLogger.Info("Operator successfully installed", "SubscriptionName", rhdhSubscriptionName)
	} else {
		// Compare the current and desired state
		if !reflect.DeepEqual(existingSubscription.Spec, rhdhSubscription.Spec) {
			// Update the existing subscription with the new Spec
			existingSubscription.Spec = rhdhSubscription.Spec
			if err := client.Update(ctx, existingSubscription); err != nil {
				rhdhLogger.Error(err, "Error occurred when updating subscription spec", "SubscriptionName", rhdhSubscriptionName)
				return err
			}
			rhdhLogger.Info("Successfully updated subscription spec", "SubscriptionName", rhdhSubscriptionName)
		}
	}

	// approve install plan
	if existingSubscription.Status.InstallPlanRef != nil && existingSubscription.Status.CurrentCSV == rhdhSubscriptionStartingCSV {
		installPlanName := existingSubscription.Status.InstallPlanRef.Name
		if err := kubeoperations.ApproveInstallPlan(client, ctx, installPlanName, existingSubscription.Namespace); err != nil {
			rhdhLogger.Error(err, "Error occurred while approving install plan for subscription", "SubscriptionName", installPlanName)
			return err
		}
	}
	return nil
}

func CreateRHDHSecret(secretNamespace string, ctx context.Context, client client.Client) error {
	logger := log.FromContext(ctx)
	logger.Info("Creating RHDH NPMrc Secret")

	secret := &corev1.Secret{}
	err := client.Get(ctx, types.NamespacedName{
		Namespace: secretNamespace,
		Name:      RegistrySecretName,
	}, secret)

	if err != nil {
		if apierrors.IsNotFound(err) {
			logger.Info("Secret does not exist. Creating secret", "Secret", RegistrySecretName)
			newSecret := &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      RegistrySecretName,
					Namespace: secretNamespace,
					Labels:    kubeoperations.AddLabel(),
				},
				Type: corev1.SecretTypeOpaque,
				StringData: map[string]string{
					".npmrc": fmt.Sprintf("registry=%s", NpmRegistry),
				},
			}

			if err := client.Create(ctx, newSecret); err != nil {
				logger.Error(err, "Error occurred when creating secret", "Secret", RegistrySecretName)
				return err
			}
			logger.Info("Successfully created secret", "Secret", RegistrySecretName)
			return nil
		}
		logger.Error(err, "Error occurred when checking secret exist", "Secret", RegistrySecretName)
		return err
	}
	logger.Info("Secret already exist", "Secret", RegistrySecretName)
	return nil
}

func HandleRHDHCR(
	rhdhConfig orchestratorv1alpha2.RHDHConfig,
	bsConfigMapList []rhdhv1alpha3.FileObjectRef,
	ctx context.Context, client client.Client) error {
	rhdhLogger := log.FromContext(ctx)

	// subscription exists; check if CRD exists for RHDH
	if err := kubeoperations.CheckCRDExists(ctx, client, rhdhCRDName); err != nil {
		if apierrors.IsNotFound(err) {
			rhdhLogger.Info("CRD resource not found or ready", "CRD", rhdhCRDName)
			return err
		}
		rhdhLogger.Error(err, "Error occurred when retrieving CRD", "CRD", rhdhCRDName)
		return err
	}

	rhdhLogger.Info("Handling RHDH CR resource")

	rhdhNamespace := rhdhConfig.Namespace
	rhdhName := rhdhConfig.Name

	if err := client.Get(ctx, types.NamespacedName{Namespace: rhdhNamespace, Name: rhdhName}, &rhdhv1alpha3.Backstage{}); err != nil {
		if apierrors.IsNotFound(err) {
			secret := rhdhv1alpha3.EnvObjectRef{
				Name: BackendAuthSecretName,
			}
			backstageCR := &rhdhv1alpha3.Backstage{
				TypeMeta: metav1.TypeMeta{
					APIVersion: rhdhAPIVersion,
					Kind:       rhdhKind,
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      rhdhName,
					Namespace: rhdhConfig.Namespace,
					Labels:    kubeoperations.AddLabel(),
				},
				Spec: rhdhv1alpha3.BackstageSpec{
					Application: &rhdhv1alpha3.Application{
						AppConfig:                   &rhdhv1alpha3.AppConfig{ConfigMaps: bsConfigMapList},
						DynamicPluginsConfigMapName: AppConfigRHDHDynamicPluginName,
						ExtraEnvs: &rhdhv1alpha3.ExtraEnvs{
							Secrets: []rhdhv1alpha3.EnvObjectRef{secret},
						},
						Replicas: util.MakePointer(rhdhReplica),
					},
				},
			}
			rhdhLogger.Info("Creating Backstage CR", "CR-Name", backstageCR.Name)
			if err := client.Create(ctx, backstageCR); err != nil {
				rhdhLogger.Error(err, "Error occurred when creating RHDH resource", "CR-Name", rhdhName)
				return err
			}
			rhdhLogger.Info("Successfully created RHDH resource", "CR-Name", rhdhName)
			return nil
		}
		rhdhLogger.Error(err, "Error occurred when retrieving RHDH resource", "CR-Name", rhdhName)
		return err
	}
	return nil
}

// GetOrCreateConfigMaps creates or gets the configmap list
func GetOrCreateConfigMaps(ctx context.Context, client client.Client,
	clusterDomain, serverlessWorkflowNamespace string,
	tektonEnabled, argoCDEnabled bool,
	rhdhConfig orchestratorv1alpha2.RHDHConfig) ([]rhdhv1alpha3.FileObjectRef, error) {

	cmLogger := log.FromContext(ctx)
	cmLogger.Info("Processing ConfigMaps...")

	configmapList := make([]rhdhv1alpha3.FileObjectRef, 0)
	namespace := rhdhConfig.Namespace
	for cmName, configDataKey := range ConfigMapNameAndConfigDataKey {
		if cmName != AppConfigRHDHDynamicPluginName {
			configmapList = append(configmapList, rhdhv1alpha3.FileObjectRef{Name: cmName})
		}
		cmLogger.Info("Starting Configmap creation for:", "CM", cmName, "NS", namespace)

		err := client.Get(ctx, types.NamespacedName{Namespace: namespace, Name: cmName}, &corev1.ConfigMap{})
		if err != nil {
			if apierrors.IsNotFound(err) {
				cmLogger.Info("Configmap does not exist, creating CM", "CM", cmName)
				configValue, err := ConfigMapTemplateFactory(cmName, clusterDomain, serverlessWorkflowNamespace, argoCDEnabled, tektonEnabled, rhdhConfig)
				if err != nil {
					cmLogger.Error(err, "Error occurred when parsing config data for configmap", "CM", cmName)
					return configmapList, fmt.Errorf("failed to parse template data for configmap: %s", err)
				} else {
					if err := CreateConfigMap(cmName, configDataKey, namespace, configValue, ctx, client); err != nil {
						cmLogger.Error(err, "Error occurred when creating ConfigMap", "CM", cmName)
						return configmapList, err
					}
				}
			}
		}
	}
	return configmapList, nil
}

func CreateConfigMap(
	name string, configDataKey string, namespace string, configValue string,
	ctx context.Context, client client.Client) error {

	logger := log.FromContext(ctx)

	// Create the ConfigMap object
	configMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels:    kubeoperations.AddLabel(),
		},
		Data: map[string]string{
			configDataKey: configValue,
		},
	}
	if err := client.Create(ctx, configMap); err != nil {
		logger.Error(err, "Error occurred when creating ConfigMap", "CM", name)
		return err
	}
	logger.Info("Successfully created ConfigMap", "CM", name)
	return nil
}

func HandleRHDHCleanUp(ctx context.Context, client client.Client, olmClientSet olmclientset.Clientset, rhdhNamespace string) error {
	rhdhLogger := log.FromContext(ctx)

	namespaceExist, _ := kubeoperations.CheckNamespaceExist(ctx, client, rhdhNamespace)
	if namespaceExist {
		backstageCRList, err := listBackstageCRs(ctx, client, rhdhNamespace)

		if err != nil || len(backstageCRList) == 0 {
			rhdhLogger.Error(err, "Failed to list RHDH CRs or have no RHDH CRs created by Orchestrator Operator and cannot perform clean up process")
			return err
		}
		if len(backstageCRList) == 1 {
			// remove namespace
			if err := kubeoperations.CleanUpNamespace(ctx, rhdhNamespace, client); err != nil {
				rhdhLogger.Error(err, "Error occurred when deleting namespace", "NS", "namespace")
				return err
			}
		}
	}

	// remove operator namespace
	if err := kubeoperations.CleanUpNamespace(ctx, rhdhOperatorNamespace, client); err != nil {
		rhdhLogger.Error(err, "Error occurred when deleting namespace", "NS", rhdhOperatorNamespace)
		return err
	}
	return nil
}

func listBackstageCRs(ctx context.Context, k8client client.Client, namespace string) ([]rhdhv1alpha3.Backstage, error) {
	rhdhLogger := log.FromContext(ctx)

	crList := &rhdhv1alpha3.BackstageList{}

	listOptions := []client.ListOption{
		client.InNamespace(namespace),
		client.MatchingLabels{kubeoperations.CreatedByLabelKey: kubeoperations.CreatedByLabelValue},
	}

	// List the CRs
	if err := k8client.List(ctx, crList, listOptions...); err != nil {
		rhdhLogger.Error(err, "Error occurred when listing RHDH CRs")
		return nil, err
	}

	rhdhLogger.Info("Successfully listed RHDH CRs", "Total", len(crList.Items))
	return crList.Items, nil
}
