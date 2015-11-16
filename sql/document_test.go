package main

import (
	"github.com/satori/go.uuid"
	"testing"
)

func TestGenerateDocumentInsert(t *testing.T) {
	uuid, _ := uuid.FromString("aa45f708-8be8-11e5-86ae-5cc5d4ded1ae")
	for _, test := range []struct {
		doc    Document
		insert string
	}{
		{
			Document{UUID: uuid, Tags: map[string]string{"key1": "val1", "key2": "val2"}},
			`INSERT INTO data (uuid, dkey, dval) VALUES ("aa45f708-8be8-11e5-86ae-5cc5d4ded1ae", "key1", "val1"),("aa45f708-8be8-11e5-86ae-5cc5d4ded1ae", "key2", "val2");`,
		},
	} {
		if generated := test.doc.GenerateinsertStatement(); generated != test.insert {
			t.Errorf("Got \n%s\nbut wanted\n%s\n", generated, test.insert)
		}
	}
}
