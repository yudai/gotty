package camelcase

import (
	"reflect"
	"testing"
)

func TestSplit(t *testing.T) {
	var testCases = []struct {
		input  string
		output []string
	}{
		{input: "", output: []string{}},
		{input: "lowercase", output: []string{"lowercase"}},
		{input: "Class", output: []string{"Class"}},
		{input: "MyClass", output: []string{"My", "Class"}},
		{input: "MyC", output: []string{"My", "C"}},
		{input: "HTML", output: []string{"HTML"}},
		{input: "PDFLoader", output: []string{"PDF", "Loader"}},
		{input: "AString", output: []string{"A", "String"}},
		{input: "SimpleXMLParser", output: []string{"Simple", "XML", "Parser"}},
		{input: "vimRPCPlugin", output: []string{"vim", "RPC", "Plugin"}},
		{input: "GL11Version", output: []string{"GL", "11", "Version"}},
		{input: "99Bottles", output: []string{"99", "Bottles"}},
		{input: "May5", output: []string{"May", "5"}},
		{input: "BFG9000", output: []string{"BFG", "9000"}},
	}

	for _, c := range testCases {
		res := Split(c.input)
		if !reflect.DeepEqual(res, c.output) {
			t.Errorf("input: '%s'\n\twant: %v\n\tgot : %v\n", c.input, c.output, res)
		}
	}
}
