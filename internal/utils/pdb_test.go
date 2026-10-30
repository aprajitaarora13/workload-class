/*
Copyright 2026 The Kubernetes Authors.

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

package utils

import (
	"reflect"
	"testing"

	policyv1 "k8s.io/api/policy/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	workloadsv1alpha1 "sigs.k8s.io/workload-class/api/v1alpha1"
)

const (
	testWCName    = "test-wc"
	testNamespace = "default"
)

func TestPDBName(t *testing.T) {
	if got, want := PDBName("my-class"), "workload-my-class"; got != want {
		t.Errorf("PDBName() = %q, want %q", got, want)
	}
}

func TestPDBBase(t *testing.T) {
	if got := PDBBase(nil); got != nil {
		t.Errorf("PDBBase(nil) = %v, want nil", got)
	}

	wc := &workloadsv1alpha1.WorkloadClass{
		ObjectMeta: metav1.ObjectMeta{Name: testWCName, Namespace: "test-ns"},
	}
	pdb := PDBBase(wc)
	if pdb == nil || pdb.Name != "workload-test-wc" || pdb.Namespace != "test-ns" {
		t.Errorf("PDBBase() = %v, want test-ns/workload-test-wc", pdb)
	}
}

func TestSyncPDBWithWorkloadClass(t *testing.T) {
	testSelector := &metav1.LabelSelector{MatchLabels: map[string]string{"app": "test"}}
	ifHealthyBudget := policyv1.IfHealthyBudget
	open := intstr.FromString(openValue)
	closed := intstr.FromInt32(closedValue)

	newWC := func(readiness workloadsv1alpha1.MaintenanceReadiness) *workloadsv1alpha1.WorkloadClass {
		return &workloadsv1alpha1.WorkloadClass{
			ObjectMeta: metav1.ObjectMeta{Name: testWCName, Namespace: testNamespace},
			Spec:       workloadsv1alpha1.WorkloadClassSpec{PodSelector: testSelector},
			Status:     workloadsv1alpha1.WorkloadClassStatus{MaintenanceReadiness: readiness},
		}
	}

	tests := []struct {
		name               string
		wc                 *workloadsv1alpha1.WorkloadClass
		pdb                *policyv1.PodDisruptionBudget
		wantErr            bool
		wantName           string
		wantNamespace      string
		wantMaxUnavailable *intstr.IntOrString
		wantUnhealthyEvict *policyv1.UnhealthyPodEvictionPolicyType
	}{
		{
			name:    "workload_class_is_nil",
			wc:      nil,
			pdb:     &policyv1.PodDisruptionBudget{},
			wantErr: true,
		},
		{
			name:    "pdb_is_nil",
			wc:      newWC(workloadsv1alpha1.ReadinessReady),
			pdb:     nil,
			wantErr: true,
		},
		{
			name:               "maintenance_readiness_ready_populates_empty_metadata",
			wc:                 newWC(workloadsv1alpha1.ReadinessReady),
			pdb:                &policyv1.PodDisruptionBudget{},
			wantName:           PDBName(testWCName),
			wantNamespace:      testNamespace,
			wantMaxUnavailable: &open,
			wantUnhealthyEvict: &ifHealthyBudget,
		},
		{
			name:               "maintenance_readiness_not_ready_preserves_existing_metadata",
			wc:                 newWC(workloadsv1alpha1.ReadinessNotReady),
			pdb:                PDBBase(newWC(workloadsv1alpha1.ReadinessNotReady)),
			wantName:           PDBName(testWCName),
			wantNamespace:      testNamespace,
			wantMaxUnavailable: &closed,
			wantUnhealthyEvict: &ifHealthyBudget,
		},
		{
			name:               "maintenance_readiness_overdue",
			wc:                 newWC(workloadsv1alpha1.ReadinessOverdue),
			pdb:                &policyv1.PodDisruptionBudget{},
			wantName:           PDBName(testWCName),
			wantNamespace:      testNamespace,
			wantMaxUnavailable: &open,
			wantUnhealthyEvict: &ifHealthyBudget,
		},
		{
			name:               "maintenance_readiness_unset_defaults_to_closed",
			wc:                 newWC(""),
			pdb:                &policyv1.PodDisruptionBudget{},
			wantName:           PDBName(testWCName),
			wantNamespace:      testNamespace,
			wantMaxUnavailable: &closed,
			wantUnhealthyEvict: &ifHealthyBudget,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := SyncPDBWithWorkloadClass(tt.wc, tt.pdb)
			if (err != nil) != tt.wantErr {
				t.Fatalf("SyncPDBWithWorkloadClass() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr {
				if tt.pdb.Name != tt.wantName || tt.pdb.Namespace != tt.wantNamespace {
					t.Errorf("SyncPDBWithWorkloadClass() metadata = %s/%s, want %s/%s", tt.pdb.Namespace, tt.pdb.Name, tt.wantNamespace, tt.wantName)
				}
				if !reflect.DeepEqual(tt.pdb.Spec.MaxUnavailable, tt.wantMaxUnavailable) {
					t.Errorf("SyncPDBWithWorkloadClass() MaxUnavailable = %v, want %v", tt.pdb.Spec.MaxUnavailable, tt.wantMaxUnavailable)
				}
				if !reflect.DeepEqual(tt.pdb.Spec.UnhealthyPodEvictionPolicy, tt.wantUnhealthyEvict) {
					t.Errorf("SyncPDBWithWorkloadClass() UnhealthyPodEvictionPolicy = %v, want %v", tt.pdb.Spec.UnhealthyPodEvictionPolicy, tt.wantUnhealthyEvict)
				}
				if !reflect.DeepEqual(tt.pdb.Spec.Selector, tt.wc.Spec.PodSelector) {
					t.Errorf("SyncPDBWithWorkloadClass() Selector = %v, want %v", tt.pdb.Spec.Selector, tt.wc.Spec.PodSelector)
				}
			}
		})
	}
}
