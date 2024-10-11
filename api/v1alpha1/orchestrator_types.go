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

// OrchestratorSpec defines the desired state of Orchestrator
type OrchestratorSpec struct {
	SonataFlowOperator   SonataFlowOperator   `json:"sonataFlowOperator,omitempty"`
	ServerlessOperator   ServerlessOperator   `json:"serverlessOperator,omitempty"`
	RhdhOperator         RHDHOperator         `json:"rhdhOperator,omitempty"`
	RhdhPlugins          RHDHPlugins          `json:"rhdhPlugins,omitempty"`
	PostgresDB           Postgres             `json:"postgres,omitempty"`
	OrchestratorPlatform OrchestratorPlatform `json:"orchestrator,omitempty"`
	Tekton               Tekton               `json:"tekton,omitempty"`
	ArgoCd               ArgoCD               `json:"argocd,omitempty"`
}

// reuse from the subscription - check from the api/compare with the subscription object
// do we want to expose all the spec within the inherent subscription
// inline embedding to add field in the subscription object
// ask Moti to confirm - breaking changes
type Subscription struct {
	Namespace           string `json:"namespace,omitempty"`
	Channel             string `json:"channel,omitempty"`
	InstallPlanApproval string `json:"installPlanApproval,omitempty"`
	Name                string `json:"name,omitempty"`
	SourceName          string `json:"sourceName,omitempty"`
	StartingCSV         string `json:"startingCSV,omitempty"`
	TargetNamespace     string `json:"targetNamespace,omitempty"`
}

type SonataFlowOperator struct {
	IsReleaseCandidate bool         `json:"isReleaseCandidate,omitempty"`
	Enabled            bool         `json:"enabled,omitempty"`
	Subscription       Subscription `json:"subscription,omitempty"`
}

type ServerlessOperator struct {
	Enabled      bool         `json:"enabled,omitempty"`
	Subscription Subscription `json:"subscription,omitempty"`
}

type BackstageSecret struct {
	BackendSecret string `json:"backendSecret,omitempty"`
}

type ClusterTokenUrl struct {
	ClusterToken string `json:"clusterToken,omitempty"`
	ClusterUrl   string `json:"clusterUrl,omitempty"`
}

type GithubBS struct {
	Token        string `json:"token,omitempty"`
	ClientID     string `json:"clientId,omitempty"`
	ClientSecret string `json:"clientSecret,omitempty"`
}

type ArgoCDBS struct {
	Enabled   bool   `json:"enabled,omitempty"`
	Namespace string `json:"namespace,omitempty"`
	Url       string `json:"url,omitempty"`
	Username  string `json:"username,omitempty"`
	Password  string `json:"password,omitempty"`
}

type NotificationEmailBS struct {
	Hostname string `json:"hostname,omitempty"`
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
}

type SecretRefBS struct {
	Name              string              `json:"name,omitempty"`
	Backstage         BackstageSecret     `json:"backstage,omitempty"`
	Github            GithubBS            `json:"github,omitempty"`
	ClusterTokenUrl   ClusterTokenUrl     `json:"k8s,omitempty"`
	ArgoCD            ArgoCDBS            `json:"argoCD,omitempty"`
	NotificationEmail NotificationEmailBS `json:"notificationEmail,omitempty"`
}

type RHDHOperator struct {
	IsReleaseCandidate   bool         `json:"isReleaseCandidate,omitempty"`
	Enabled              bool         `json:"enabled,omitempty"`
	EnabledGuestProvider bool         `json:"enabledGuestProvider,omitempty"`
	CatalogBranch        string       `json:"catalogBranch,omitempty"`
	Subscription         Subscription `json:"subscription,omitempty"`
	SecretRef            SecretRefBS  `json:"secretRef,omitempty"`
}

type PluginDetails struct {
	Package   string `json:"package,omitempty"`
	Integrity string `json:"integrity,omitempty"`
}

type NotificationConfig struct {
	Enabled   bool   `json:"enabled,omitempty"`
	Port      int64  `json:"port,omitempty"`
	Sender    string `json:"sender,omitempty"`
	Recipient string `json:"replyTo,omitempty"`
}

type RHDHPlugins struct {
	NpmRegistry        string                   `json:"npmRegistry,omitempty"`
	Scope              string                   `json:"scope,omitempty"`
	Plugins            map[string]PluginDetails `json:"plugins,omitempty"`
	NotificationConfig NotificationConfig       `json:"notificationConfig,omitempty"`
}

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

type OrchestratorPlatform struct {
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

// OrchestratorStatus defines the observed state of Orchestrator
type OrchestratorStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Conditions []metav1.Condition `json:"conditions,omitempty"`
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
