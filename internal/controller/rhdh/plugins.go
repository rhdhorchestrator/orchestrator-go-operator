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
			Package:   "backstage-plugin-orchestrator-1.6.0-rc.7.tgz",
			Integrity: "sha512-tT7IVjCMxmVvpKG1yClC/W2y1/ObHvACLYmR+W0MLMuSB5Jnsdj1OmCd0gGbdpmaUySpNi7vc7mJ1alJ8/JvHw==",
		},
		OrchestratorBackend: {
			Package:   "backstage-plugin-orchestrator-backend-dynamic-1.6.0-rc.7.tgz",
			Integrity: "sha512-mW6stwzp/Nl4aU9kkzG7XsQrFwRtUGdMN2qMv86Vo7ketyG+WzQ7g5v36bbGS/1rNLZa8R0+ksulMqOON2J3JQ==",
		},
		ScaffolderBackendOrchestrator: {
			Package:   "backstage-plugin-scaffolder-backend-module-orchestrator-dynamic-1.6.0-rc.7.tgz",
			Integrity: "sha512-OtPqBNtuPJ35gjagRG6DDplrjwQYQerkYJA8cA7zjdzeJGBJtEGReG4EEjzaXj4sOyvQ6lXhIUFbYYs1Qnni/A==",
		},
		OrchestratorFormWidgets: {
			Package: "backstage-plugin-orchestrator-form-widgets-1.6.0-rc.7.tgz",
			Integrity: "sha512-VWX/taVAqFTvpBDujPbtUB6VLPbg3Lxhf0GI43yt5Jlm0zyxR54h1yOAwRmuxRD41X0txPxawN4uxE3WSpOYrQ==",
		},
	}

}
