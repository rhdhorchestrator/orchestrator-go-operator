package v1alpha1

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

type SonataFlowClusterSpec struct {
	PlatformRef PlatformRef `json:"platformRef,omitempty"`
}

type PlatformRef struct {
	PlatformName      string `json:"name,omitempty"`
	PlatformNamespace string `json:"namespace,omitempty"`
}

type SonataFlowClusterStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type SonataFlowCluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SonataFlowClusterSpec   `json:"spec,omitempty"`
	Status SonataFlowClusterStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// SonataFlowPlatformList contains a list of SonataFlow
type SonataFlowClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SonataFlowCluster `json:"items"`
}

func init() {
	SchemeBuilder.Register(&SonataFlowCluster{}, &SonataFlowClusterList{})
}
