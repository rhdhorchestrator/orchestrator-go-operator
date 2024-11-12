package rhdh

import (
	"context"
	"fmt"
	orchestratorv1alpha1 "github.com/parodos-dev/orchestrator-operator/api/v1alpha1"
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
