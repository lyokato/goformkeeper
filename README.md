# GoFormKeeper

GoFormKeeper provides you a easy way to validate form parameters in Golang.

# Dependency

This library depends on "gopkg.in/yaml.v1"

## TODO

ADDS MORE TEST

### Getting Started

```html
<form action="/signin" method="POST">
  <input type="text" name="email" />
  <input type="password" name="password" />
  <button type="submit'>Sign in</button>
</form>
```

Create 'rules.yml'

```yaml
forms:
  signin:
    fields:
      - name: email
        required: true
        message: "Input email address correctly"
        constraints:
        - type: email
        - type: length
          criteria:
            from: 0
            to: 20
      - name: password
        required: true
        message: "Input password correctly"
        constraints:
        - type: length
          message: "password length should be 5 - 20"
          criteria:
            from: 5
            to: 20
```

Load rule file with goformkeeper.LoadRuleFromFile(filePath)
or you can loadd multiple rule files once with goformkeeper.LoadRuleFromDir(dirPath)
Then you can get goformkeeper.Rule object.

Next, you should call rule.Validate(ruleName, request).
It returns formkeeper.Result object.


Simple example with martini and pongo2

```go
package main

import (
  "fmt"
  "net/http"
  "github.com/flosch/pongo2"
  "github.com/go-martini/martini"
  "github.com/lyokato/goformkeeper"
  "github.com/martini-contrib/render"
)

func main() {

  m := martini.Classic()

  m.Use(render.Renderer())

  rule, err := goformkeeper.LoadRuleFromFile("conf/rule.yml")
  if err != nil {
    fmt.Println(err)
    return
  }

  m.Get("/", func(res http.ResponseWriter, req *http.Request, render render.Render) {

    tpl, err := pongo2.FromFile("templates/index.html")
    if err != nil {
      http.Error(res, err.Error(), http.StatusInternalServerError)
      return
    }

    err = tpl.ExecuteWriter(pongo2.Context{
      "title": "Hello World!",
    }, res)
    if err != nil {
      http.Error(res, err.Error(), http.StatusInternalServerError)
    }
  })

  m.Post("/signin", func(res http.ResponseWriter, req *http.Request, render render.Render) {

    results, err := rule.Validate("signin", req)
    if results.HasFailure() {
      tpl, err := pongo2.FromFile("templates/index.html")
      if err != nil {
        http.Error(res, err.Error(), http.StatusInternalServerError)
        return
      }
      err = tpl.ExecuteWriter(pongo2.Context{
        "form":  results,
        "title": "Hello, World!",
      }, res)
      if err != nil {
        http.Error(res, err.Error(), http.StatusInternalServerError)
      }
      return
    }
  }

```

You can put "error messages block" on your html

```html
{% if form.HasFailure %}
<p>Error Found</p>
<ul>
  {% for msg in form.Messages %}
  <li>{{ msg }}</li>
  {% endfor %}
</ul>
{% endif %}
```

Or you also can set messages for each form-element

```html
<form action="/signin" method="POST">

{% if form.FailedOn("email") %}
<p>INVALID: {{ form.MessageOn("email") }}</p>
{% endif %}

<input type="text" name="email" /><br />

{% if form.FailedOn("password") %}
<p>INVALID: {{ form.MessageOn("password") }}</p>
{% endif %}

<input type="password" name="password" /><br />

<button type="submit">Sign in</button><br />
</form>
```
