package v1alpha1

import (
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const GrafanaPlaylistKind = "GrafanaPlaylist"

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// GrafanaPlaylistSpec defines the desired state of GrafanaPlaylist
// +k8s:openapi-gen=true
type GrafanaPlaylistSpec struct {
	Name     string                `json:"name"`
	Uid      string                `json:"uid,omitempty"`
	Interval time.Duration         `json:"interval"`
	Items    []GrafanaPlaylistItem `json:"items"`
}

// GrafanaPlaylistItem defines and item in a Grafana playlist
// +k8s:openapi-gen=true
type GrafanaPlaylistItem struct {
	Order int    `json:"order"`
	Title string `json:"title"`
	Type  string `json:"type"`
	Value string `json:"value"`
}

// GrafanaPlaylistStatus defines the observed state of GrafanaPlaylist
// +k8s:openapi-gen=true
type GrafanaPlaylistStatus struct {
	Phase   StatusPhase `json:"phase"`
	Message string      `json:"message"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// GrafanaPlaylist is the Schema for the grafanaplaylists API
// +k8s:openapi-gen=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=grafanaplaylists,scope=Namespaced
type GrafanaPlaylist struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   GrafanaPlaylistSpec   `json:"spec,omitempty"`
	Status GrafanaPlaylistStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// GrafanaPlaylistList contains a list of GrafanaPlaylist
type GrafanaPlaylistList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []GrafanaPlaylist `json:"items"`
}

func init() {
	SchemeBuilder.Register(&GrafanaPlaylist{}, &GrafanaPlaylistList{})
}
