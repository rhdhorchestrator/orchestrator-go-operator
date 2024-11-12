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
			Package:   "backstage-plugin-orchestrator@1.2.0",
			Integrity: "sha512-FhM13wVXjjF39syowc4RnMC/gKm4TRlmh8lBrMwPXAw1VzgIADI8H6WVEs837poVX/tYSqj2WhehwzFqU6PuhA==",
		},
		OrchestratorBackend: {
			Package:   "backstage-plugin-orchestrator-backend-dynamic@1.2.0",
			Integrity: "sha512-lyw7IHuXsakTa5Pok8S2GK0imqrmXe3z+TcL7eB2sJYFqQPkCP5la1vqteL9/1EaI5eI6nKZ60WVRkPEldKBTg==",
		},
		Notification: {
			Package:   "plugin-notifications-dynamic@1.2.0",
			Integrity: "sha512-1mhUl14v+x0Ta1o8Sp4KBa02izGXHd+wsiCVsDP/th6yWDFJsfSMf/DyMIn1Uhat1rQgVFRUMg8QgrvbgZCR/w==",
		},
		NotificationBackend: {
			Package:   "plugin-notifications-backend-dynamic@1.2.0",
			Integrity: "sha512-pCFB/jZIG/Ip1wp67G0ZDJPp63E+aw66TX1rPiuSAbGSn+Mcnl8g+XlHLOMMTz+NPloHwj2/Tp4fSf59w/IOSw==",
		},
		Signals: {
			Package:   "plugin-signals-dynamic@1.2.0",
			Integrity: "sha512-5tbZyRob0JDdrI97HXb7JqFIzNho1l7JuIkob66J+ZMAPCit+pjN1CUuPbpcglKyyIzULxq63jMBWONxcqNSXw==",
		},
		SignalsBackend: {
			Package:   "plugin-signals-backend-dynamic@1.2.0",
			Integrity: "sha512-DIISzxtjeJ4a9mX3TLcuGcavRHbCtQ5b52wHn+9+uENUL2IDbFoqmB4/9BQASaKIUSFkRKLYpc5doIkrnTVyrA==",
		},
		NotificationsEmail: {
			Package:   "plugin-notifications-backend-module-email-dynamic@1.2.0",
			Integrity: "sha512-dtmliahV5+xtqvwdxP2jvyzd5oXTbv6lvS3c9nR8suqxTullxxj0GFg1uU2SQ2uKBQWhOz8YhSmrRwxxLa9Zqg==",
		},
	}

}
