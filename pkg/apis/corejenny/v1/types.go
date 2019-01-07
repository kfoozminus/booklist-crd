package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const NamespaceDefault string = "default"

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type Podjenny struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              PodjennySpec `json:"spec,omitempty"`
}

type PodjennySpec struct {
	Image string `json:"image,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type PodjennyList struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Items             []Podjenny `json:"items,omitempty"`
}
