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
			Package:   "backstage-plugin-orchestrator-1.6.0-rc.13.tgz",
			Integrity: "sha512-8hG2rviqBzzEVRhbHdyZxRZBPLV2fbsDhmTqZSbB6Yt9fDvKNZ/WIaoDGcMmv4cUcmfI3ozTdd5935/1RJG/nA==",
		},
		OrchestratorBackend: {
			Package:   "backstage-plugin-orchestrator-backend-dynamic-1.6.0-rc.13.tgz",
			Integrity: "sha512-bnfDpTa3snJMByumIGOSt0iuT0kQEAA7w9lLY1SrLm++3maWFUw6zAe70DQsT7qv1d+gx6pXCdmyTAxk8L+4dw==",
		},
		ScaffolderBackendOrchestrator: {
			Package:   "backstage-plugin-scaffolder-backend-module-orchestrator-dynamic-1.6.0-rc.13.tgz",
			Integrity: "sha512-2ocAxZYxsymLzVNe2ebiD1b9/TxJZyh7TYDcM6PeC1G416Mqf8QJZMuHh44hfBpdDkRRK8qRrVEDOAfDC2J5sA==",
		},
		OrchestratorFormWidgets: {
			Package:   "backstage-plugin-orchestrator-form-widgets-1.6.0-rc.13.tgz",
			Integrity: "sha512-F49J8XOql9TPlETzqy5ll0XHM4HZiWQbEzM8RsOdMTaMh6FNoTR04LATFDYTi/gQg4PC5qT8W7wSxH7gH/Gj3w==",
		},
	}

}
