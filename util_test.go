package goformkeeper

import (
	"testing"
)

func TestUniqueString(t *testing.T) {
	arr := []string{"foo", "foo", "bar", "buz"}
	results := Uniq(arr)
	if len(results) != 3 {
		t.Errorf("Uniq returns invalid number of results")
	}
	if results[0] != "bar" {
		t.Errorf("Uniq 1st results is wrong %s", results[0])
	}
	if results[1] != "buz" {
		t.Errorf("Uniq 2nd results is wrong %s", results[1])
	}
	if results[2] != "foo" {
		t.Errorf("Uniq 3rd results is wrong %s", results[2])
	}
}
