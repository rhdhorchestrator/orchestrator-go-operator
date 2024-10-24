package rhdh

import (
	"bytes"
	"fmt"
	"github.com/parodos-dev/orchestrator-operator/api/v1alpha1"
	"text/template"
)

func ConfigMapTemplateFactory(
	cmTemplateType string, clusterDomain string,
	operator v1alpha1.RHDHOperator, plugins v1alpha1.RHDHPlugins) (string, error) {
	switch cmTemplateType {
	case AppConfigRHDHName:
		configData := RHDHConfig{
			TargetNamespace: operator.Subscription.TargetNamespace,
			ArgoCDUsername:  operator.SecretRef.ArgoCD.Username,
			ArgoCDPassword:  operator.SecretRef.ArgoCD.Password,
			ArgoCDUrl:       operator.SecretRef.ArgoCD.Url,
			ArgoCDEnabled:   operator.SecretRef.ArgoCD.Enabled,
			BackendSecret:   operator.SecretRef.Backstage.BackendSecret,
			ClusterDomain:   clusterDomain,
		}
		formattedConfig, err := parseConfigTemplate(RHDHConfigTempl, configData)
		if err != nil {
			return "", err
		}
		return formattedConfig, nil
	case AppConfigRHDHAuthName:
		configData := RHDHConfigAuth{
			GitHubToken:         operator.SecretRef.Github.Token,
			Environment:         "development",
			GitHubClientId:      operator.SecretRef.Github.ClientID,
			GitHubClientSecret:  operator.SecretRef.Github.ClientSecret,
			EnableGuestProvider: operator.EnableGuestProvider,
		}
		formattedConfig, err := parseConfigTemplate(RHDHAuthTempl, configData)
		if err != nil {
			return "", err
		}
		return formattedConfig, nil
	case AppConfigRHDHCatalogName:
		configData := RHDHConfigCatalog{
			EnableGuestProvider: operator.EnableGuestProvider,
			CatalogBranch:       operator.CatalogBranch,
		}
		formattedConfig, err := parseConfigTemplate(RHDHCatalogTempl, configData)
		if err != nil {
			return "", err
		}
		return formattedConfig, nil
	case AppConfigRHDHDynamicPluginName:
		pluginsMap := getPlugins()
		configData := RHDHDynamicPluginConfig{
			K8ClusterToken:               operator.SecretRef.ClusterTokenUrl.ClusterToken,
			K8ClusterUrl:                 operator.SecretRef.ClusterTokenUrl.ClusterUrl,
			TektonEnabled:                false,
			ArgoCDEnabled:                operator.SecretRef.ArgoCD.Enabled,
			ArgoCDUrl:                    operator.SecretRef.ArgoCD.Url,
			ArgoCDUsername:               operator.SecretRef.ArgoCD.Username,
			ArgoCDPassword:               operator.SecretRef.ArgoCD.Password,
			OrchestratorBackendPackage:   pluginsMap[OrchestratorBackend].Package,
			OrchestratorBackendIntegrity: pluginsMap[OrchestratorBackend].Integrity,
			OrchestratorPackage:          pluginsMap[Orchestrator].Package,
			OrchestratorIntegrity:        pluginsMap[Orchestrator].Integrity,
			Scope:                        plugins.Scope,
			NotificationPackage:          pluginsMap[Notification].Package,
			NotificationIntegrity:        pluginsMap[Notification].Integrity,
			SignalsPackage:               pluginsMap[Signals].Package,
			SignalsIntegrity:             pluginsMap[Signals].Integrity,
			SignalsBackendPackage:        pluginsMap[SignalsBackend].Package,
			SignalsBackendIntegrity:      pluginsMap[SignalsBackend].Integrity,
			NotificationBackendPackage:   pluginsMap[NotificationBackend].Package,
			NotificationBackendIntegrity: pluginsMap[NotificationBackend].Integrity,
			NotificationEmailPackage:     pluginsMap[NotificationsEmail].Package,
			NotificationEmailIntegrity:   pluginsMap[NotificationsEmail].Integrity,
			NotificationEmailEnabled:     plugins.NotificationsConfig.Enabled,
			NotificationEmailHostname:    operator.SecretRef.NotificationsEmail.Hostname,
			NotificationEmailUsername:    operator.SecretRef.NotificationsEmail.Username,
			NotificationEmailPassword:    operator.SecretRef.NotificationsEmail.Password,
			NotificationEmailSender:      plugins.NotificationsConfig.Sender,
			NotificationEmailReplyTo:     plugins.NotificationsConfig.Recipient,
			NotificationEmailPort:        plugins.NotificationsConfig.Port,
			WorkflowNamespace:            "sonataflow-infra",
		}
		formattedConfig, err := parseConfigTemplate(RHDHDynamicPluginTempl, configData)
		if err != nil {
			return "", err
		}
		return formattedConfig, nil
	default:
		return "", nil
	}
}

func parseConfigTemplate(templateString string, configData any) (string, error) {
	// parse the template
	templ, err := template.New("config").Parse(templateString)
	if err != nil {
		fmt.Printf("Error occurred when parsing template: %v\n", err)
		return "", err
	}

	// execute template with the dynamic data
	var output bytes.Buffer
	err = templ.Execute(&output, configData)
	if err != nil {
		fmt.Printf("Error occurred when executing template: %v\n", err)
		return "", err
	}
	return output.String(), nil
}
