package rhdh

import (
	"context"
	"fmt"
	olmclientset "github.com/operator-framework/operator-lifecycle-manager/pkg/api/client/clientset/versioned"
	orchestratorv1alpha1 "github.com/parodos-dev/orchestrator-operator/api/v1alpha1"
	operations "github.com/parodos-dev/orchestrator-operator/internal/controller/kube"
	"github.com/parodos-dev/orchestrator-operator/internal/controller/util"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	rhdh "redhat-developer/red-hat-developer-hub-operator/api/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

const (
	BackstageOperatorGroup               = "rhdh-operator-group"
	BackstageAPIVersion                  = "rhdh.redhat.com/v1alpha1"
	BackstageKind                        = "Backstage"
	BackstageCRName                      = "backstage"
	BackstageReplica               int32 = 1
	RegistrySecretName                   = "dynamic-plugins-npmrc"
	AppConfigRHDHName                    = "app-config-rhdh"
	AppConfigRHDHAuthName                = "app-config-rhdh-auth"
	AppConfigRHDHCatalogName             = "app-config-rhdh-catalog"
	AppConfigRHDHDynamicPluginName       = "dynamic-plugins-rhdh"
)

var ConfigMapNameAndConfigDataKey = map[string]string{
	AppConfigRHDHName:              "app-config-rhdh.yaml",
	AppConfigRHDHAuthName:          "app-config-auth.gh.yaml",
	AppConfigRHDHCatalogName:       "app-config-catalog.yaml",
	AppConfigRHDHDynamicPluginName: "dynamic-plugins.yaml",
}

func CreateBSSecret(secretName string, secretNamespace, npmRegistry string,
	ctx context.Context, client client.Client) error {
	logger := log.FromContext(ctx)
	logger.Info("Creating Backstage NPMrc Secret")

	secret := &corev1.Secret{}
	err := client.Get(ctx, types.NamespacedName{
		Namespace: secretNamespace,
		Name:      secretName,
	}, secret)

	if err != nil {
		if apierrors.IsNotFound(err) {
			logger.Info("Secret does not exist. Creating secret", "Secret", secretName)
			newSecret := &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      secretName,
					Namespace: secretNamespace,
					Labels:    operations.AddLabel(),
				},
				Type: corev1.SecretTypeOpaque,
				StringData: map[string]string{
					".npmrc": fmt.Sprintf("registry=%s", npmRegistry),
				},
			}

			if err := client.Create(ctx, newSecret); err != nil {
				logger.Error(err, "Error occurred when creating secret", "Secret", secretName)
				return err
			}
			logger.Info("Successfully created secret", "Secret", secretName)
			return nil
		}
		logger.Error(err, "Error occurred when checking secret exist", "Secret", secretName)
		return err
	}
	logger.Info("Secret already exist", "Secret", secretName)
	return nil
}

func HandleCRCreation(
	operator orchestratorv1alpha1.RHDHOperator,
	pluginsDetails orchestratorv1alpha1.RHDHPlugins,
	clusterDomain string,
	ctx context.Context, client client.Client) error {
	bsLogger := log.FromContext(ctx)

	bsLogger.Info("Handling Backstage resources")

	bsConfigMapList := GetConfigmapList(ctx, client, clusterDomain, operator, pluginsDetails)

	if err := client.Get(ctx, types.NamespacedName{
		Namespace: operator.Subscription.TargetNamespace,
		Name:      BackstageCRName,
	}, &rhdh.Backstage{}); apierrors.IsNotFound(err) {
		secret := rhdh.ObjectKeyRef{
			Name: operator.SecretRef.Name,
		}
		backstageCR := &rhdh.Backstage{
			TypeMeta: metav1.TypeMeta{
				APIVersion: BackstageAPIVersion,
				Kind:       BackstageKind,
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      BackstageCRName,
				Namespace: operator.Subscription.TargetNamespace,
				Labels:    operations.AddLabel(),
			},
			Spec: rhdh.BackstageSpec{
				Application: &rhdh.Application{
					AppConfig:                   &rhdh.AppConfig{ConfigMaps: bsConfigMapList},
					DynamicPluginsConfigMapName: AppConfigRHDHDynamicPluginName,
					ExtraEnvs: &rhdh.ExtraEnvs{
						Secrets: []rhdh.ObjectKeyRef{secret},
					},
					Replicas: util.MakePointer(BackstageReplica),
				},
			},
		}
		if err := client.Create(ctx, backstageCR); err != nil {
			bsLogger.Error(err, "Error occurred when creating Backstage resource")
			return err
		}
		bsLogger.Info("Successfully created Backstage resource")
	}
	return nil
}

func GetConfigmapList(ctx context.Context, client client.Client, clusterDomain string,
	operator orchestratorv1alpha1.RHDHOperator,
	rhdhPlugins orchestratorv1alpha1.RHDHPlugins) []rhdh.ObjectKeyRef {

	cmLogger := log.FromContext(ctx)
	cmLogger.Info("Creating configmaps")

	configmapList := make([]rhdh.ObjectKeyRef, 0)
	namespace := operator.Subscription.TargetNamespace
	for cmName, configDataKey := range ConfigMapNameAndConfigDataKey {
		if err := client.Get(ctx, types.NamespacedName{
			Namespace: namespace,
			Name:      cmName,
		}, &corev1.ConfigMap{}); apierrors.IsNotFound(err) {
			configValue, err := ConfigMapTemplateFactory(cmName, clusterDomain, operator, rhdhPlugins)
			if err != nil {
				cmLogger.Error(err, "Error occurred when parsing config data for configmap", "CM", cmName)
				continue
			} else {
				if err := CreateConfigMap(cmName, configDataKey, namespace, configValue, ctx, client); err == nil {
					if cmName != AppConfigRHDHDynamicPluginName {
						configmapList = append(configmapList, rhdh.ObjectKeyRef{Name: cmName})
					}
				}
			}
		}

	}
	return configmapList
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
			Labels:    operations.AddLabel(),
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

func HandleBackstageCleanup(ctx context.Context, client client.Client, olmClientSet olmclientset.Clientset) error {
	logger := log.FromContext(ctx)

	rhdhNamespace := "rhdh-operator" // remove hardcoded TODO
	subscriptionName := "rhdh"       // remove hardcoded TODO

	namespaceExist, _ := operations.CheckNamespaceExist(ctx, client, rhdhNamespace)
	if namespaceExist {
		backstageCRList, err := listBackstageCRs(ctx, client, rhdhNamespace)

		if err != nil || len(backstageCRList) == 0 {
			logger.Error(err, "Failed to list backstage CRs or have no Backstage CRs created by Orchestrator Operator and cannot perform clean up process")
			return err
		}
		if len(backstageCRList) == 1 {
			// remove namespace
			if err := operations.CleanUpNamespace(ctx, rhdhNamespace, client); err != nil {
				logger.Error(err, "Error occurred when deleting namespace", "NS", "namespace")
				return err
			}
			// remove subscription and csv
			if err := operations.CleanUpSubscriptionAndCSV(ctx, olmClientSet, subscriptionName, rhdhNamespace); err != nil {
				logger.Error(err, "Error occurred when deleting Subscription and CSV", "Subscription", subscriptionName)
				return err
			}
			// remove all CRDs, optional (ensure all CRs and namespace have been removed first)
		}
	}
	return nil
}

func listBackstageCRs(ctx context.Context, k8client client.Client, namespace string) ([]rhdh.Backstage, error) {
	logger := log.FromContext(ctx)

	crList := &rhdh.BackstageList{}

	listOptions := []client.ListOption{
		client.InNamespace(namespace),
		client.MatchingLabels{operations.CreatedByLabelKey: operations.CreatedByLabelValue},
	}

	// List the CRs
	if err := k8client.List(ctx, crList, listOptions...); err != nil {
		logger.Error(err, "Error occurred when listing Backstage CRs")
		return nil, err
	}

	logger.Info("Successfully listed Backstage CRs", "Total", len(crList.Items))
	return crList.Items, nil
}
