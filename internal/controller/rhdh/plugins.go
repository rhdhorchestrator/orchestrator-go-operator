package rhdh

type Plugin struct {
	Package   string
	Integrity string
}

const Orchestrator string = "orchestrator"
const OrchestratorBackend string = "orchestratorBackend"
const ScaffolderBackendOrchestrator string = "scaffolderBackendOrchestrator"
const OrchestratorFormWidgets string = "orchestratorFormWidgets"

func getPlugins() map[string]Plugin {
	return map[string]Plugin{
		Orchestrator: {
			Package:   "backstage-plugin-orchestrator-1.6.0-rc.10.tgz",
			Integrity: "sha512-JZQVm6dDtG4NAMGzP/7N2S2ktkWx4Z4bf+WEkMAGaOa6rVNiaX2gYU9hrcbgk4RJuYLQG9ziNKBgxcuS4fbDcQ==",
		},
		OrchestratorBackend: {
			Package:   "backstage-plugin-orchestrator-backend-dynamic-1.6.0-rc.10.tgz",
			Integrity: "sha512-8MYLBHfb7PgZHUx+5/0Vp+O7fCCfnfCw6Q9SF2+WXdY4vyedQpj8L08ST6qwL+yEAIaz92P/2KTrV+ZHTnaFGw==",
		},
		ScaffolderBackendOrchestrator: {
			Package:   "backstage-plugin-scaffolder-backend-module-orchestrator-dynamic-1.6.0-rc.10.tgz",
			Integrity: "sha512-YXWTWRcH1gvp+9PToEHimqpGTU1HMHjqOrsUKblgKgXp443xtFoglLGQrJeB3rAiAC2LZ++gbLKZK1wmBA3jOg==",
		},
		OrchestratorFormWidgets: {
			Package:   "backstage-plugin-orchestrator-form-widgets-1.6.0-rc.10.tgz",
			Integrity: "sha512-OxexajNyT9nMG5x+jswq9GKA/FqCUsbhkHLaV530qH5OJV3naXQ6kJVGQT0nJVih60/rk4yNG3s7afMvBtqW0g==",
		},
	}

}
