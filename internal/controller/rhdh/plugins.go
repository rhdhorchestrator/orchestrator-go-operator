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
			Package:   "backstage-plugin-orchestrator-1.6.0-rc.3.tgz",
			Integrity: "sha512-b0Px4lYGVgwr0pd3VFg6bFt26B8Mkv/HYfTDlhqC3jFRdb8WZGYOSIbSjCegpE12uRxBn+P3e+4qqqg2NS3lMQ==",
		},
		OrchestratorBackend: {
			Package:   "backstage-plugin-orchestrator-backend-dynamic-1.6.0-rc.3.tgz",
			Integrity: "sha512-ghGboDXc24f5jZLUMNkw86l8P+FDPYIvea8OMrcSrCCGRiSazEAgZd7IwzbJ61s0tIY5m5bDd7PHJOfleizXqQ==",
		},
		ScaffolderBackendOrchestrator: {
			Package:   "backstage-plugin-scaffolder-backend-module-orchestrator-dynamic-1.6.0-rc.3.tgz",
			Integrity: "sha512-L94IksLT0BF0YRB1BQ+IAEoG0NkCNojy2tQtD6e39JgsbC/Ht9mytNLxWRAa/+ppV+yz+mFGHDyiqaa1YQaRTA==",
		},
		OrchestratorFormWidgets: {
			Package: "backstage-plugin-orchestrator-form-widgets-1.6.0-rc.3.tgz",
			Integrity: "sha512-86TWZctRwmQC0MPTa2QBxwBgE4k26CN633jDv0F4iaT0TKRfr9fhT4HZEAOyBxmPe/P2QlPj5BQchF7T7YTkzA==",
		},
	}

}
