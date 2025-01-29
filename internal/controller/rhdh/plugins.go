package rhdh

type Plugin struct {
	Package   string
	Integrity string
}

const Orchestrator string = "orchestrator"
const OrchestratorBackend string = "orchestratorBackend"
const Notification string = "notifications"
const NotificationBackend string = "notificationsBackend"
const Signals string = "signals"
const SignalsBackend string = "signalsBackend"
const NotificationsEmail string = "notificationsEmail"

func getPlugins() map[string]Plugin {
	return map[string]Plugin{
		Orchestrator: {
			Package:   "backstage-plugin-orchestrator@1.4.0-rc.3",
			Integrity: "sha512-vGGd9hUmDriEMmP2TfzLVa3JSnfot2Blg+aftDnu/lEphsY1s2gdA4Z5lCxUk7aobcKE6JO/f0sIl2jyxZ7ktw==",
		},
		OrchestratorBackend: {
			Package:   "backstage-plugin-orchestrator-backend-dynamic@1.4.0-rc.3",
			Integrity: "sha512-tS5cJGwjzP9esdTZvUFjw0O7+w9gGBI/+VvJrtqYJBDGXcEAq9iYixGk67ddQVW5eeUM7Tk1WqJlNj282aAWww==",
		},
	}

}
