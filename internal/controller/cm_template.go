package controller

import (
	"bytes"
	"fmt"
	"github.com/parodos-dev/orchestrator-operator/api/v1alpha1"
	"os"
	"text/template"
)

var RHDHAuthTempl = `
    integrations:
      github:
        - host: github.com
          token: {{ .GitHubToken }}
    auth:
      environment: {{ .Environment }}
    providers:
  {{- if .GitHubClientId }}
    github:
      development:
        clientId: {{ .GitHubClientId }}
        clientSecret: {{ .GitHubClientSecret }}
  {{- end }}
  {{- if .EnableGuestProvider }}
    guest:
      dangerouslyAllowOutsideDevelopment: true
      userEntityRef: user:default/guest
  {{- end }}
`

type RHDHConfigAuth struct {
	GitHubToken         string
	Environment         string
	GitHubClientId      string
	GitHubClientSecret  string
	EnableGuestProvider bool
}

var RHDHConfigTempl = `
    app:
      title: Red Hat Developer Hub
      baseUrl: https://backstage-backstage-{{ .TargetNamespace }}.{{ .ClusterDomain }}
    backend:
      auth:
        externalAccess:
          - type: static
            options:
              token: {{ .BackendSecret }}
              subject: orchestrator
      baseUrl: https://backstage-backstage-{{ .TargetNamespace }}.{{ .ClusterDomain }}
      csp:
        script-src: ["'self'", "'unsafe-inline'", "'unsafe-eval'"]
        script-src-elem: ["'self'", "'unsafe-inline'", "'unsafe-eval'"]
        connect-src: ["'self'", 'http:', 'https:', 'data:']
      cors:
        origin: https://backstage-backstage-{{ .TargetNamespace }}.{{ include "cluster.domain" . }}
      database:
        client: pg
        connection:
          password: ${POSTGRESQL_ADMIN_PASSWORD}
          user: ${POSTGRES_USER}
          host: ${POSTGRES_HOST}
          port: ${POSTGRES_PORT}
    {{- if .ArgoCDEnabled }}
    argocd:
      appLocatorMethods:
      - instances:
        - name: main
          url: {{ .ArgoCDUrl }}
          username: {{ .ArgoCDUsername }}
          password: {{ .ArgoCDPassword }}
        type: config
    {{- end }}
`

type RHDHConfig struct {
	TargetNamespace string
	ArgoCDUsername  string
	ArgoCDPassword  string
	ArgoCDUrl       string
	ArgoCDEnabled   bool
	BackendSecret   string
	ClusterDomain   string
}

var RHDHCatalogTempl = `
    catalog:
      rules:
        - allow:
            [
              Component,
              System,
              Group,
              Resource,
              Location,
              Template,
              API,
              User,
              Domain,
            ]
      locations:
      {{- if .EnableGuestProvider }}
        - type: url
          target: https://github.com/parodos-dev/orchestrator-helm-chart/blob/main/resources/users.yaml
      {{- end }}
        - type: url
          target: https://github.com/parodos-dev/workflow-software-templates/blob/{{ .Values.rhdhOperator.catalogBranch }}/entities/workflow-resources.yaml
        - type: url
          target: https://github.com/parodos-dev/workflow-software-templates/blob/{{ .Values.rhdhOperator.catalogBranch }}/scaffolder-templates/basic-workflow/template.yaml
        - type: url
          target: https://github.com/parodos-dev/workflow-software-templates/blob/{{ .Values.rhdhOperator.catalogBranch }}/scaffolder-templates/complex-assessment-workflow/template.yaml
`

func ConfigMapTemplateFactory(cmTemplateType string, operator v1alpha1.RHDHOperator) string {
	switch cmTemplateType {
	case AppConfigRHDHName:
		configData := RHDHConfig{
			TargetNamespace: operator.Subscription.TargetNamespace,
			ArgoCDUsername:  operator.SecretRef.ArgoCD.Username,
			ArgoCDPassword:  operator.SecretRef.ArgoCD.Password,
			ArgoCDUrl:       operator.SecretRef.ArgoCD.Url,
			ArgoCDEnabled:   operator.SecretRef.ArgoCD.Enabled,
			BackendSecret:   operator.SecretRef.Backstage.BackendSecret,
			ClusterDomain:   os.Getenv("CLUSTER_DOMAIN"),
		}
		formattedConfig, _ := parseConfigTemplate(RHDHConfigTempl, configData)
		return formattedConfig
	case AppConfigRHDHAuthName:
		configData := RHDHConfigAuth{
			GitHubToken:         operator.SecretRef.Github.Token,
			Environment:         "development",
			GitHubClientId:      operator.SecretRef.Github.ClientID,
			GitHubClientSecret:  operator.SecretRef.Github.ClientSecret,
			EnableGuestProvider: operator.EnableGuestProvider,
		}
		formattedConfig, _ := parseConfigTemplate(RHDHAuthTempl, configData)
		return formattedConfig
	case AppConfigRHDHCatalogName:
		configData := struct {
			EnableGuestProvider bool
		}{operator.EnableGuestProvider}
		formattedConfig, _ := parseConfigTemplate(RHDHCatalogTempl, configData)
		return formattedConfig
	//case AppConfigRHDHDynamicPluginName:
	//	return nil
	default:
		return ""
	}
}

func parseConfigTemplate(templateString string, configData any) (string, error) {
	// parse the template
	t, err := template.New("config").Parse(templateString)
	if err != nil {
		fmt.Printf("Error occurred when parsing template: %v\n", err)
		return "", err
	}

	// execute template with the dynamic data
	var output bytes.Buffer
	err = t.Execute(&output, configData)
	if err != nil {
		fmt.Printf("Error occurred when executing template: %v\n", err)
		return "", err
	}
	return output.String(), nil
}
