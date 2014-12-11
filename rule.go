package goformkeeper

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"

	yaml "gopkg.in/yaml.v1"
)

type Rule struct {
	Fields     map[string]*Field
	Selections map[string]*Selection
	Forms      map[string]*Form
}

type Form struct {
	Fields     []*Field
	Selections []*Selection
}

type Field struct {
	Name        string
	Ref         string
	Required    bool
	Message     string
	Filters     []string
	Constraints []*Constraint
	FallThrough bool
}

type Selection struct {
	Name        string
	Ref         string
	Count       *Count
	Message     string
	Filters     []string
	Constraints []*Constraint
	FallThrough bool
}

type Count struct {
	From int
	To   int
}

func newRule() *Rule {
	return &Rule{
		Fields:     make(map[string]*Field),
		Selections: make(map[string]*Selection),
		Forms:      make(map[string]*Form),
	}
}

func LoadRuleFromDir(dirPath string) (*Rule, error) {
	r := newRule()
	err := filepath.Walk(dirPath,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() {
				rule, err := LoadRuleFromFile(path)
				if err != nil {
					return err
				}
				r.Merge(rule)
			}
			return nil
		})
	if err != nil {
		return nil, err
	}
	return r, nil
}

func LoadRuleFromFile(filePath string) (*Rule, error) {
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("Failed to read form-rule %s: %s", filePath, err.Error())
	}

	r := &Rule{}
	err = yaml.Unmarshal(data, &r)
	if err != nil {
		return nil, fmt.Errorf("Failed to parse form-rule %s: %s", filePath, err.Error())
	}
	return r, nil
}

func (field *Field) GetFilterNames() []string {
	return field.Filters
}

func (field *Field) mergeReferenceIfNeeded(rule *Rule) {
	if field.Ref != "" {
		ref, found := rule.Fields[field.Ref]
		if found {
			if field.Name == "" {
				field.Name = ref.Name
			}
			if field.Message == "" {
				field.Message = ref.Message
			}
			field.Required = ref.Required
			field.Constraints = ref.Constraints
			field.Filters = ref.Filters
		}
	}
}

func (selection *Selection) GetFilterNames() []string {
	return selection.Filters
}

func (selection *Selection) mergeReferenceIfNeeded(rule *Rule) {
	if selection.Ref != "" {
		ref, found := rule.Selections[selection.Ref]
		if found {
			if selection.Name == "" {
				selection.Name = ref.Name
			}
			if selection.Message == "" {
				selection.Message = ref.Message
			}
			selection.Count = ref.Count
			selection.Constraints = ref.Constraints
			selection.Filters = ref.Filters
		}
	}
}

func (r *Rule) Merge(r2 *Rule) {
	for k, v := range r2.Fields {
		r.Fields[k] = v
	}
	for k, v := range r2.Selections {
		r.Selections[k] = v
	}
	for k, v := range r2.Forms {
		r.Forms[k] = v
	}
}

const defaultMaxMemory = 32 << 20

func (rule *Rule) Validate(formName string, req *http.Request) (*Result, error) {
	form, found := rule.Forms[formName]
	if !found {
		return nil, fmt.Errorf("Form rule not found '%s'", formName)
	}

	// we need to pick multiple form values from r.Form
	if req.Form == nil {
		req.ParseMultipartForm(defaultMaxMemory)
	}

	result := NewResult()

	for _, field := range form.Fields {
		field.mergeReferenceIfNeeded(rule)
		if field.Name == "" {
			return nil, fmt.Errorf("Field name not found on a rule for '%s'", formName)
		}
		value, err := filter(field, req.FormValue(field.Name))
		if err != nil {
			return nil, err
		}
		err = field.validate(result, value)
		if err != nil {
			return nil, err
		}
	}

	for _, selection := range form.Selections {
		selection.mergeReferenceIfNeeded(rule)
		if selection.Name == "" {
			return nil, errors.New("Selection name not found")
		}
		values := req.Form[selection.Name]
		filteredValues := make([]string, len(values))
		for _, value := range values {
			filteredValue, err := filter(selection, value)
			if err != nil {
				return nil, err
			}
			if filteredValue != "" {
				filteredValues = append(filteredValues, filteredValue)
			}
		}
		err := selection.validate(result, filteredValues)
		if err != nil {
			return nil, err
		}
	}

	return result, nil
}

func (field *Field) validate(result *Result, value string) error {
	if value == "" {
		if field.Required {
			result.putRequiredFailure(field.Name, field.Message)
		} else {
			result.ValidFields[field.Name] = ""
			return nil
		}
	} else {
		failure := NewFailureForField(field)
		passAll := true
		for _, constraint := range field.Constraints {
			pass, err := validate(value, constraint)
			if err != nil {
				return err
			}
			if !pass {
				passAll = false
				failure.failOnConstraint(constraint)
				if !field.FallThrough {
					break
				}
			}
		}
		if passAll {
			result.ValidFields[field.Name] = value
		} else {
			result.AddFailure(failure)
		}
	}
	return nil
}

func (selection *Selection) validate(result *Result, values []string) error {
	count := len(values)
	if count >= selection.Count.From && count <= selection.Count.To {
		if count == 0 {
			result.ValidSelections[selection.Name] = []string{}
			return nil
		}
		failure := NewFailureForSelection(selection)
		passAll := true
		for _, value := range values {
			for _, constraint := range selection.Constraints {
				pass, err := validate(value, constraint)
				if err != nil {
					return err
				}
				if !pass {
					passAll = false
					failure.failOnConstraint(constraint)
					break
				}
			}
		}
		if passAll {
			result.ValidSelections[selection.Name] = values
		} else {
			result.AddFailure(failure)
		}
	} else {
		result.putRequiredFailure(selection.Name, selection.Message)
	}
	return nil
}
