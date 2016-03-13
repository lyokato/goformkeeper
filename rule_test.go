package goformkeeper

import (
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"testing"

	"github.com/kr/pretty"
)

func TestLoadRule(t *testing.T) {
	path := "./tests/rules.yml"
	dir, _ := os.Getwd()
	path = filepath.Join(dir, path)

	rule, err := LoadRuleFromFile(path)
	if err != nil {
		t.Errorf("Failed to load rule %s", err.Error())
	}

	if len(rule.Forms) == 2 {
		t.Errorf("RULE:", pretty.Formatter(rule))
		//t.Errorf("Form: got %v\nwant %v", conf.Template.Path, expectedTemplatePath)
	}

	req := &http.Request{Method: "GET"}
	url, _ := url.Parse("http://www.example.org/?username=foobar&password=foobarfoobar")
	req.URL = url

	result, err := rule.Validate("signin", req)
	if err != nil {
		t.Errorf("Failed to validate: %s", err.Error())
		return
	}

	if result.ValidParam("other") != "default" {
		t.Errorf("RESULT:", pretty.Formatter(result.Messages()))
		t.Errorf("Failed validation: want %s, got %s", "default", result.ValidParam("other"))
	}

	if result.ValidParam("username") != "FOOBAR" {
		t.Errorf("RESULT:", pretty.Formatter(result.Messages()))
		t.Errorf("Failed validation: want %s, got %s", "FOOBAR", result.ValidParam("username"))
	}

	if !result.FailedOnConstraint("password", "length") {
		t.Errorf("RESULT:", pretty.Formatter(result.Messages()))
	}

	req2 := &http.Request{Method: "GET"}
	url2, _ := url.Parse("http://www.example.org/?username=foobar&password=foobarfoobar&choise=2")
	req2.URL = url2

	result2, err := rule.Validate("signin", req2)
	if err != nil {
		t.Errorf("found error: %s", err.Error())
		return
	}
	if !result2.FailedOnConstraint("choise", "included") {
		t.Errorf("RESULT:", pretty.Formatter(result2.Messages()))
	}

	req3 := &http.Request{Method: "GET"}
	url3, _ := url.Parse("http://www.example.org/?username=foobar&password=foobarfoobar&choise=3")
	req3.URL = url3

	result3, err := rule.Validate("signin", req3)
	if err != nil {
		t.Errorf("Failed to validate: %s", err.Error())
		return
	}

	if result3.ValidParam("choise") != "3" {
		t.Errorf("RESULT: %v", result)
		return
	}

}
