# GoFormKeeper

GoFormKeeper provides you a easy way to validate form parameters in Golang.

## Installation

This library depends on "gopkg.in/yaml.v1"
So, go get this package beforehand

```
go get gopkg.in/yaml.v1
go get github.com/lyokato/goformkeeper
```

## TODO

ADDS MORE TEST

Translation to English

## Getting Started

あなたのWebアプリケーションに次のようなformがあり、
このformに対するユーザーの入力をチェックしたいとします。

```html
<form action="/signin" method="POST">
  <input type="text" name="email" />
  <input type="password" name="password" />
  <button type="submit">Sign in</button>
</form>
```

次のように、ruleを定義したYAMLファイルを用意します。

下の例では、ルールのセットに`signin`という名前を付けて、
`fields`以下に、各フィールドのvalidationのルールを定義してあるのが
なんとなく分かるでしょうか。

ルールの定義の仕方について、詳しくは後で説明します。

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

Webアプリケーションは、このように用意されたルールファイルを読み込み、
そこから生成された`Rule`オブジェクトを利用して、HTTP requestをチェックします。

以下はmartiniとpongo2を使った簡単なサンプルです。

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

  // Display Input Form
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

    // valid input

    email    := results.ValidParam("email")
    password := results.ValidParam("password")
  }

```


まず冒頭部分で事前に定義されたルールファイルを読み込んでいます。
ここで、`Rule`オブジェクトを作成しています。ファイルが存在しなかったり、
ファイルのフォーマットに問題があって`Rule`オブジェクトが生成できなかった場合は、
errが返ります。

```go
rule, err := goformkeeper.LoadRuleFromFile("conf/rule.yml")
```

次に、Postメソッドに注目して下さい。
```go
results, err := rule.Validate("signin", req)
if err != nil {
  // プログラム内部の問題、デベロッパーが直すべき
}

if results.HasFailure() {
  // ユーザー入力値の問題、エンドユーザーにメッセージを表示して修正を求める
}
```

ルールファイルの中で定義された`signin`のルールセットに従って
HTTP requestをチェックし、その結果を`Results`オブジェクトとして返します。
ユーザーの入力値が、定義されたルールにそぐわなければ
`Results`オブジェクトの`HasFailure`メソッドがtrueを返します。

ここで、戻り値のerrではなく、`HasFailure`を使って分岐をしている点に注意してください。
ここでerrがnilでは無い場合、そのerrが表すのは、ユーザーの入力による問題ではなくプログラム内部の問題です。
例えば指定された`signin`というルールが存在しない、などの場合にエラーと判断されます。
プログラム内部の問題はデベロッパーが修正すべき問題ですので、エンドユーザーによる不正入力値とは処理を分けます。

ユーザーが、あらかじめ指定されたruleに違反する入力を行ったかどうかは
`HasFailure`でチェックします。

エラーがなかった場合は入力値に問題なかったと判断し、
処理を進めますが、その祭に、resultsオブジェクトの
`ValidParam`メソッドを利用して以下のように、検証済みの値を取得できます。

```go
email    := results.ValidParam("email")
password := results.ValidParam("password")

// myApp.Login(email, password)
```

元の`http.Request`から値を直接取得するのとどう違うのかというと、
ルールで、filterが指定されいた場合、`ValidParam`で取得できる値は
フィルタ済みの値になります。

例えばtrim, lowercase, uppercaseというようなフィルタを指定することが可能です。フィルター機能については、詳しくは別の頁で説明をします。

また、このメソッドを通すことで、検証済みの値であることが保証されます。



次にHTML Templateの生成部分を見てみましょう
この例ではpongo2を利用していますので、以下のように
templateにparameterを渡しています。

`Results`オブジェクトをパラメータとして渡しています。

```go
err = tpl.ExecuteWriter(pongo2.Context{
  "form":  results,
  "title": "Hello, World!",
}, res)
```

GoFormKeeperは、エラーメッセージのハンドリング機能を備えています。
以下のようにすれば、ルールに外れた入力が行われたフィールドに関する
メッセージのリストが表示されます。
ここで表示されるメッセージ文字列は、Rule fileで定義されたものです。

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

上の例では、`Messages`メソッドを利用して、メッセージをまとめてリスト表示しましたが、
次のように、`FailedOn`, `MessageOn`メソッドを使って、invalidな入力が行われたフィールドの
それぞれのコンポーネントの側に、別々にエラーメッセージを添えることもできます。

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

`FailedOnConstraint`, `MessageOnConstraint`を使えば、
どの種類の検証に失敗したかをチェックすることが可能です。
例えば、「入力値の長さに問題にあった場合」や「入力値が数字でなかった場合など」、
検証の種類により、細かく処理を分けることも可能です

```html
<form action="/signin" method="POST">

{% if form.FailedOnConstraint("email", "length") %}
<p>INVALID: {{ form.MessageOnConstraint("email", "length") }}</p>
{% endif %}

{% if form.FailedOnConstraint("email", "email") %}
<p>INVALID: {{ form.MessageOnConstraint("email", "email") }}</p>
{% endif %}

<input type="text" name="email" /><br />

</form>
```

## Rule File Format

## Constraints

## Filters
