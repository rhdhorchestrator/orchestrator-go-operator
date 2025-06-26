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
			Package:   "backstage-plugin-orchestrator-1.6.0-rc.14.tgz",
			Integrity: "sha512-9gmptRRqrx0cZjThctJJLYCuzPa1av5S9NdAitAFsGvx5KVgVK8VinOKpyHgLbM6cdBDdIne0wdVFaLijxQHjg==",
		},
		OrchestratorBackend: {
			Package:   "backstage-plugin-orchestrator-backend-dynamic-1.6.0-rc.14.tgz",
			Integrity: "sha512-kuN16JcbbPSvBdr0iJRUlwMaXppTir+edpsYQerXzHsZhaU9cEzXEI5tAsUklOb5qSAi9ENeCZTK0jTlp8wUlg==",
		},
		ScaffolderBackendOrchestrator: {
			Package:   "backstage-plugin-scaffolder-backend-module-orchestrator-dynamic-1.6.0-rc.14.tgz",
			Integrity: "sha512-HpP6WFu8xTc2HTg2QLLRCW59nyt4MhFhlww477BUcBwiHzohJipG30b3eT/OcSMaFIMgOaD2dCMlLfxWLab9Qg==",
		},
		OrchestratorFormWidgets: {
			Package:   "backstage-plugin-orchestrator-form-widgets-1.6.0-rc.14.tgz",
			Integrity: "sha512-ZQwbHD7wWQ9ElOetThPLBwbcmpNR8A+S+XLtG9Q2c7EycxA+mgXTtq8GQV1iteavHhCmcmpCriF3lsdLax61+g==",
		},
	}

}
