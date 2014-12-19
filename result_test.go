package goformkeeper

import (
	"testing"
)

func TestResultRequired(t *testing.T) {

	r := NewResult()

	if r.HasFailure() {
		t.Errorf("Result shouldn't have failure")
	}

	r.putRequiredFailure("Field01", "Field01 Is Empty")

	if !r.HasFailure() {
		t.Errorf("Result should have failure")
	}

	if !r.FailedOn("Field01") {
		t.Errorf("FailedOn(\"Field01\") should return true")
	}

	if !r.FailedOnConstraint("Field01", "required") {
		t.Errorf("FailedOnConstraint(\"Field01\", \"required\") should return true")
	}

	if r.MessageOn("Field01") != "Field01 Is Empty" {
		t.Errorf("MessageOn(\"required\") returns invalid value %s", r.MessageOn("Field01"))
	}

	if r.MessageOnConstraint("Field01", "required") != "Field01 Is Empty" {
		t.Errorf("MessageOnConstraint(\"Field01\", \"required\") returns invalid value %s", r.MessageOnConstraint("Field01", "required"))
	}
}

func TestResult2(t *testing.T) {

	r := NewResult()

	if r.HasFailure() {
		t.Errorf("Result shouldn't have failure")
	}

	f01 := NewFailureForField("Field02", "Field02 Has Error")
	f01.failOnConstraint("length", "Field02 Length is Invalid")
	r.AddFailure(f01)

	if !r.HasFailure() {
		t.Errorf("Result should have failure")
	}

	if r.FailedOn("Field01") {
		t.Errorf("FailedOn(\"Field01\") should return false")
	}

	if !r.FailedOn("Field02") {
		t.Errorf("FailedOn(\"Field02\") should return true")
	}

	if !r.FailedOnConstraint("Field02", "length") {
		t.Errorf("FailedOnConstraint(\"Field02\", \"required\") should return true")
	}

	if r.MessageOn("Field02") != "Field02 Has Error" {
		t.Errorf("MessageOn(\"Field02\") returns invalid value %s", r.MessageOn("Field02"))
	}

	if r.MessageOnConstraint("Field02", "length") != "Field02 Length is Invalid" {
		t.Errorf("MessageOnConstraint(\"Field02\", \"length\") returns invalid value %s", r.MessageOnConstraint("Field02", "length"))
	}

	f01.failOnConstraint("email", "Field02 is not Email Address")

	f02 := NewFailureForField("Field03", "Field03 Has Error")
	f02.failOnConstraint("length", "Field03 Length is Invalid")
	r.AddFailure(f02)

	m := r.Messages()

	if len(m) != 2 {
		t.Errorf("Messages() returns invalid number")
	}

	if m[0] != "Field02 Has Error" {
		t.Errorf("first error message returns wrong string: %s", m[0])
	}

	if m[1] != "Field03 Has Error" {
		t.Errorf("second error message returns wrong string: %s", m[1])
	}

	m2 := r.MessagesOn("Field02")

	if len(m2) != 2 {
		t.Errorf("MessagesOn(\"Field02\") returns invalid number")
	}

	if m2[0] != "Field02 Length is Invalid" {
		t.Errorf("first error message returns wrong string: %s", m2[0])
	}

	if m2[1] != "Field02 is not Email Address" {
		t.Errorf("second error message returns wrong string: %s", m2[1])
	}

	errorFields := r.FailedFields()
	if len(errorFields) != 2 {
		t.Errorf("FailedFields() returns invalid number")
	}

	if errorFields[0] != "Field02" {
		t.Errorf("FailedFields() returns wrong field: %s", errorFields[0])
	}

	if errorFields[1] != "Field03" {
		t.Errorf("FailedFields() returns wrong field: %s", errorFields[1])
	}
}
