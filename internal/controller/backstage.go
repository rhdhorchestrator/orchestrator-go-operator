package controller

import (
	"context"
	"fmt"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

const (
	BackstageOperatorGroup = "rhdh-operator-group"
)

func createBSSecret(
	secretName string, secretNamespace, npmRegistry string,
	ctx context.Context, client client.Client) {

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
	}
	logger.Info("Successfully created secret", "Secret", secretName)
}
