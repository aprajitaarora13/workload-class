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

package controller

import (
	"context"
	"strings"

	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	workloadsv1alpha1 "sigs.k8s.io/workload-class/api/v1alpha1"
	"sigs.k8s.io/workload-class/internal/utils"
)

// WorkloadClassGuardrailReconciler reconciles a WorkloadClassGuardrail object
type WorkloadClassGuardrailReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=workloads.x-k8s.io,resources=workloadclassguardrails,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=workloads.x-k8s.io,resources=workloadclassguardrails/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=workloads.x-k8s.io,resources=workloadclassguardrails/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *WorkloadClassGuardrailReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	g := &workloadsv1alpha1.WorkloadClassGuardrail{}
	if err := r.Get(ctx, req.NamespacedName, g); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	validationCondition := r.validate(ctx, g)
	meta.SetStatusCondition(&g.Status.Conditions, validationCondition)
	err := r.Status().Update(ctx, g)

	return ctrl.Result{}, err
}

func (r *WorkloadClassGuardrailReconciler) validate(ctx context.Context, g *workloadsv1alpha1.WorkloadClassGuardrail) metav1.Condition {
	log := logf.FromContext(ctx)
	var violations []string
	if err := utils.WeekdaysValid(g.Spec.Constraints.Disruption.DaysOfWeek); err != nil {
		log.Error(err, "validation of DaysOfWeek failed")
		violations = append(violations, err.Error())
	}

	return condition(violations)
}

func condition(violations []string) metav1.Condition {
	if len(violations) > 0 {
		return metav1.Condition{
			Type:               workloadsv1alpha1.ConditionTypeValidated,
			Status:             metav1.ConditionFalse,
			Reason:             workloadsv1alpha1.ReasonValidationFailed,
			Message:            strings.Join(violations, "; "),
			LastTransitionTime: metav1.Now(),
		}
	}

	return metav1.Condition{
		Type:               workloadsv1alpha1.ConditionTypeValidated,
		Status:             metav1.ConditionTrue,
		Reason:             workloadsv1alpha1.ReasonValidationPassed,
		Message:            "WorkloadClass adheres to all Guardrail constraints",
		LastTransitionTime: metav1.Now(),
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *WorkloadClassGuardrailReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&workloadsv1alpha1.WorkloadClassGuardrail{}).
		Named("workloadclassguardrail").
		Complete(r)
}
