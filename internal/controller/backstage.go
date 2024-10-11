package controller

import (
	"context"
	"fmt"
	orchestratorv1alpha1 "github.com/parodos-dev/orchestrator-operator/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

var ConfigMapNames = []string{
	AppConfigRHDHName,
	AppConfigRHDHAuthName,
	AppConfigRHDHCatalogName,
	AppConfigRHDHDynamicPluginName,
}

func createBSSecret(
	secretName string, secretNamespace,
	npmRegistry string,
	ctx context.Context, client client.Client) error {

	logger := log.FromContext(ctx)

	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretName,
			Namespace: secretNamespace,
		},
		Type: corev1.SecretTypeOpaque,
		StringData: map[string]string{
			".npmrc": fmt.Sprintf("registry=%s", npmRegistry),
		},
	}

	if err := client.Create(ctx, secret); err != nil {
		logger.Error(err, "Error occurred when creating secret", "Secret", secretName)
		return err
	}
	logger.Info("Successfully created secret", "Secret", secretName)
	return nil
}

func handleCRCreation(
	operator orchestratorv1alpha1.RHDHOperator,
	ctx context.Context, client client.Client) error {
	bsLogger := log.FromContext(ctx)

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
				AppConfig:                   &rhdh.AppConfig{ConfigMaps: getConfigmapList(ctx, client, operator)},
				DynamicPluginsConfigMapName: AppConfigRHDHDynamicPluginName,
				ExtraEnvs: &rhdh.ExtraEnvs{
					Secrets: []rhdh.ObjectKeyRef{secret},
				},
				Replicas: makePointer(BackstageReplica),
			},
		},
	}
	if err := client.Create(ctx, backstageCR); err != nil {
		bsLogger.Error(err, "Error occurred when creating Backstage rescource")
		return err
	}
	bsLogger.Info("Successfully created Backstage resource")
	return nil
}

func getConfigmapList(ctx context.Context, client client.Client, operator orchestratorv1alpha1.RHDHOperator) []rhdh.ObjectKeyRef {
	configmapList := make([]rhdh.ObjectKeyRef, 0)
	for _, configKey := range ConfigMapNames {
		configValue := ConfigMapTemplateFactory(configKey, operator)
		if err := createConfigMap(configKey, operator.Subscription.TargetNamespace, configValue, ctx, client); err == nil {
			configmapList = append(configmapList, rhdh.ObjectKeyRef{
				Name: "name",
				Key:  configKey,
			})
		}
	}
	return configmapList
}

func createConfigMap(
	name string, namespace string, configValue string,
	ctx context.Context, client client.Client) error {

	logger := log.FromContext(ctx)
	// Create the ConfigMap object
	configMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Data: map[string]string{
			fmt.Sprintf("%s.yaml", name): configValue,
		},
	}
	if err := client.Create(ctx, configMap); err != nil {
		logger.Error(err, "Error occurred when creating ConfigMap", "CM", name)
		return err
	}
	logger.Info("Successfully created ConfigMap", "CM", name)
	return nil
}

func makePointer[T any](t T) *T {
	return &t
}
