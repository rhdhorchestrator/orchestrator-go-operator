package rhdh

type ContainerEnvVar struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type InitContainer struct {
	Name string            `json:"name"`
	Env  []ContainerEnvVar `json:"env"`
}

type PatchSpec struct {
	Spec struct {
		Template struct {
			Spec struct {
				InitContainers []InitContainer `json:"initContainers"`
			} `json:"spec"`
		} `json:"template"`
	} `json:"spec"`
}
