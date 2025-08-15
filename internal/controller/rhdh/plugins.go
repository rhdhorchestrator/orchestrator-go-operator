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
			Package:   "backstage-plugin-orchestrator@1.6.1",
			Integrity: "sha512-6qQ/TLvrf4+gDhrF5JtKQ51hTrNkhEw0jE4lWvLmhauZKeD0EeJVYOlbAvDJZjmx7iJZXLFFydR6EnYuaHBZ+A==",
		},
		OrchestratorBackend: {
			Package:   "backstage-plugin-orchestrator-backend-dynamic@1.6.1",
			Integrity: "sha512-oAHyLnLWzPMeCuUCc2syuG1bJ+7say7n+AjXu/oEi2t59ULCKI6zFpBSy0GvXd7zoBC9ruW/slhEG+APKmTQUg==",
		},
		ScaffolderBackendOrchestrator: {
			Package:   "backstage-plugin-scaffolder-backend-module-orchestrator-dynamic@1.6.1",
			Integrity: "sha512-FPd9bZZhlnYqPej4gCWR1eXaGOPouticrufd8kvHNwfJcO3eRCzPr5yC9E9tbEqyzvZvQBDfljcBeswORhIqfQ==",
		},
		OrchestratorFormWidgets: {
			Package:   "backstage-plugin-orchestrator-form-widgets@1.6.1",
			Integrity: "sha512-jWuawuAxVo7DDSX26t+L4DPhCxR8cpl3AMvUQnWKejzj2/1GwL/FHfffQwa2sSF2xtOKfkAJwnv5p4/5ocjcaQ==",
		},
	}

}
