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
	"fmt"

	policyv1 "k8s.io/api/policy/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	workloadsv1alpha1 "sigs.k8s.io/workload-class/api/v1alpha1"
)

const (
	PDBNamePrefix = "workload-"

	openValue   = "100%" // Allows 100% of Pods to be unavailable
	closedValue = 0      // Allows 0 Pods to be unavailable
)

// PDBName returns the name of the PDB based on the WorkloadClass name.
// PDBs made for WorkloadClasses will have the prefix `workload-`.
func PDBName(wcName string) string {
	return PDBNamePrefix + wcName
}

// PDBBase returns a basic PDB, configured based on the WorkloadClass.
func PDBBase(wc *workloadsv1alpha1.WorkloadClass) *policyv1.PodDisruptionBudget {
	if wc == nil {
		return nil
	}
	return &policyv1.PodDisruptionBudget{
		ObjectMeta: metav1.ObjectMeta{
			Name:      PDBName(wc.Name),
			Namespace: wc.Namespace,
		},
	}
}

// SyncPDBWithWorkloadClass configures the PDB to match the WorkloadClass' disruption policy.
func SyncPDBWithWorkloadClass(wc *workloadsv1alpha1.WorkloadClass, pdb *policyv1.PodDisruptionBudget) error {
	if wc == nil {
		return fmt.Errorf("failed to sync PDB with WorkloadClass, WorkloadClass is nil")
	}

	if pdb == nil {
		return fmt.Errorf("failed to sync PDB with WorkloadClass, PDB is nil")
	}

	if pdb.Name == "" {
		pdb.Name = PDBName(wc.Name)
	}
	if pdb.Namespace == "" {
		pdb.Namespace = wc.Namespace
	}

	// Selectors must match exactly
	pdb.Spec.Selector = wc.Spec.PodSelector.DeepCopy()

	// IfHealthyBudget will block evictions of unhealthy Pods if there is no budget
	policy := policyv1.IfHealthyBudget
	pdb.Spec.UnhealthyPodEvictionPolicy = &policy

	open := intstr.FromString(openValue)
	closed := intstr.FromInt32(closedValue)

	maxUnavailableMap := map[workloadsv1alpha1.MaintenanceReadiness]*intstr.IntOrString{
		workloadsv1alpha1.ReadinessReady:    &open,
		workloadsv1alpha1.ReadinessNotReady: &closed,
		workloadsv1alpha1.ReadinessOverdue:  &open,
	}

	// Set MaxUnavailable based on the WorkloadClass' MaintenanceReadiness
	if val, ok := maxUnavailableMap[wc.Status.MaintenanceReadiness]; ok {
		pdb.Spec.MaxUnavailable = val
	} else {
		pdb.Spec.MaxUnavailable = &closed
	}

	return nil
}
