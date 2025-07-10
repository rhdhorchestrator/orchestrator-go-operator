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
			Package:   "backstage-plugin-orchestrator-1.6.1-rc.1.tgz",
			Integrity: "sha512-TYFpSbH4qX09Vzm5wyoUoKpjEQ1idej//KXszD8f6jlqduyVj/KndONIhtAxwHtslIopQVNojv7C5oFJs9+AyQ==",
		},
		OrchestratorBackend: {
			Package:   "backstage-plugin-orchestrator-backend-dynamic-1.6.1-rc.1.tgz",
			Integrity: "sha512-ch4Mn+1oGEeQALJ0RY9dfjNj8QQlU0csWm4Vdsr8nQAdW6QLB6A0cuJxhv7Xpum/NzZgVPhOK6BNmb1dHIFr4g==",
		},
		ScaffolderBackendOrchestrator: {
			Package:   "backstage-plugin-scaffolder-backend-module-orchestrator-dynamic-1.6.1-rc.1.tgz",
			Integrity: "sha512-j/81ZK/+sNdBFdrliCX2q7u0HBhhsx2e6ysfcK1/wj1PW3zHeDDB03w/AFvtPbdnEl5Lq0iZtdF9hNHPt6xV/A==",
		},
		OrchestratorFormWidgets: {
			Package:   "backstage-plugin-orchestrator-form-widgets-1.6.1-rc.1.tgz",
			Integrity: "sha512-HasqhJHrY4+fQL9EctC1GQDYkw2mfpL/I//ut5RFBXgNM3+DpCh5DmW8QHAfvzWilfuSFJb3cBOfTrdOoOaDMw==",
		},
	}

}
