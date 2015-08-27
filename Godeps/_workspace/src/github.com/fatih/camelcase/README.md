# CamelCase [![GoDoc](http://img.shields.io/badge/go-documentation-blue.svg?style=flat-square)](http://godoc.org/github.com/fatih/camelcase) [![Build Status](http://img.shields.io/travis/fatih/camelcase.svg?style=flat-square)](https://travis-ci.org/fatih/camelcase)

CamelCase is a Golang (Go) package to split the words of a camelcase type
string into a slice of words. It can be used to convert a camelcase word (lower
or upper case) into any type of word.

## Install

```bash
go get github.com/fatih/camelcase
```

## Usage and examples

```go
splitted := camelcase.Split("GolangPackage")

fmt.Println(splitted[0], splitted[1]) // prints: "Golang", "Package"
```

Both lower camel case and upper camel case are supported. For more info please
check: [http://en.wikipedia.org/wiki/CamelCase](http://en.wikipedia.org/wiki/CamelCase)

Below are some example cases:

```
lowercase =>       ["lowercase"]
Class =>           ["Class"]
MyClass =>         ["My", "Class"]
MyC =>             ["My", "C"]
HTML =>            ["HTML"]
PDFLoader =>       ["PDF", "Loader"]
AString =>         ["A", "String"]
SimpleXMLParser => ["Simple", "XML", "Parser"]
vimRPCPlugin =>    ["vim", "RPC", "Plugin"]
GL11Version =>     ["GL", "11", "Version"]
99Bottles =>       ["99", "Bottles"]
May5 =>            ["May", "5"]
BFG9000 =>         ["BFG", "9000"]
```
