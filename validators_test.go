package goformkeeper

import (
	"testing"
)

func TestLooseEmail(t *testing.T) {

	v := LooseEmailAddressValidator{}

	if ok, _ := v.Validate("example+test@example.com", &Criteria{}); !ok {
		t.Errorf("allow + in the email")
	}

}
