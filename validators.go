package goformkeeper

import (
	"errors"
	"fmt"
	"net/mail"
	"net/url"
	"regexp"
	"unicode/utf8"
)

type Constraint struct {
	Type     string
	Message  string
	Criteria map[string]interface{}
}

type Criteria struct {
	values map[string]interface{}
}

type Validator interface {
	Validate(string, *Criteria) (bool, error)
}

func (c *Criteria) Has(key string) bool {
	_, found := c.values[key]
	return found
}

func (c *Criteria) Bool(key string) (bool, error) {
	value, found := c.values[key]
	if found {
		v, ok := value.(bool)
		if ok {
			return v, nil
		} else {
			return false, fmt.Errorf("Couldn't cast to bool '%s'", key)
		}
	} else {
		return false, fmt.Errorf("Param not found '%s'", key)
	}
}

func (c *Criteria) Int(key string) (int, error) {
	value, found := c.values[key]
	if found {
		v, ok := value.(int)
		if ok {
			return v, nil
		} else {
			return 0, fmt.Errorf("Couldn't cast to int '%s'", key)
		}
	} else {
		return 0, fmt.Errorf("Param not found '%s'", key)
	}
}

func (c *Criteria) String(key string) (string, error) {
	value, found := c.values[key]
	if found {
		v, ok := value.(string)
		if ok {
			return v, nil
		} else {
			return "", fmt.Errorf("Couldn't cast to string '%s'", key)
		}
	} else {
		return "", fmt.Errorf("Param not found '%s'", key)
	}
}

type RegExpValidator struct{}

func (v *RegExpValidator) Validate(value string, criteria *Criteria) (bool, error) {
	if criteria != nil && criteria.Has("regex") {
		regex, err := criteria.String("regex")
		if err != nil {
			return false, err
		}
		if m, _ := regexp.MatchString(regex, value); !m {
			return false, nil
		}
		return true, nil
	} else {
		return false, errors.New("Criteria for 'regex' not enough")
	}
}

type LooseEmailAddressValidator struct{}

// thanks to https://github.com/StefanSchroeder/Golang-Regex-Tutorial/blob/master/01-chapter3.markdown
func (v *LooseEmailAddressValidator) Validate(value string, criteria *Criteria) (bool, error) {
	if m, _ := regexp.MatchString("^(\\w[-._\\w]*\\w@\\w[-._\\w]*\\w\\.\\w{2,3})$", value); !m {
		return false, nil
	}
	return true, nil
}

type EmailAddressValidator struct{}

func (v *EmailAddressValidator) Validate(value string, criteria *Criteria) (bool, error) {
	_, err := mail.ParseAddress(value)
	if err != nil {
		return false, nil
	}
	return true, nil
}

type URLValidator struct{}

func (v *URLValidator) Validate(value string, criteria *Criteria) (bool, error) {
	_, err := url.ParseRequestURI(value)
	if err != nil {
		return false, nil
	}
	return true, nil
}

type AlphabetValidator struct{}

func (v *AlphabetValidator) Validate(value string, criteria *Criteria) (bool, error) {
	if m, _ := regexp.MatchString("^[a-zA-Z]+$", value); !m {
		return false, nil
	}
	return true, nil
}

type AlphabetAndNumberValidator struct{}

func (v *AlphabetAndNumberValidator) Validate(value string, criteria *Criteria) (bool, error) {
	if m, _ := regexp.MatchString("^[0-9a-zA-Z]+$", value); !m {
		return false, nil
	}
	return true, nil
}

type AsciiValidator struct{}

func (v *AsciiValidator) Validate(value string, criteria *Criteria) (bool, error) {
	if m, _ := regexp.MatchString("^[\\x20-\\x7E]+$", value); !m {
		return false, nil
	}
	return true, nil
}

type AsciiWithoutSpaceValidator struct{}

func (v *AsciiWithoutSpaceValidator) Validate(value string, criteria *Criteria) (bool, error) {
	if m, _ := regexp.MatchString("^[\\x21-\\x7E]+$", value); !m {
		return false, nil
	}
	return true, nil
}

type RuneCountValidator struct{}

func (v *RuneCountValidator) Validate(value string, criteria *Criteria) (bool, error) {
	if criteria == nil {
		return false, errors.New("Criteria for 'rune_count' not enough")
	}
	if criteria.Has("eq") {
		eq, err := criteria.Int("eq")
		if err != nil {
			return false, err
		}
		return utf8.RuneCount([]byte(value)) == eq, nil
	} else if criteria.Has("to") && criteria.Has("from") {
		to, err := criteria.Int("to")
		if err != nil {
			return false, err
		}
		from, err := criteria.Int("from")
		if err != nil {
			return false, err
		}
		count := utf8.RuneCount([]byte(value))
		return count >= from && count <= to, nil
	} else {
		return false, errors.New("Criteria for 'rune_count' not enough")
	}
}

type LengthValidator struct{}

func (v *LengthValidator) Validate(value string, criteria *Criteria) (bool, error) {
	if criteria == nil {
		return false, errors.New("Criteria for 'length' not enough")
	}
	if criteria.Has("eq") {
		eq, err := criteria.Int("eq")
		if err != nil {
			return false, err
		}
		return len(value) == eq, nil
	} else if criteria.Has("to") && criteria.Has("from") {
		to, err := criteria.Int("to")
		if err != nil {
			return false, err
		}
		from, err := criteria.Int("from")
		if err != nil {
			return false, err
		}
		return len(value) >= from && len(value) <= to, nil
	} else {
		return false, errors.New("Criteria for 'length' not enough")
	}
}

var validators map[string]Validator

func init() {
	validators = make(map[string]Validator)
	AddValidator("length", &LengthValidator{})
	AddValidator("rune_count", &RuneCountValidator{})
	AddValidator("alphabet", &AlphabetValidator{})
	AddValidator("alnum", &AlphabetAndNumberValidator{})
	AddValidator("ascii", &AsciiValidator{})
	AddValidator("ascii_without_space", &AsciiWithoutSpaceValidator{})
	AddValidator("regex", &RegExpValidator{})
	AddValidator("url", &URLValidator{})
	AddValidator("email", &EmailAddressValidator{})
	AddValidator("loose_email", &LooseEmailAddressValidator{})
}

func AddValidator(name string, validator Validator) {
	validators[name] = validator
}

func validate(value string, constraint *Constraint) (bool, error) {
	validator, found := validators[constraint.Type]
	if !found {
		return false, fmt.Errorf("Validator not found: %s", constraint.Type)
	}
	criteria := &Criteria{constraint.Criteria}
	ok, err := validator.Validate(value, criteria)
	return ok, err
}
