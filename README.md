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
  fk "github.com/lyokato/goformkeeper"
  "github.com/martini-contrib/render"
)

func main() {

  m := martini.Classic()

  m.Use(render.Renderer())

  rule, err := fk.LoadRuleFromFile("conf/rule.yml")
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
一つ一つのconstraintのデータ構造は、`type`、`criteria`、`message`の三つのパラメータで構成されます。
`message`は、制約毎に出したい場合のみ定義すれば大丈夫です。


`criteria`は、そのconstraintを検証するに当たっての補助的な条件の指定です。
下の例ではlengthという文字の長さの制約が指定されていますが、
その補助的な条件として、0以上、20以下という条件がcriteriaにより指定されています。

```yaml
constraints:
  - type: email
  - type: length
    message: "Name length should be 0..10"
    criteria:
      from: 0
      to: 20
```

`type`の種類によっては、`criteria`が必要ないものもあります。
`criteria`に含めるパラメータは制約のtypeごとに違うものになります。

### Selection

`<select/>`や`<checkbox/>`など、複数の値を扱うコンポーネントに対してはどうすればよいでしょうか。

```html
<input type="checkbox" name="hobby[]" value="1" checked>
<label>check1</label>
<input type="checkbox" name="hobby[]" value="2" checked>
<label>check2</label>
<input type="checkbox" name="hobby[]" value="3" checked>
<label>check3</label>
```

このために`selection`というルールが使えます

```yaml
forms:
  signin:
    selection:
      - name: preference
        message: "Check You Preference"
        count:
          eq: 10
      - name: hobby
        message: "Check You Hobby"
        count:
          from: 0
          to: 10
        constraints:
          - type: length
            message: "Name length should be 0..10"
            criteria:
              from: 0
              to: 10
    fields: 
      # ...
```

`fields`のルールセットとは別に`selection`というルールセットを定義します。

`selection`として定義できるデータ構造は、基本的には`fields`で定義されたフィールド用のルールと同じですが、`required`の代わりに`count`を定義します。

「このチェックボックスでは、3つチェックされなければならない」というような条件にしたいときは、次のように`eq`を使います。


```yaml
  - name: preference
    message: "Check You Preference"
    count:
      eq: 3
```

「このチェックボックスでは、1個以上、3個以下チェックされなければならない」というような条件にしたいときは、次のように`from`と`to`を組み合わせて使います。

```yaml
  - name: hobby
    message: "Check You Hobby"
    count:
      from: 1
      to: 3
```

また、filterやconstraintsが指定されていた場合は、このcheckboxやselectなどで指定された全ての値に対して、それらを使って検証を行います。

### Reference

このように、それぞれのフォームに対してYAMLデータを定義していきますが、何度も重複する項目が出現することがあります。
たとえば`username`は、signinフォームやsignupフォームなど、様々な場所で入力を要求される可能性があります。

そのような場合にはリファレンス指定が使えます。

```yaml
fields:
  username:
    name: username
    required: true
    message: "Input Name"
    filters:
      - trim
      - uppercase
    constraints:
      - type: length
        message: "Name length should be 0..10"
        criteria:
          from: 0
          to: 10
  password:
    name: password
    required: true
    message: "Input Password"
    filters:
      - trim
      - lowercase
    constraints:
      - type: length
        message: "Password length should be 0..10"
        criteria:
          from: 0
          to: 10

forms:
  signin:
    fields: 
      - ref: username
      - ref: password
  signup:
    fields: 
      - ref: username
      - ref: password
      # and rules for other fields
```

上記のように、`forms`定義の外にfieldルールを定義しておきます。
そうすると各フォーム用のfieldルールの中から、`ref`を使って参照することが可能です。

リファレンスを使いつつ、一部の設定だけ書き換えたいときは、次のように、refに並べる形で、今まで通りパラメータを書くだけです。

```yaml
forms:
  signin:
    fields: 
      - ref: username
        name: changedName
```

### Constraints

プリセットの制約について説明していきます。

#### length

長さをチェックします。
マルチバイト文字列に対する文字数チェックは`length`ではなく、`rune_count`のほうを利用してください。

`from`,`to`で範囲指定する方法と`eq`で数を指定する方法があります。

```yaml
  - type: length
    criteria:
      eq: 10
```

```yaml
  - type: length
    criteria:
      from: 3
      to: 10
```

#### rune_count

文字数をチェックします。マルチバイトの文字は複数バイトでも一文字とカウントされます。
`length`制約と同様に、`from`,`to`で範囲指定する方法と`eq`で数を指定する方法があります。

```yaml
  - type: rune_count
    criteria:
      from: 3
      to: 10
```

```yaml
  - type: rune_count
    criteria:
      eq: 10
```

#### included

指定された複数の文字列の中に値が含まれているかを検証します

```yaml
  - type: included
    criteria:
      in: ["3", "6", "9"]
```


#### alphabet

アルファベットだけで構成されているかどうかを検証します

```yaml
  - type: alphabet
```

#### alnum

アルファベットと数値だけで構成されているかどうかを検証します

```yaml
  - type: alphabet
```

#### ascii

ASCII文字列だけで構成されているかどうかを検証します

```yaml
  - type: ascii
```
#### ascii_without_space

空白を除いたASCII文字列だけで構成されているかどうかを検証します

```yaml
  - type: ascii_without_space
```
#### regex

指定された正規表現にマッチするかを検証します

```yaml
  - type: regex
    criteria:
      regex: "^[0-9]+$"
```
#### url

URLかどうかを検証します

```yaml
  - type: url
```

#### email

Emailアドレスかどうかを検証します

```yaml
  - type: email
```

#### loose_email

Emailアドレスかどうかを検証しますが、
こちらは少し緩めのチェックになっています。

```yaml
  - type: loose_email
```

### Filters

プリセットのフィルタについて説明していきます。

#### trim

前後の空白を削除します。

#### lowercase

文字列を全て小文字に変換します。

#### uppercase

文字列を全て大文字に変換します。

### Custom Constraints

制約を自分で作る場合は以下のように、
Validatorインターフェースを実装したstructを用意し、
AddValidatorで名前をつけて定義するだけです。

Validateメソッドの中の書き方などは、goformkeeper/validators.goの中で定義されている他のValidatorがどのように書かれているかを参照するとよいでしょう。

```go
type MyValidator struct{}

func (v *MyValidator) Validate(value string, criteria *Criteria) (bool, error) {
  // ...
}

goformkeeper.AddValidator("my_constraint", &MyValidator{})
```

### Custom Filters

フィルタを自分で作る場合は以下のように
AddFilterFuncで名前を付けて、対応する関数を渡すだけです。
関数は一つの文字列を受け取り、一つの文字列を返すものでなければなりません。

```go
goformkeeper.AddFilterFunc("my_filter", MyFilterFunc)
```

実際、プリセットのフィルタは次のように定義されているだけです。

```go
AddFilterFunc("trim", strings.TrimSpace)
AddFilterFunc("lowercase", strings.ToLower)
AddFilterFunc("uppercase", strings.ToUpper)
```

## Author

Lyo Kato <lyo.kato _at_ gmail.com>

## License

Copyright (c) 2014 by Lyo Kato

MIT License

