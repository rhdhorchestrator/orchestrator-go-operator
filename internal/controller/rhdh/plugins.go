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
			Package:   "backstage-plugin-orchestrator@1.4.0-rc.7.tgz",
			Integrity: "sha512-Vclb+TIL8cEtf9G2nx0UJ+kMJnCGZuYG/Xcw0Otdo/fZGuynnoCaAZ6rHnt4PR6LerekHYWNUbzM3X+AVj5cwg==",
		},
		OrchestratorBackend: {
			Package:   "backstage-plugin-orchestrator-backend-dynamic@1.4.0-rc.7.tgz",
			Integrity: "sha512-bxD0Au2V9BeUMcZBfNYrPSQ161vmZyKwm6Yik5keZZ09tenkc8fNjipwJsWVFQCDcAOOxdBAE0ibgHtddl3NKw==",
		},
	}

}
