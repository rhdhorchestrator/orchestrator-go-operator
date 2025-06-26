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
			Package:   "backstage-plugin-orchestrator@1.6.0",
			Integrity: "sha512-fOSJv2PgtD2urKwBM7p9W6gV/0UIHSf4pkZ9V/wQO0eg0Zi5Mys/CL1ba3nO9x9l84MX11UBZ2r7PPVJPrmOtw==",
		},
		OrchestratorBackend: {
			Package:   "backstage-plugin-orchestrator-backend-dynamic@1.6.0",
			Integrity: "sha512-Kr55YbuVwEADwGef9o9wyimcgHmiwehPeAtVHa9g2RQYoSPEa6BeOlaPzB6W5Ke3M2bN/0j0XXtpLuvrlXQogA==",
		},
		ScaffolderBackendOrchestrator: {
			Package:   "backstage-plugin-scaffolder-backend-module-orchestrator-dynamic@1.6.0",
			Integrity: "sha512-Bueeix4661fXEnfJ9y31Yw91LXJgw6hJUG7lPVdESCi9VwBCjDB9Rm8u2yPqP8sriwr0OMtKtqD+Odn3LOPyVw==",
		},
		OrchestratorFormWidgets: {
			Package:   "backstage-plugin-orchestrator-form-widgets@1.6.0",
			Integrity: "sha512-Tqn6HO21Q1TQ7TFUoRhwBVCtSBzbQYz+OaanzzIB0R24O6YtVx3wR7Chtr5TzC05Vz5GkBO1+FZid8BKpqljgA==",
		},
	}

}
