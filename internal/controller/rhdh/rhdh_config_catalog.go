package rhdh

const RHDHCatalogTempl = `catalog:
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
      target: https://github.com/rhdhorchestrator/orchestrator-go-operator/blob/main/docs/resources/users.yaml
    {{- end }}
    - type: url
      target: https://github.com/rhdhorchestrator/workflow-software-templates/blob/{{ .CatalogBranch }}/entities/workflow-resources.yaml
    - type: url
      target: https://github.com/rhdhorchestrator/workflow-software-templates/blob/{{ .CatalogBranch }}/scaffolder-templates/github-workflows/basic-workflow/template.yaml
    - type: url
      target: https://github.com/rhdhorchestrator/workflow-software-templates/blob/{{ .CatalogBranch }}/scaffolder-templates/github-workflows/advanced-workflow/template.yaml
    - type: url
      target: https://github.com/redhat-developer/rhdh-plugins/blob/main/workspaces/orchestrator/plugins/orchestrator-common/src/generated/docs/api-doc/orchestrator-api.yaml
    - type: url
      target: https://github.com/rhdhorchestrator/workflow-software-templates/blob/{{ .CatalogBranch }}/scaffolder-templates/gitlab-workflows/basic-workflow/template.yaml
    - type: url
      target: https://github.com/rhdhorchestrator/workflow-software-templates/blob/{{ .CatalogBranch }}/scaffolder-templates/gitlab-workflows/advanced-workflow/template.yaml
    - type: url
      target: https://github.com/rhdhorchestrator/workflow-software-templates/blob/{{ .CatalogBranch }}/scaffolder-templates/gitlab-workflows/convert-workflow-to-template/template.yaml
    - type: url
      target: https://github.com/rhdhorchestrator/workflow-software-templates/blob/{{ .CatalogBranch }}/scaffolder-templates/github-workflows/convert-workflow-to-template/template.yaml
`

type RHDHConfigCatalog struct {
	EnableGuestProvider bool
	CatalogBranch       string
}
