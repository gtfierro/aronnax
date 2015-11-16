package main

import (
	"github.com/satori/go.uuid"
	"testing"
)

func TestGenerateDocumentInsert(t *testing.T) {
	uuid, _ := uuid.FromString("aa45f708-8be8-11e5-86ae-5cc5d4ded1ae")
	for _, test := range []struct {
		doc    Document
		values []string
	}{
		{
			Document{UUID: uuid, Tags: map[string]string{"key1": "val1", "key2": "val2"}},
			[]string{`("aa45f708-8be8-11e5-86ae-5cc5d4ded1ae", "key1", "val1")`, `("aa45f708-8be8-11e5-86ae-5cc5d4ded1ae", "key2", "val2")`},
		},
		{
			Document{UUID: uuid, Tags: map[string]string{"key1": ""}},
			[]string{`("aa45f708-8be8-11e5-86ae-5cc5d4ded1ae", "key1", NULL)`},
		},
	} {
		generatedValues := test.doc.GenerateValues()
		var found bool
		for _, gv := range generatedValues {
			found = false
			for _, v := range test.values {
				if gv == v {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Got \n%s\nbut wanted\n%s\n", generatedValues, test.values)
			}
		}
	}
}
