/*
Copyright 2024.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	RunningPhase   OrchestratorPhase = "Running"
	CompletedPhase OrchestratorPhase = "Completed"
	FailedPhase    OrchestratorPhase = "Failed"
)

// OrchestratorSpec defines the desired state of Orchestrator
type OrchestratorSpec struct {
	ServerlessLogicOperator ServerlessLogicOperator `json:"serverlessLogicOperator,omitempty"`
	ServerlessOperator      ServerlessOperator      `json:"serverlessOperator,omitempty"`
	RHDHConfig              RHDHConfig              `json:"rhdh,omitempty"`
	PostgresDB              Postgres                `json:"postgres,omitempty"`
	OrchestratorConfig      OrchestratorConfig      `json:"orchestrator,omitempty"`
	Tekton                  Tekton                  `json:"tekton,omitempty"`
	ArgoCd                  ArgoCD                  `json:"argocd,omitempty"`
}

type ServerlessLogicOperator struct {
	Enabled bool `json:"enabled,omitempty"`
}

type ServerlessOperator struct {
	Enabled bool `json:"enabled,omitempty"`
}

type RHDHConfig struct {
	RHDHName        string      `json:"name,omitempty"`
	RHDHNamespace   string      `json:"namespace,omitempty"`
	InstallOperator bool        `json:"installOperator,omitempty"`
	DevMode         bool        `json:"devMode,omitempty"`
	RHDHPlugins     RHDHPlugins `json:"plugins,omitempty"`
}

type RHDHPlugins struct {
	NotificationsConfig NotificationConfig `json:"notificationsEmail,omitempty"`
}

type NotificationConfig struct {
	Enabled   bool   `json:"enabled,omitempty"`
	Port      int    `json:"port,omitempty"`
	Sender    string `json:"sender,omitempty"`
	Recipient string `json:"replyTo,omitempty"`
}

//type PluginDetails struct {
//	Package   string `json:"package,omitempty"`
//	Integrity string `json:"integrity,omitempty"`
//}

type Postgres struct {
	ServiceName      string             `json:"serviceName,omitempty"`
	ServiceNameSpace string             `json:"serviceNamespace,omitempty"`
	AuthSecret       PostgresAuthSecret `json:"authSecret,omitempty"`
	DatabaseName     string             `json:"database,omitempty"`
}

type PostgresAuthSecret struct {
	SecretName  string `json:"name,omitempty"`
	UserKey     string `json:"userKey,omitempty"`
	PasswordKey string `json:"passwordKey,omitempty"`
}

type OrchestratorConfig struct {
	Namespace          string             `json:"namespace,omitempty"`
	SonataFlowPlatform SonataFlowPlatform `json:"sonataFlowPlatform,omitempty"`
}

type SonataFlowPlatform struct {
	Resources Resource `json:"resources,omitempty"`
}

type Resource struct {
	Requests MemoryCpu `json:"requests,omitempty"`
	Limits   MemoryCpu `json:"limits,omitempty"`
}

type MemoryCpu struct {
	Memory string `json:"memory,omitempty"`
	Cpu    string `json:"cpu,omitempty"`
}

type Tekton struct {
	Enabled bool `json:"enabled,omitempty"`
}

type ArgoCD struct {
	Enabled   bool `json:"enabled,omitempty"`
	Namespace bool `json:"namespace,omitempty"`
}

type OrchestratorPhase string

// OrchestratorStatus defines the observed state of Orchestrator
type OrchestratorStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Conditions []metav1.Condition `json:"conditions,omitempty"`
	// +kubebuilder:validation:Enum={"Running","Completed", "Failed"}
	Phase OrchestratorPhase `json:"phase,omitempty" protobuf:"bytes,1,opt,casttype=OrchestratorPhase"`
}

//+kubebuilder:object:root=true

// Orchestrator is the Schema for the orchestrators API
// +kubebuilder:subresource:status
type Orchestrator struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   OrchestratorSpec   `json:"spec,omitempty"`
	Status OrchestratorStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// OrchestratorList contains a list of Orchestrator
type OrchestratorList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Orchestrator `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Orchestrator{}, &OrchestratorList{})
}
