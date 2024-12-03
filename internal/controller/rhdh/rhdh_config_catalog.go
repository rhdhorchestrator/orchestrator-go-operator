package rhdh

const RHDHCatalogTempl = `
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
      target: https://github.com/parodos-dev/workflow-software-templates/blob/{{ .CatalogBranch }}/entities/workflow-resources.yaml
    - type: url
      target: https://github.com/parodos-dev/workflow-software-templates/blob/{{ .CatalogBranch }}/scaffolder-templates/basic-workflow/template.yaml
    - type: url
      target: https://github.com/parodos-dev/workflow-software-templates/blob/{{ .CatalogBranch }}/scaffolder-templates/complex-assessment-workflow/template.yaml
    - type: url
      target: https://github.com/redhat-developer/rhdh-plugins/blob/main/workspaces/orchestrator/plugins/orchestrator-common/src/generated/docs/api-doc/orchestrator-api.yaml
`

type RHDHConfigCatalog struct {
	EnableGuestProvider bool
	CatalogBranch       string
}
