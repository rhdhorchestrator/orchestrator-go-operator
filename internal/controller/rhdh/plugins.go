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
			Package:   "backstage-plugin-orchestrator-1.6.0-rc.11.tgz",
			Integrity: "sha512-jZNnWMh4Hjs9cem5uvmt89j3Yded61EVYk/NbzeqPfW9NFP7oOzNwyfQeG59hsNQ/PEiL7XIEzGp0vIHbHL+Cg==",
		},
		OrchestratorBackend: {
			Package:   "backstage-plugin-orchestrator-backend-dynamic-1.6.0-rc.11.tgz",
			Integrity: "sha512-ebw1Uk5xsRzIfMziZ4sZU6UPxfQcHnEqAldAGDxT1vih5t+S/w1LuGlRGhajk3JXeA+RUXA6xfFcr9nNgYGegQ==",
		},
		ScaffolderBackendOrchestrator: {
			Package:   "backstage-plugin-scaffolder-backend-module-orchestrator-dynamic-1.6.0-rc.11.tgz",
			Integrity: "sha512-ufHBFH1maL4w9/pxweFFzK2pGbhDfbJwSfPKmQyCtilJRUR7bLSPKXP2ZYYrR8tysiQ0vB6fw2FtuNQNDqHtqQ==",
		},
		OrchestratorFormWidgets: {
			Package:   "backstage-plugin-orchestrator-form-widgets-1.6.0-rc.11.tgz",
			Integrity: "sha512-cZwpaHA1MoydTxRU3404AqgjzX+IeW1/2jnqzSHn4XwYNAg5xk5udTSq+zVabmimdeeKXWB3iCWi1sP4bLkdKQ==",
		},
	}

}
