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
			Package:   "backstage-plugin-orchestrator-1.5.1.tgz",
			Integrity: "sha512-7VOe+XGTUzrdO/av0DNHbydOjB3Lo+XdCs6fj3JVODLP7Ypd3GXHf/nssYxG5ZYC9F1t9MNeguE2bZOB6ckqTA==",
		},
		OrchestratorBackend: {
			Package:   "backstage-plugin-orchestrator-backend-dynamic-1.5.1.tgz",
			Integrity: "sha512-VIenFStdq9QvvmgmEMG8O7b2wqIebvEcqNeJ9SWZ8jen9t+efTK6D3Rde74LQ1no1QaHLx8RoxNCOuTUEF8O/g==",
		},
		ScaffolderBackendOrchestrator: {
			Package:   "backstage-plugin-scaffolder-backend-module-orchestrator-dynamic-1.5.1.tgz",
			Integrity: "sha512-bnVQjVsUZ470Vgm2kd5Lo/bVa2fF0q4GufBDc/8oTQsnP3zZJQqKFvFElBTCjY76RqkECydlvZ1UFybSzvockQ==",
		},
	}

}
