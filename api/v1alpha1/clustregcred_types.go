package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ClustRegCredSpec defines the desired state of ClustRegCred
type ClustRegCredSpec struct {
	// Registry is the container registry URL (e.g., https://index.docker.io/v1/)
	// +kubebuilder:validation:MinLength=1
	Registry string `json:"registry"`

	// Username for authenticating to the container registry
	// +kubebuilder:validation:MinLength=1
	Username string `json:"username"`

	// Password for authenticating to the container registry
	// +kubebuilder:validation:MinLength=1
	Password string `json:"password"`

	// Email associated with the container registry account
	// +kubebuilder:default=""
	Email string `json:"email"`

	// SecretName is the name to be used for the image pull secret in each namespace
	// +kubebuilder:validation:MinLength=1
	SecretName string `json:"secretName"`

	// Namespaces is the list of namespaces where the image pull secret should be created
	// +kubebuilder:validation:MinItems=1
	Namespaces []string `json:"namespaces"`
}

// ClustRegCredStatus defines the observed state of ClustRegCred
// +kubebuilder:subresource:status

type ClustRegCredStatus struct {
	SyncedNamespaces []string           `json:"syncedNamespaces,omitempty"`
	LastSynced       string             `json:"lastSynced,omitempty"`
	Phase            string             `json:"phase,omitempty"`
	Reason           string             `json:"reason,omitempty"`
	Conditions       []metav1.Condition `json:"conditions,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// ClustRegCred is the Schema for the clustregcreds API
type ClustRegCred struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ClustRegCredSpec   `json:"spec,omitempty"`
	Status ClustRegCredStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ClustRegCredList contains a list of ClustRegCred
type ClustRegCredList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ClustRegCred `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ClustRegCred{}, &ClustRegCredList{})
}
