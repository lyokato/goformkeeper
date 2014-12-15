# GoFormKeeper

GoFormKeeper provides you a easy way to validate form parameters in Golang.
(This library is not stable version yet)

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


### Input Form

あなたのWebアプリケーションに次のようなformがあり、
このformに対するユーザーの入力をチェックしたいとします。

次の例では、signinの処理のために、
ユーザーに対してemailとpasswordの入力を要求しています。
サーバー側では、これらの値が適切に入力されたか検証する必要があります。

```html
<form action="/signin" method="POST">
  <input type="text" name="email" />
  <input type="password" name="password" />
  <button type="submit">Sign in</button>
</form>
```

### Rule File

次のように、ruleを定義したYAMLファイルを用意します。

下の例では、ルールのセットに`signin`という名前を付けて、
`fields`以下に、各フィールドの検証ルールを定義してあるのが
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

### Application Example

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

#### Loading Rule File

まず冒頭部分で事前に定義されたルールファイルを読み込んでいます。
ここで、`Rule`オブジェクトを作成しています。ファイルが存在しなかったり、
ファイルのフォーマットに問題があって`Rule`オブジェクトが生成できなかった場合は、
errが返ります。

```go
rule, err := goformkeeper.LoadRuleFromFile("conf/rule.yml")
```

#### Validation

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

#### Valid Parameter

```go
email    := results.ValidParam("email")
password := results.ValidParam("password")

// myApp.Login(email, password)
```

元の`http.Request`から値を直接取得するのとどう違うのかというと、
このフィールドにfilter ルールが指定されいた場合、`ValidParam`で取得できる値は
フィルタ済みの値になります。

例えば`trim`, `lowercase`, `uppercase`というようなフィルタを指定することが可能です。
フィルター機能については、詳しくは別の頁で説明をします。

また、このメソッドを通すことで、検証済みの値であることが保証されます。


#### Error Message Handling

次にHTML Templateの生成部分を見てみましょう。
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
どの制約の検証に失敗したかをチェックすることが可能です。
例えば、「入力値の長さに問題にあった場合」や「入力値が数字でなかった場合など」、
制約の種類により、細かく処理を分けることも可能です

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

Ruleファイルの書き方を説明します。
上の例で使ったファイルをもう一度見てみましょう。

signinに使うルールが定義されています。

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

signinに使うルールが定義されています。
singinではなく、ユーザー登録によるsignup用のフォームも追加したくなったとします。

以下のようにsignupのrule setを追加します。

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
  signup:
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
      - name: username
        required: true
        message: "Input Username correctly"
        constraints:
        - type: length
          message: "password length should be 5 - 20"
          criteria:
            from: 5
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

このように、formごとにルールセットを追加していきます。


作成したルールファイルは次のように読み込む事ができます。

```go
rule, err := goformkeeper.LoadRuleFromFile("conf/rule.yml")
```

ただし、このままルールを増やしていくとruleファイルのサイズが膨大になっていき、メンテナンスがしにくくなっていくでしょう。
そのような場合はルールファイルを複数に分けて書くことを推奨します。


例えばsignin.ymlとsignup.ymlに分離します。

conf/rule/signin.yml
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

conf/rule/signup.yml
```yaml
forms:
  signup:
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
      - name: username
        required: true
        message: "Input Username correctly"
        constraints:
        - type: length
          message: "password length should be 5 - 20"
          criteria:
            from: 5
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

このようにルールファイルを分割する場合は、
`LoadRuleFromFile`ではなく、`LoadRuleFromDir`を使います。
ディレクトリ内の全てのRuleファイルを読み込み、
統合された一つの`Rule`オブジェクトを作ることができます。

```go
rule, err := goformkeeper.LoadRuleFromDir("conf/rule")
```

### Fields

で次に`fields`以下の設定を見ていきます
次のようなデータ構造で作られたルールを、検証が必要なフィールド毎に用意してリストにします。

```yaml
fields:
  - name: email
    required: true
    message: "Input email address correctly"
    filters:
      - trim
      - lowercase
    constraints:
      - type: email
      - type: length
        criteria:
          from: 0
          to: 20
```

このデータ構造は次のパラメータで構成されます。

#### name

必須パラメータです。
この名前はHTML上のformの中の検証したいコンポーネントに付けた名前と同じにして下さい

#### required

この値をtrueにした場合、そのフィールドパラメータが存在しなかったり、空文字列だった場合に検証失敗と判断します。
この値がfalseであった場合は、値が空であっても、以降のconstraintsの検証をスキップし、検証成功と同じ扱いにします。
フィールドパラメータが空でなかった場合は、通常の処理として指定されたconstraintsによる検証を順次行います。

#### message

このフィールドで検証失敗した場合に、ユーザーに表示したいメッセージ文字列を定義します。
メッセージは制約ごとに分けて書く事も可能ですが、フィールド毎に一つのメッセージで十分な場合はここで定義します。

#### filters

このフィールドに対して処理をかけたいフィルターをリストアップします。
フィルターはまず最初に実行され、constraintの検証は、フィルターされた結果に対して行われます。

#### constraints

ここにconstraintをリストアップしていきます。
一つ一つのconstraintのデータ構造は以下のように、`type`、`criteria`、`message`の三つのパラメータで構成されます。
`type`の種類によっては、`criteria`が必要ないものもあります。
`criteria`に含めるパラメータは制約のtypeごとに違うものになります。
messageは、制約毎に出したい場合のみ定義すれば大丈夫です。

```yaml
constraints:
  - type: email
  - type: length
    message: "Name length should be 0..10"
    criteria:
      from: 0
      to: 20
```

### Constraints

#### length
#### rune_count
#### alphabet
#### alnum
#### ascii
#### ascii_without_space
#### regex
#### url
#### email
#### loose_email

### Filters

#### trim
#### lowercase
#### uppercase

### Reference
### Selection

## Author

Lyo Kato <lyo.kato _at_ gmail.com>

## License

Copyright (c) 2014 by Lyo Kato

MIT License

