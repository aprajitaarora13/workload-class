# workload-class

A high-level Kubernetes control layer for managing workload outcomes and disruption guardrails during maintenance.

## Overview

`workload-class` provides a two-tier API under `workloads.x-k8s.io/v1alpha1` that allows platform administrators and workload owners to collaborate on managing workload disruptions during cluster maintenance:

1. **`WorkloadClassGuardrail` (Cluster-scoped)**: Defined by platform engineers to enforce cluster-wide boundaries on disruption policies (such as allowed days of the week, maximum number of maintenance windows, and maximum non-disruption duration to prevent maintenance starvation).
2. **`WorkloadClass` (Namespace-scoped)**: Defined by workload owners to declare pod disruption policies (temporal maintenance windows, minimum initial run duration, graceful termination cool-down, and outside-of-window subject exceptions) within platform guardrails.

### Pod Selection & Namespace Default Precedence

A `WorkloadClass` is resolved for a Pod using the following order of precedence:

1. **Namespace-level default (`workloads.x-k8s.io/default-class`)**: If a `Namespace` carries the `workloads.x-k8s.io/default-class: <workload-class-name>` label (`DefaultWorkloadClassLabel`), that `WorkloadClass` applies to all Pods in the namespace, overriding its own `spec.podSelector` as well as any other `WorkloadClass` in that namespace.
2. **Pod label selector (`spec.podSelector`)**: If no namespace default is set, the `WorkloadClass` in the Pod's namespace whose `spec.podSelector` matches the Pod with the highest specificity (`len(matchLabels) + len(matchExpressions)`) is selected. If multiple `WorkloadClass` resources match with the same specificity, the oldest `WorkloadClass` (by `metadata.creationTimestamp`) takes precedence.

### Relationship to Upstream `scheduling.k8s.io` `Workload`

`workloads.x-k8s.io` `WorkloadClass` and upstream `scheduling.k8s.io` `Workload` (KEP-4671) address orthogonal layers of workload management:

- **`scheduling.k8s.io` `Workload`** models a **runtime scheduling unit and pod-group topology** (such as gang scheduling constraints for batch/AI jobs and gang disruption policies that define *what unit* must be disrupted together—e.g., whether evicting a single Pod requires disrupting and recreating the entire Pod group).
- **`workloads.x-k8s.io` `WorkloadClass`** defines **operational maintenance and disruption policies** (*when* and *under what organizational guardrails* disruptions are allowed—such as recurring time-zone-aware maintenance windows, minimum initial run durations, and starvation bounds via `maxNonDisruptionDurationDays`).

The two APIs are complementary: `scheduling.k8s.io` `Workload` governs structural pod-group scheduling and gang disruption scope, while `WorkloadClass` governs temporal readiness and guardrail enforcement for maintenance disruptions.

## Community, discussion, contribution, and support

Learn how to engage with the Kubernetes community on the [community page](http://kubernetes.io/community/).

You can reach the maintainers of this project at:

- [Slack channel](https://kubernetes.slack.com/messages/sig-apps)
- [Mailing List](https://groups.google.com/a/kubernetes.io/g/sig-apps)

### Code of conduct

Participation in the Kubernetes community is governed by the [Kubernetes Code of Conduct](code-of-conduct.md).
