package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// TestSuiteRunSpec defines the desired state of the TestSuiteRun
type TestSuiteRunSpec struct {
	MachineType string `json:"machineType"`
	TestName    string `json:"testName"`
}

// TestSuiteRunStatus defines the observed state of the TestSuiteRun
type TestSuiteRunStatus struct {
	Phase string `json:"phase,omitempty"`
}

// TestSuiteRun is the Schema for the TestSuiteRun API
type TestSuiteRun struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              TestSuiteRunSpec   `json:"spec,omitempty"`
	Status            TestSuiteRunStatus `json:"status,omitempty"`
}

// TestSuiteRunList contains a list of TestSuiteRun
type TestSuiteRunList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []TestSuiteRun `json:"items"`
}
