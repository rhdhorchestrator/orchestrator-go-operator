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
			Package:   "backstage-plugin-orchestrator-1.6.0-rc.9.tgz",
			Integrity: "sha512-0/Eo9SqRtC9AmWkdJk+nhJSmSDBvKg1eWl0to5rOqsQiWRk57MUEaWRLwjK6fwu9975EJw3XvrTrmgYmFsI0mg==",
		},
		OrchestratorBackend: {
			Package:   "backstage-plugin-orchestrator-backend-dynamic-1.6.0-rc.9.tgz",
			Integrity: "sha512-LQVUYGUSelYDubbwMG5PT9ITYlaghsTCp37ktIsLjC9Qlr2NeA20xAIV4oDhtNVkoRraR6iFmVVKXnD/D2yrLg==",
		},
		ScaffolderBackendOrchestrator: {
			Package:   "backstage-plugin-scaffolder-backend-module-orchestrator-dynamic-1.6.0-rc.9.tgz",
			Integrity: "sha512-4F563LxlAzGakDx4J63szF0i8YyO6ZVRz0i9Bp/Qessdp1E+zlRCgyIqHWSgQGUopzVzNrT20LmHQUzosH0naw==",
		},
		OrchestratorFormWidgets: {
			Package:   "backstage-plugin-orchestrator-form-widgets-1.6.0-rc.9.tgz",
			Integrity: "sha512-O5lwQ4dezu6ueZEHJ3rUXsjBGs8N5zTK540L8nAufx2DYyKApBuEMj4PJStW6rKFuA53HB49+y35wUhS40Fw1g==",
		},
	}

}
