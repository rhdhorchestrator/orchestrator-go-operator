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
		Notification: {
			Package:   "plugin-notifications-dynamic@1.3.0",
			Integrity: "sha512-iYLgIy0YdP/CdTLol07Fncmo9n0J8PdIZseiwAyUt9RFJzKIXmoi2CpQLPKMx36lEgPYUlT0rFO81Ie2CSis4Q==",
		},
		NotificationBackend: {
			Package:   "plugin-notifications-backend-dynamic@1.3.0",
			Integrity: "sha512-Pw9Op/Q+1MctmLiVvQ3M+89tkbWkw8Lw0VfcwyGSMiHpK/Xql1TrSFtThtLlymRgeCSBgxHYhh3MUusNQX08VA==",
		},
		Signals: {
			Package:   "plugin-signals-dynamic@1.3.0",
			Integrity: "sha512-+E8XeTXcG5oy+aNImGj/MY0dvEkP7XAsu4xuZjmAqOHyVfiIi0jnP/QDz8XMbD1IjCimbr/DMUZdjmzQiD0hSQ==",
		},
		SignalsBackend: {
			Package:   "plugin-signals-backend-dynamic@1.3.0",
			Integrity: "sha512-5Bl6C+idPXtquQxMZW+bjRMcOfFYcKxcGZZFv2ITkPVeY2zzxQnAz3vYHnbvKRSwlQxjIyRXY6YgITGHXWT0nw==",
		},
		NotificationsEmail: {
			Package:   "plugin-notifications-backend-module-email-dynamic@1.3.0",
			Integrity: "sha512-sm7yRoO6Nkk3B7+AWKb10maIrb2YBNSiqQaWmFDVg2G9cbDoWr9wigqqeQ32+b6o2FenfNWg8xKY6PPyZGh8BA==",
		},
	}

}
