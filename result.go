package goformkeeper

type Result struct {
	ValidFields     map[string]string
	ValidSelections map[string][]string
	Failures        map[string]*Failure
}

func NewResult() *Result {
	return &Result{
		ValidFields:     make(map[string]string),
		ValidSelections: make(map[string][]string),
		Failures:        make(map[string]*Failure),
	}
}

type Failure struct {
	FieldName   string
	Constraints map[string]*ConstraintFailure
	Message     string
}

type ConstraintFailure struct {
	ConstraintType string
	Message        string
}

func NewFailureForSelection(selection *Selection) *Failure {
	return &Failure{
		FieldName:   selection.Name,
		Message:     selection.Message,
		Constraints: make(map[string]*ConstraintFailure),
	}
}

func NewFailureForField(field *Field) *Failure {
	return &Failure{
		FieldName:   field.Name,
		Message:     field.Message,
		Constraints: make(map[string]*ConstraintFailure),
	}
}

func (failure *Failure) failOnConstraint(constraint *Constraint) {
	constraintFailure := &ConstraintFailure{
		ConstraintType: constraint.Type,
		Message:        constraint.Message,
	}
	failure.Constraints[constraint.Type] = constraintFailure
}

func (result *Result) AddFailure(failure *Failure) {
	result.Failures[failure.FieldName] = failure
}

func (result *Result) putRequiredFailure(fieldName, message string) {
	cFailures := make(map[string]*ConstraintFailure, 0)
	cFailures["required"] = &ConstraintFailure{
		ConstraintType: "required",
		Message:        message,
	}
	result.Failures[fieldName] = &Failure{
		FieldName:   fieldName,
		Message:     message,
		Constraints: cFailures,
	}
}

func (result *Result) ValidParam(name string) string {
	return result.ValidFields[name]
}

func (result *Result) ValidSelection(name string) []string {
	return result.ValidSelections[name]
}

func (result *Result) HasFailure() bool {
	return len(result.Failures) > 0
}

func (result *Result) FailedOn(field string) bool {
	_, found := result.Failures[field]
	return found
}

func (result *Result) FailedOnConstraint(fieldName, constraintName string) bool {
	failure, found := result.Failures[fieldName]
	if !found {
		return false
	}
	_, found = failure.Constraints[constraintName]
	return found
}

func (result *Result) FailedFields() []string {
	fields := make([]string, 0)
	for fieldName, _ := range result.Failures {
		fields = append(fields, fieldName)
	}
	return fields
}

func (result *Result) FailedConstraintsOn(fieldName string) []string {
	failure, found := result.Failures[fieldName]
	if !found {
		return []string{}
	}
	constraints := make([]string, 0)
	for constraintName, _ := range failure.Constraints {
		constraints = append(constraints, constraintName)
	}
	return constraints
}

func (result *Result) Messages() []string {
	builder := NewUniqueStringArrayBuilder(0)
	for fieldName, _ := range result.Failures {
		message := result.MessageOn(fieldName)
		if message != "" {
			builder.Add(message)
		}
	}
	return builder.Build()
}

func (result *Result) MessageOn(field string) string {
	failure, found := result.Failures[field]
	if found {
		return failure.Message
	} else {
		return ""
	}
}

func (result *Result) MessagesOn(fieldName string) []string {
	failure, found := result.Failures[fieldName]
	if !found {
		return []string{}
	}
	builder := NewUniqueStringArrayBuilder(10)
	for _, constraint := range failure.Constraints {
		if constraint.Message != "" {
			builder.Add(constraint.Message)
		}
	}
	return builder.Build()
}

func (result *Result) MessageOnConstraint(fieldName, constraintName string) string {
	failure, found := result.Failures[fieldName]
	if !found {
		return ""
	}
	constraint, found := failure.Constraints[constraintName]
	if !found {
		return failure.Message
	} else {
		if constraint.Message == "" {
			return failure.Message
		} else {
			return constraint.Message
		}
	}
}
