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
			Package:   "backstage-plugin-orchestrator-1.5.0-rc.3.tgz",
			Integrity: "sha512-SkvEkftCmeta/0hBbWFkLAgfkJenL/xn23kpHS4cXlkSaXN6Cn7V/tsLIoYiDuBghQ3jbivNiU0EcNgBkCIK2w==",
		},
		OrchestratorBackend: {
			Package:   "backstage-plugin-orchestrator-backend-dynamic-1.5.0-rc.3.tgz",
			Integrity: "sha512-TTUyOStMrjipF3i7Bzyz4GxAW+g6KBA8x5rlFE2jJjh4gtI8K1w2zNOGRfq0dniSxkZOYdhO15SX6CenFf4UrA==",
		},
		ScaffolderBackendOrchestrator: {
			Package:   "backstage-plugin-scaffolder-backend-module-orchestrator-dynamic-1.5.0-rc.3.tgz",
			Integrity: "sha512-C+iazAp/i+x/iCtlA/l2Vc/AVO4TMfDxwn0hrXVEVbO9FHdaKt63EdLQ2dvtWslEnlhv/9eDAGp0L8Ct6lbRZA==",
		},
	}

}
