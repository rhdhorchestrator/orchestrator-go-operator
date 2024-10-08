package controller

import (
	"context"
	"fmt"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	rhdh "redhat-developer/red-hat-developer-hub-operator/api/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

const (
	BackstageOperatorGroup       = "rhdh-operator-group"
	BackstageAPIVersion          = "rhdh.redhat.com/v1alpha1"
	BackstageKind                = "Backstage"
	BackstageCRName              = "backstage"
	BackstageReplica       int32 = 1
	RegistrySecretName           = "dynamic-plugins-npmrc"
)

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
	targetNamespace string, secretRefName string,
	ctx context.Context, client client.Client) error {
	bsLogger := log.FromContext(ctx)

	secret := rhdh.ObjectKeyRef{
		Name: "name",
		Key:  secretRefName,
	}
	backstageCR := &rhdh.Backstage{
		TypeMeta: metav1.TypeMeta{
			APIVersion: BackstageAPIVersion,
			Kind:       BackstageKind,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      BackstageCRName,
			Namespace: targetNamespace,
		},
		Spec: rhdh.BackstageSpec{
			Application: &rhdh.Application{
				//AppConfig:                   &rhdh.AppConfig{ConfigMaps: getBSConfigmaps()},
				DynamicPluginsConfigMapName: "dynamic-plugins-rhdh",
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

func getBSConfigmaps() ([]rhdh.ObjectKeyRef, error) {
	return make([]rhdh.ObjectKeyRef, 0), nil
}

func makePointer[T any](t T) *T {
	return &t
}
