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
			Package:   "backstage-plugin-orchestrator@1.3.0",
			Integrity: "sha512-A/twx1SOOGDQjglLzOxQikKO0XOdPP1jh2lj9Y/92bLox8mT+eaZpub8YLwR2mb7LsUIUImg+U6VnKwoAV9ATA==",
		},
		OrchestratorBackend: {
			Package:   "backstage-plugin-orchestrator-backend-dynamic@1.3.0",
			Integrity: "sha512-Th5vmwyhHyhURwQo28++PPHTvxGSFScSHPJyofIdE5gTAb87ncyfyBkipSDq7fwj4L8CQTXa4YP6A2EkHW1npg==",
		},
	}

}
