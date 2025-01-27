package rhdh

const RHDHAuthTempl = `
integrations:
  {{- if and ( .GitHubToken) }}
  github:
    - host: github.com
      token: {{ printf "${%s}" .GitHubToken }}
  {{- end }}
  {{- if and ( .GitLabToken) ( .GitLabHost) }}
  gitlab:
    - host: {{ printf "${%s}" .GitLabHost }}
      token: {{ printf "${%s}" .GitLabToken }}
      apiBaseUrl: https://{{ printf "${%s}" .GitLabHost }}/api/v4
  {{- end }}

{{- if and ( .GitHubToken) }}
auth:
  environment: {{ .Environment }}
{{- end }}
  providers:
    {{- if and ( .GitHubClientId) ( .GitHubClientSecret) }}
    github:
      development:
        clientId: {{ printf "${%s}" .GitHubClientId }}
        clientSecret: {{ printf "${%s}" .GitHubClientSecret }}
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
	GitLabHost          string
	GitLabToken         string
	EnableGuestProvider bool
}
