package rhdh

type RHDHDynamicPluginConfig struct {
	K8ClusterToken               string
	K8ClusterUrl                 string
	TektonEnabled                bool
	ArgoCDEnabled                bool
	ArgoCDUrl                    string
	ArgoCDUsername               string
	ArgoCDPassword               string
	OrchestratorBackendPackage   string
	OrchestratorBackendIntegrity string
	OrchestratorPackage          string
	OrchestratorIntegrity        string
	Scope                        string
	NotificationPackage          string
	NotificationIntegrity        string
	SignalsPackage               string
	SignalsIntegrity             string
	SignalsBackendPackage        string
	SignalsBackendIntegrity      string
	NotificationBackendPackage   string
	NotificationBackendIntegrity string
	NotificationEmailPackage     string
	NotificationEmailIntegrity   string
	NotificationEmailEnabled     bool
	NotificationEmailHostname    string
	NotificationEmailUsername    string
	NotificationEmailPassword    string
	NotificationEmailSender      string
	NotificationEmailReplyTo     string
	NotificationEmailPort        int
	WorkflowNamespace            string
}

const RHDHDynamicPluginTempl = `
includes: 
  - dynamic-plugins.default.yaml
plugins:
  {{- if and (.K8ClusterToken) (.K8ClusterUrl) }}
  - package: ./dynamic-plugins/dist/backstage-plugin-kubernetes-backend-dynamic
	disabled: false
	pluginConfig:
	  kubernetes:
		customResources:
		  - group: 'tekton.dev'
			apiVersion: 'v1'
			plural: 'pipelines'
		  - group: 'tekton.dev'
			apiVersion: 'v1'
			plural: 'pipelineruns'
		  - group: 'tekton.dev'
			apiVersion: 'v1'
			plural: 'taskruns'
		  - group: 'route.openshift.io'
			apiVersion: 'v1'
			plural: 'routes'
		serviceLocatorMethod:
		  type: 'multiTenant'
		clusterLocatorMethods:
		  - type: 'config'
			clusters:
			  - name: 'Default Cluster'
				url: {{ printf "${%s}" .K8ClusterUrl }}
				authProvider: 'serviceAccount'
				skipTLSVerify: true
				serviceAccountToken: {{ printf "${%s}" .K8ClusterToken }}
  - package: ./dynamic-plugins/dist/backstage-plugin-kubernetes
	disabled: false
  {{- if .TektonEnabled }}
  - package: ./dynamic-plugins/dist/janus-idp-backstage-plugin-tekton
	disabled: false
  {{- end }}
  {{- end }}
  {{- if and (.ArgoCDEnabled) (.ArgoCDUrl) (.ArgoCDUsername) }}
  - disabled: false
	package: ./dynamic-plugins/dist/janus-idp-backstage-plugin-argocd
  - disabled: false
	package: ./dynamic-plugins/dist/roadiehq-backstage-plugin-argo-cd-backend-dynamic
  - disabled: false
	package: ./dynamic-plugins/dist/roadiehq-scaffolder-backend-argocd-dynamic
  {{- end }}
  - disabled: false
	package: "{{ .Scope }}/{{ .OrchestratorBackendPackage }}"
	integrity: {{ .OrchestratorBackendIntegrity }}
	pluginConfig:
	  orchestrator:
		dataIndexService:
		  url: http://sonataflow-platform-data-index-service.{{ .WorkflowNamespace }}
  - disabled: false
	package: "{{ .Scope }}/{{ .OrchestratorPackage }}"
	integrity: {{ .OrchestratorIntegrity }}
	pluginConfig:
	  dynamicPlugins:
		frontend:
		  janus-idp.backstage-plugin-orchestrator:
			appIcons:
			  - importName: OrchestratorIcon
				module: OrchestratorPlugin
				name: orchestratorIcon
			dynamicRoutes:
			  - importName: OrchestratorPage
				menuItem:
				  icon: orchestratorIcon
				  text: Orchestrator
				module: OrchestratorPlugin
				path: /orchestrator
  - disabled: false
	package: "{{ .Scope }}/{{ .NotificationPackage }}"
	integrity: {{ .NotificationIntegrity }}
	pluginConfig:
	  dynamicPlugins:
		frontend:
		  redhat.plugin-notifications:
			dynamicRoutes:
			  - importName: NotificationsPage
				menuItem:
				  config:
					props:
					  titleCounterEnabled: true
					  webNotificationsEnabled: false
				  importName: NotificationsSidebarItem
				path: /notifications
  - disabled: false
	package: "{{ .Scope }}/{{ .SignalsPackage }}"
	integrity: {{ .SignalsIntegrity }}
	pluginConfig:
	  dynamicPlugins:
		frontend:
		  redhat.plugin-signals: {}
  - disabled: false
	package: "{{ .Scope }}/{{ .NotificationBackendPackage }}"
	integrity: {{ .NotificationBackendIntegrity }}
  - disabled: false
	package: "{{ .Scope }}/{{ .SignalsBackendPackage }}"
	integrity: {{ .SignalsBackendIntegrity }}
  {{- if and (.NotificationEmailEnabled) (.NotificationEmailHostname) }}
  - disabled: false
	package: "{{ .Scope }}/{{ .NotificationEmailPackage }}"
	integrity: {{ .NotificationEmailIntegrity}}
	pluginConfig:
	  notifications:
		processors:
		  email:
			transportConfig:
			  transport: smtp
			  hostname: {{ printf "${%s}" .NotificationEmailHostname }}
			  port: {{ .NotificationEmailPort }}
			  secure: false
			{{- if .NotificationEmailUsername }}
			  username: {{ printf "${%s}" .NotificationEmailUsername }}
			{{- end}}
			{{- if .NotificationEmailPassword }}
			  password: {{ .NotificationEmailPassword }}
			{{- end}}
			sender: {{ .NotificationEmailSender }}
			{{- if .NotificationEmailReplyTo }}
			replyTo: {{ .NotificationEmailReplyTo }}
			{{- end}}
			broadcastConfig:
			  receiver: "none"
			concurrencyLimit: 10
			cache:
			  ttl:
				days: 1
  {{- end }}
`
