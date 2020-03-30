package bson

import (
	"encoding/json"
	"testing"

	"github.com/qiniu/x/bufiox"
	"gopkg.in/mgo.v2/bson"
)

// -----------------------------------------------------------------------------

type M bson.M

func TestBson(t *testing.T) {

	doc := M{"a": 1, "b": true, "c": "Hello"}
	b, err := bson.Marshal(doc)
	if err != nil {
		t.Fatal("bson.Marshal failed:", err)
	}

	in := bufiox.NewReaderBuffer(b)
	v, err := Type.Match(in, nil)
	if err != nil {
		t.Fatal("Match failed:", err)
	}
	ret, err := json.Marshal(v)
	if err != nil {
		t.Fatal("json.Marshal failed:", err)
	}
	if string(ret) != `{"a":1,"b":true,"c":"Hello"}` {
		t.Fatal("ret:", string(ret))
	}
}

// -----------------------------------------------------------------------------
