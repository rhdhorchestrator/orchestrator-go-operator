package rhdh

type Plugin struct {
	Package   string
	Integrity string
}

const Orchestrator string = "orchestrator"
const OrchestratorBackend string = "orchestratorBackend"
const ScaffolderBackendOrchestrator string = "scaffolderBackendOrchestrator"

func getPlugins() map[string]Plugin {
	return map[string]Plugin{
		Orchestrator: {
			Package:   "backstage-plugin-orchestrator@1.4.0-rc.7",
			Integrity: "sha512-Vclb+TIL8cEtf9G2nx0UJ+kMJnCGZuYG/Xcw0Otdo/fZGuynnoCaAZ6rHnt4PR6LerekHYWNUbzM3X+AVj5cwg==",
		},
		OrchestratorBackend: {
			Package:   "backstage-plugin-orchestrator-backend-dynamic@1.4.0-rc.7",
			Integrity: "sha512-bxD0Au2V9BeUMcZBfNYrPSQ161vmZyKwm6Yik5keZZ09tenkc8fNjipwJsWVFQCDcAOOxdBAE0ibgHtddl3NKw==",
		},
		ScaffolderBackendOrchestrator: {
			Package:   "backstage-plugin-scaffolder-backend-module-orchestrator-dynamic-1.5.0-rc.1.tgz",
			Integrity: "sha512-nvVU4TnWttq5OM9/4e0dOIDMsa4q2MH2G1LsnpaGnNKbYWmJXlhLBwy/4fOOhkf+Y+b+FIlMjNgiZoaM6HUsQA==",
		},
	}

}
