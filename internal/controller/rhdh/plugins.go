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
			Package:   "backstage-plugin-orchestrator-1.6.1-rc.2.tgz",
			Integrity: "sha512-TJ58d5CqFcNmvhBPJp+/7nt0gZo4ILqRjE2+9ZHjIVht2X0gCJqqGYF41sTgBotb2biOD024W/5xp2qQzRbaww==",
		},
		OrchestratorBackend: {
			Package:   "backstage-plugin-orchestrator-backend-dynamic-1.6.1-rc.2.tgz",
			Integrity: "sha512-qveMcu8jO2KsKzgXioNmmbQKxGUbUloWbDxZfa3sQDSGakB6RSE5kNPTAy1QmCvBqufeOFrfv36LpV7d757SHA==",
		},
		ScaffolderBackendOrchestrator: {
			Package:   "backstage-plugin-scaffolder-backend-module-orchestrator-dynamic-1.6.1-rc.2.tgz",
			Integrity: "sha512-fq5oUIVyigshMUHD5N85937wCLIQVixV+mvVmCjl99FvY7A4/5X11vASHOFx+1cLW7zBZDT5hc3zJlPDBR2zWQ==",
		},
		OrchestratorFormWidgets: {
			Package:   "backstage-plugin-orchestrator-form-widgets-1.6.1-rc.2.tgz",
			Integrity: "sha512-1KDZmf+iJUevivLsamiD/wvGhuK9PZeGrPNz5wevFC4eXYHB1Iq+Nugjq1IqBQWKIY3jhAIMwba1ZVz4jlu7/A==",
		},
	}

}
