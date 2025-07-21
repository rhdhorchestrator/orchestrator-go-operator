/*
Copyright 2024 Red Hat, Inc.

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

package v1alpha3

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	RunningPhase   OrchestratorPhase = "Running"
	CompletedPhase OrchestratorPhase = "Completed"
	FailedPhase    OrchestratorPhase = "Failed"
)

// OrchestratorSpec defines the desired state of Orchestrator
type OrchestratorSpec struct {
	// Configuration for ServerlessLogic. Optional
	// +kubebuilder:default={installOperator: true}
	ServerlessLogicOperator ServerlessLogicOperator `json:"serverlessLogic,omitempty"`

	// Configuration for Serverless (K-Native) Operator. Optional
	// +kubebuilder:default={installOperator: true}
	ServerlessOperator ServerlessOperator `json:"serverless,omitempty"`

	// Configuration for RHDH (Backstage).
	// +kubebuilder:validation:Required
	RHDHConfig RHDHConfig `json:"rhdh"`

	// Configuration for existing database instance
	// Used by Data index and Job service
	// +kubebuilder:validation:Required
	PostgresConfig PostgresConfig `json:"postgres"`

	// Configuration for Orchestrator. Optional
	PlatformConfig PlatformConfig `json:"platform,omitempty"`

	// Contains the configuration for the infrastructure services required for the Orchestrator to serve workflows
	// by leveraging the OpenShift Serverless and OpenShift Serverless Logic capabilities. Optional
	// +kubebuilder:default={enabled: false}
	Tekton Tekton `json:"tekton,omitempty"`

	// Configuration for ArgoCD. Optional
	// +kubebuilder:default={enabled: false}
	ArgoCd ArgoCD `json:"argocd,omitempty"`
}

type ServerlessLogicOperator struct {
	// Determines whether to install the ServerlessLogic operator
	// +kubebuilder:default=true
	InstallOperator bool `json:"installOperator"`
}

type ServerlessOperator struct {
	// Determines whether to install the Serverless operator
	// +kubebuilder:default=true
	InstallOperator bool `json:"installOperator"`
}

type RHDHConfig struct {
	// Name of RHDH CR, whether existing or to be installed
	// +kubebuilder:validation:Required
	Name string `json:"name"`

	// Namespace of RHDH Instance, whether existing or to be installed
	// +kubebuilder:validation:Required
	Namespace string `json:"namespace"`

	// Determines whether the RHDH operator should be installed
	// This determines the deployment of the RHDH instance.
	// Defaults to false
	// +kubebuilder:default=false
	InstallOperator bool `json:"installOperator,omitempty"`

	// Determines whether to enable the guest provider in RHDH.
	// This should be used for development purposes ONLY and should not be enabled in production.
	// Defaults to false.
	// +kubebuilder:default=false
	DevMode bool `json:"devMode,omitempty"`

	// Configuration for RHDH Plugins.
	RHDHPlugins RHDHPlugins `json:"plugins,omitempty"`
}

type RHDHPlugins struct {
	// Notification email plugin configuration
	NotificationsConfig NotificationConfig `json:"notificationsEmail,omitempty"`
}

type NotificationConfig struct {
	// Determines whether to install the Notifications Email plugin
	// Requires setting the hostname and credentials in RHDH secret
	// The secret backstage-backend-auth-secret is created as pre-requisite
	// See plugin configuration at https://github.com/backstage/backstage/blob/master/plugins/notifications-backend-module-email/config.d.ts
	// +kubebuilder:default=false
	Enabled bool `json:"enabled,omitempty"`

	// SMTP server port
	// +kubebuilder:default=587
	Port int `json:"port,omitempty"`

	// Email address of the Sender
	// +kubebuilder:default=""
	Sender string `json:"sender,omitempty"`

	// Email address of the Recipient
	// +kubebuilder:default=""
	Recipient string `json:"replyTo,omitempty"`
}

type PostgresConfig struct {
	// Name of the PostgresConfig DB service to be used by platform services
	// +kubebuilder:validation:Required
	Name string `json:"name"`

	// Namespace of the PostgresConfig DB service to be used by platform services
	// +kubebuilder:validation:Required
	Namespace string `json:"namespace"`

	// PostgreSQL connection credentials details
	// +kubebuilder:validation:Required
	AuthSecret PostgresAuthSecret `json:"authSecret"`

	// Existing database instance used by data index and job service
	// +kubebuilder:validation:Required
	DatabaseName string `json:"database"`
}

type PostgresAuthSecret struct {
	// Name of existing secret to use for PostgreSQL credentials.
	// +kubebuilder:validation:Required
	SecretName string `json:"name"`

	// Name of key in existing secret to use for PostgreSQL credentials.
	// +kubebuilder:validation:Required
	UserKey string `json:"userKey"`

	// Name of key in existing secret to use for PostgreSQL credentials.
	// +kubebuilder:validation:Required
	PasswordKey string `json:"passwordKey"`
}

type PlatformConfig struct {
	// Namespace of the workflow pods (Data Index and Job Service) and SonataFlow CR.
	// +kubebuilder:validation:Required
	Namespace string `json:"namespace"`

	// Resource configuration to be used for the data index and job services.
	Resources Resource `json:"resources,omitempty"`

	// Configuration for existing eventing to be used by sonataflow platform
	Eventing Eventing `json:"eventing,omitempty"`

	// Configuration for sonataflow platform monitoring
	Monitoring MonitoringConfig `json:"monitoring,omitempty"`
}

type Eventing struct {
	// Configuration for K-Native broker.
	Broker Broker `json:"broker,omitempty"`
}

type MonitoringConfig struct {
	// Determines whether to enable the platform monitoring
	// +kubebuilder:default=false
	Enabled bool `json:"enabled,omitempty"`
}

type Broker struct {
	// Name of existing Broker instance
	Name string `json:"name,omitempty"`

	// Namespace of existing Broker instance
	Namespace string `json:"namespace,omitempty"`
}

type Resource struct {
	// Describe the minimum amount of compute resources required.
	// Requests cannot exceed Limits. More info: https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/
	Requests MemoryCpu `json:"requests,omitempty"`
	// Describes the maximum amount of compute resources allowed.
	// More info: https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/'
	Limits MemoryCpu `json:"limits,omitempty"`
}

type MemoryCpu struct {
	// Defines the memory resource limits
	// +kubebuilder:default="1Gi"
	Memory string `json:"memory,omitempty"`

	// Defines the CPU resource limits
	// +kubebuilder:default="500m"
	Cpu string `json:"cpu,omitempty"`
}

type Tekton struct {
	// Determines whether to create the Tekton pipeline resources. Defaults to false.
	// +kubebuilder:default=false
	Enabled bool `json:"enabled,omitempty"`
}

type ArgoCD struct {
	// Determines whether to install the ArgoCD plugin and create the orchestrator AppProject
	// +kubebuilder:default=false
	Enabled bool `json:"enabled,omitempty"`

	// Namespace where the ArgoCD operator is installed and watching for argoapp CR instances
	// Ensure to add the Namespace if ArgoCD is installed
	Namespace string `json:"namespace,omitempty"`
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
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=".metadata.creationTimestamp",description="Age"
// +kubebuilder:printcolumn:name="Phase",type=string,JSONPath=".status.phase",description="Status"
// +kubebuilder:metadata:annotations=orchestrator-package=backstage-plugin-orchestrator-1.6.1-rc.2.tgz
// +kubebuilder:metadata:annotations=orchestrator-integrity=sha512-TJ58d5CqFcNmvhBPJp+/7nt0gZo4ILqRjE2+9ZHjIVht2X0gCJqqGYF41sTgBotb2biOD024W/5xp2qQzRbaww==
// +kubebuilder:metadata:annotations=orchestrator-backend-dynamic-package=backstage-plugin-orchestrator-backend-dynamic-1.6.1-rc.2.tgz
// +kubebuilder:metadata:annotations=orchestrator-backend-dynamic-integrity=sha512-qveMcu8jO2KsKzgXioNmmbQKxGUbUloWbDxZfa3sQDSGakB6RSE5kNPTAy1QmCvBqufeOFrfv36LpV7d757SHA==
// +kubebuilder:metadata:annotations=orchestrator-scaffolder-backend-package=backstage-plugin-scaffolder-backend-module-orchestrator-dynamic-1.6.1-rc.2.tgz
// +kubebuilder:metadata:annotations=orchestrator-scaffolder-backend-integrity=sha512-fq5oUIVyigshMUHD5N85937wCLIQVixV+mvVmCjl99FvY7A4/5X11vASHOFx+1cLW7zBZDT5hc3zJlPDBR2zWQ==
// +kubebuilder:metadata:annotations=orchestrator-form-widgets-package=backstage-plugin-orchestrator-form-widgets-1.6.1-rc.2.tgz
// +kubebuilder:metadata:annotations=orchestrator-form-widgets-integrity=sha512-1KDZmf+iJUevivLsamiD/wvGhuK9PZeGrPNz5wevFC4eXYHB1Iq+Nugjq1IqBQWKIY3jhAIMwba1ZVz4jlu7/A==
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
