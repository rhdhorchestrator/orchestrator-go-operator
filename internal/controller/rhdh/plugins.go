package rhdh

type Plugin struct {
	Package   string
	Integrity string
}

const Orchestrator string = "orchestrator"
const OrchestratorBackend string = "orchestratorBackend"

func getPlugins() map[string]Plugin {
	return map[string]Plugin{
		Orchestrator: {
			Package:   "backstage-plugin-orchestrator-1.5.1-rc.1.tgz",
			Integrity: "sha512-Pw1PfiGTLJzjRCINA3+s3zWy6L0WTmCZjOxa/iZlkpTN6ywAcmPD43wtWaFT5Hxhlr5NMa/2af291F4G1gAqcg==",
		},
		OrchestratorBackend: {
			Package:   "backstage-plugin-orchestrator-backend-dynamic-1.5.1-rc.1.tgz",
			Integrity: "sha512-3YJYUFSaxNTQc5k40sfmVdibXFzbYEeiyBrGLOdcG6mB+Vdm90Hbv+isa6y3bbsK22juXbtAjsl39TtVGI1esg==",
		},
	}

}
