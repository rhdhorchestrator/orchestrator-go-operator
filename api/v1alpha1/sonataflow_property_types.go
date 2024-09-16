package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type SonataFlowPlatformSpec struct {
	Build    PlatformBuild    `json:"build,omitempty"`
	Services PlatformServices `json:"services,omitempty"`
}

type PlatformBuild struct {
	Template Template `json:"template,omitempty"`
}

type Template struct {
	Resource Resource `json:"resource,omitempty"`
}

type PlatformServices struct {
	DataIndex  DataIndex  `json:"dataIndex,omitempty"`
	JobService JobService `json:"jobService,omitempty"`
}

type DataIndex struct {
	Enabled     bool        `json:"enabled,omitempty"`
	Persistence Persistence `json:"persistence,omitempty"`
	PodTemplate PodTemplate `json:"podTemplate,omitempty"`
}

type JobService struct {
	Enabled     bool        `json:"enabled,omitempty"`
	Persistence Persistence `json:"persistence,omitempty"`
	PodTemplate PodTemplate `json:"podTemplate,omitempty"`
}

type Persistence struct {
	Postgresql Postgresql `json:"postgresql,omitempty"`
}

type Postgresql struct {
	SecretRef  PostgresAuthSecret `json:"secretRef,omitempty"`
	ServiceRef ServiceRef         `json:"serviceRef,omitempty"`
}

type ServiceRef struct {
	Name      string `json:"name,omitempty"`
	Namespace string `json:"namespace,omitempty"`
}

type PodTemplate struct {
	Container Container `json:"container,omitempty"`
}

type Container struct {
	Image string `json:"image,omitempty"`
}

type SonataFlowPlatformStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type SonataFlow struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SonataFlowPlatformSpec   `json:"spec,omitempty"`
	Status SonataFlowPlatformStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// SonataFlowPlatformList contains a list of SonataFlow
type SonataFlowPlatformList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SonataFlow `json:"items"`
}

func init() {
	SchemeBuilder.Register(&SonataFlow{}, &SonataFlowPlatformList{})
}
