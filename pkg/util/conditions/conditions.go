package conditions

import (
	"fmt"
	timestamppb "google.golang.org/protobuf/types/known/timestamppb"
	"kubeclusteragent/gen/go/agent/v1alpha1"
	"sort"
)

// TrueCondition returns a condition with Status=True and the given type.
func TrueCondition(t v1alpha1.ConditionType) *v1alpha1.Condition {
	return &v1alpha1.Condition{
		Type:   t,
		Status: "True",
	}
}

// FalseCondition returns a condition with Status=False and the given type.
func FalseCondition(t v1alpha1.ConditionType, reason string, severity string, messageFormat string, messageArgs ...interface{}) *v1alpha1.Condition {
	return &v1alpha1.Condition{
		Type:     t,
		Status:   "False",
		Reason:   reason,
		Severity: severity,
		Message:  fmt.Sprintf(messageFormat, messageArgs...),
	}
}

// UnknownCondition returns a condition with Status=Unknown and the given type.
func UnknownCondition(t v1alpha1.ConditionType, reason string, messageFormat string, messageArgs ...interface{}) *v1alpha1.Condition {
	return &v1alpha1.Condition{
		Type:    t,
		Status:  "Unknown",
		Reason:  reason,
		Message: fmt.Sprintf(messageFormat, messageArgs...),
	}
}

// MarkTrue sets Status=True for the condition with the given type.
func MarkTrue(to *v1alpha1.ClusterStatus, t v1alpha1.ConditionType) {
	Set(to, TrueCondition(t))
}

// MarkUnknown sets Status=Unknown for the condition with the given type.
func MarkUnknown(to *v1alpha1.ClusterStatus, t v1alpha1.ConditionType, reason, messageFormat string, messageArgs ...interface{}) {
	Set(to, UnknownCondition(t, reason, messageFormat, messageArgs...))
}

// MarkFalse sets Status=False for the condition with the given type.
func MarkFalse(to *v1alpha1.ClusterStatus, t v1alpha1.ConditionType, reason string, severity string, messageFormat string, messageArgs ...interface{}) {
	Set(to, FalseCondition(t, reason, severity, messageFormat, messageArgs...))
}

// Delete deletes the condition with the given type.
func Delete(to *v1alpha1.ClusterStatus, t v1alpha1.ConditionType) {
	if to == nil {
		return
	}

	conditions := to.GetConditions()
	newConditions := make([]*v1alpha1.Condition, 0, len(conditions))
	for _, condition := range conditions {
		if condition.Type != t {
			newConditions = append(newConditions, condition)
		}
	}
	to.Conditions = newConditions
}

func DeleteAll(to *v1alpha1.ClusterStatus) {
	if to == nil {
		return
	}
	to.Conditions = make([]*v1alpha1.Condition, 0)
}

func Set(to *v1alpha1.ClusterStatus, condition *v1alpha1.Condition) {
	if to == nil || condition == nil {
		return
	}
	// Check if the new conditions already exists, and change it only if there is a status
	// transition (otherwise we should preserve the current last transition time)-
	conditions := to.GetConditions()
	exists := false
	for i := range conditions {
		existingCondition := conditions[i]
		if existingCondition.Type == condition.Type {
			exists = true
			if !hasSameState(existingCondition, condition) {
				condition.LastTransitionTime = timestamppb.Now()
				conditions[i] = condition
				break
			}
			condition.LastTransitionTime = existingCondition.LastTransitionTime
			break
		}
	}

	// If the condition does not exist, add it, setting the transition time only if not already set
	if !exists {
		if condition.LastTransitionTime.GetSeconds() == 0 {
			condition.LastTransitionTime = timestamppb.Now()
		}
		conditions = append(conditions, condition)
	}

	// Sorts conditions for convenience of the consumer, i.e. kubectl.
	sort.Slice(conditions, func(i, j int) bool {
		return lexicographicLess(conditions[i], conditions[j])
	})

	to.Conditions = conditions
}

// lexicographicLess returns true if a condition is less than another in regard to the
// to order of conditions designed for convenience of the consumer, i.e. kubectl.
// According to this order the Ready condition always goes first, followed by all the other
// conditions sorted by Type.
func lexicographicLess(i, j *v1alpha1.Condition) bool {
	return (i.Type == v1alpha1.ConditionType_ClusterReady || i.Type < j.Type) && j.Type != v1alpha1.ConditionType_ClusterReady
}

// hasSameState returns true if a condition has the same state of another; state is defined
// by the union of following fields: Type, Status, Reason, Severity and Message (it excludes LastTransitionTime).
func hasSameState(i, j *v1alpha1.Condition) bool {
	return i.Type == j.Type &&
		i.Status == j.Status &&
		i.Reason == j.Reason &&
		i.Severity == j.Severity &&
		i.Message == j.Message
}
