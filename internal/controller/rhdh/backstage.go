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
	ctx context.Context, client client.Client) error {
	bsLogger := log.FromContext(ctx)

	bsLogger.Info("Handling Backstage resources")

	if err := client.Get(ctx, types.NamespacedName{
		Namespace: operator.Subscription.TargetNamespace,
		Name:      BackstageCRName,
	}, &rhdh.Backstage{}); apierrors.IsNotFound(err) {
		secret := rhdh.ObjectKeyRef{
			Name: "name",
			Key:  operator.SecretRef.Name,
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
					AppConfig:                   &rhdh.AppConfig{ConfigMaps: GetConfigmapList(ctx, client, operator, pluginsDetails)},
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

func GetConfigmapList(ctx context.Context, client client.Client,
	operator orchestratorv1alpha1.RHDHOperator, rhdhPlugins orchestratorv1alpha1.RHDHPlugins) []rhdh.ObjectKeyRef {
	cmLogger := log.FromContext(ctx)
	configmapList := make([]rhdh.ObjectKeyRef, 0)
	cmLogger.Info("Creating configmaps")
	for cmName, configDataKey := range ConfigMapNameAndConfigDataKey {
		configValue, err := ConfigMapTemplateFactory(cmName, operator, rhdhPlugins)
		if err != nil {
			cmLogger.Error(err, "Error occurred when creating configmap", "CM", cmName)
			continue
		} else {
			if err := CreateConfigMap(cmName, configDataKey, operator.Subscription.TargetNamespace, configValue, ctx, client); err == nil {
				configmapList = append(configmapList, rhdh.ObjectKeyRef{
					Name: "name",
					Key:  cmName,
				})
			}
		}
	}
	return configmapList
}

func CreateConfigMap(
	name string, configDataKey string, namespace string, configValue string,
	ctx context.Context, client client.Client) error {

	logger := log.FromContext(ctx)
	if err := client.Get(ctx, types.NamespacedName{
		Namespace: namespace,
		Name:      name,
	}, &corev1.ConfigMap{}); apierrors.IsNotFound(err) {
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
	}
	return nil
}
