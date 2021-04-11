{{ .Header }}

package {{ .Package }}

import (
{{ .Imports }}
)

// {{ .CRD.Kind }}Spec defines the desired state of a
// {{ .CRD.Kind }}.
type {{ .CRD.Kind }}Spec struct {
	xpv1.ResourceSpec `json:",inline"`
	ForProvider       {{ .CRD.Kind }}Parameters `json:"forProvider"`
}

// {{ .CRD.Kind }}Status represents the observed state of a
// {{ .CRD.Kind }}.
type {{ .CRD.Kind }}Status struct {
	xpv1.ResourceStatus `json:",inline"`
	AtProvider          {{ .CRD.Kind }}Observation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true

// {{ .CRD.Kind }} is a managed resource that represents a Google IAM Service Account.
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="DISPLAYNAME",type="string",JSONPath=".spec.forProvider.displayName"
// +kubebuilder:printcolumn:name="EMAIL",type="string",JSONPath=".status.atProvider.email"
// +kubebuilder:printcolumn:name="DISABLED",type="boolean",JSONPath=".status.atProvider.disabled"
// +kubebuilder:resource:scope=Cluster
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,gcp}
type {{ .CRD.Kind }} struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   {{ .CRD.Kind }}Spec   `json:"spec"`
	Status {{ .CRD.Kind }}Status `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// {{ .CRD.Kind }}List contains a list of {{ .CRD.Kind }} types
type {{ .CRD.Kind }}List struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []{{ .CRD.Kind }} `json:"items"`
}