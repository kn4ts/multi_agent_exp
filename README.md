# multi_agent_exp

## 内容物
* `server/main.go` ...サーバのメインコード
* `network_operator.go` ...ネットワーク関連のコード
* `supervisor.go` ...監視システム関連のコード
* `timer.go` ...タイマーのサブコード
* `logger.go` ...ロガーのサブコード

* `client/main.go` ...テスト用クライアントのメインコード

## 使い方
`server/`フォルダ下で以下を実行する．
```
> go run main.go
```

もしくは実行ファイルをビルドして実行する．
```
> go build
> server.exe
```

## Go言語の環境構築について

### 1. インストール
`golang インストール`で検索して環境にあったものをインストールする．

### 2. `GOPATH`の設定  
Go言語のパスは`GOROOT`と`GOPATH`の2つ．
このうち`GOPATH`は好きに設定していいようなので自分のOneDriveの直下に`go`フォルダを作ってそこに設定した．
```
C\:Users\USERNAME\OneDrive - NAME\go
```

### 3. goフォルダの構成
OneDriveに作った`go`フォルダの中身は以下のような構成にした．
```
go/
 └ src/
      └ github.com/
                  └ USERNAME/
			    ├ MyRepository1/
			    ├ MyRepository2/
			    ├ ...
			    
```

### 4. リポジトリの構成
リポジトリ内の構成は以下の通りにした．
```
MyRepositoryX/
│	   ├ myapp/	(mainパッケージ)
│	   │	 └ main.go
│	   ├ sub1.go
│	   ├ sub2.go
│	   ├ ...
│	   
├ MyRepositoryY/
├ ...

```

参考：https://future-architect.github.io/articles/20200528/

### 5. 依存関係について
上記の構成では，`main.go`は`main`パッケージとして書き，`MyRepositoryX`のパスを`import`すれば`sub1.go`や`sub2.go`内に定義した関数や構造体を利用できる．
`sub1.go`や`sub2.go`は`MyRepositoryX`パッケージとして書く．

### 6. 実行，ビルド
`myapp`フォルダ内で以下のコマンドでコードを走らせることができる．
（コード作成中はほとんどこれで対話っぽくプログラミング）
```
> go run main.go
```

また実行ファイルがほしいときは`myapp`フォルダ内で以下のコマンドで実行ファイル（デフォルトでは所属するフォルダ名と同名）をビルドできる．
```
> go build
```

### 7. その他のコマンド
以下のコマンドでデフォルトのフォーマッタを利用できる．
```
> go fmt
```

以下のコマンドでビルドされたファイル等を取り除ける．
```
> go clean
```
