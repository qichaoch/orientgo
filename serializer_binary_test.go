package orient

import (
	"bytes"
	"encoding/base64"
	"math/big"
	"reflect"
	"testing"
	"time"

	"github.com/qichaoch/orientgo/obinary/rw"
)

func TestDeserializeRecordData(t *testing.T) {
	data, err := base64.StdEncoding.DecodeString(`AAASY2FyZXRha2VyAAAAJQcIbmFtZQAAAC0HBmFnZQAAADMBAA5NaWNoYWVsCkxpbnVzHg==`)
	if err != nil {
		t.Fatal(err)
	}

	rec := NewEmptyDocument()
	rec.SetSerializer(&BinaryRecordFormat{})
	rec.Fill(NewEmptyRID(), 0, data)

	if doc, err := rec.ToDocument(); err != nil {
		t.Fatal(err)
	} else if len(doc.Fields()) != 3 {
		t.Fatal("wrong fields count in document")
	} else if doc.GetField("caretaker").Value.(string) != "Michael" ||
		doc.GetField("name").Value.(string) != "Linus" ||
		doc.GetField("age").Value.(int32) != 15 {
		t.Fatal("wrong values in document: ", doc)
	}
}

func testBase64Compare(t *testing.T, out []byte, origBase64 string) {
	orig, _ := base64.StdEncoding.DecodeString(origBase64)
	if bytes.Compare(out, orig) != 0 {
		t.Fatalf("different buffers:\n%v\n%v\n", out, orig)
	}
}

func TestSerializeCommandNoParams(t *testing.T) {
	query := "SELECT FROM V WHERE Id = ?"
	buf := bytes.NewBuffer(nil)
	if err := NewSQLCommand(query).ToStream(buf); err != nil {
		t.Fatal(err)
	}
	testBase64Compare(t, buf.Bytes(), "AAAAGlNFTEVDVCBGUk9NIFYgV0hFUkUgSWQgPSA/AAA=")
}

func TestSerializeCommandIntParam(t *testing.T) {
	query := "SELECT FROM V WHERE Id = ?"
	buf := bytes.NewBuffer(nil)
	if err := NewSQLCommand(query, int32(25)).ToStream(buf); err != nil {
		t.Fatal(err)
	}
	testBase64Compare(t, buf.Bytes(), "AAAAGlNFTEVDVCBGUk9NIFYgV0hFUkUgSWQgPSA/AQAAAB0AABRwYXJhbWV0ZXJzAAAAEwwAAgcCMAAAABwBMgA=")
}

func testSerializeEmbMap(t *testing.T, off int, mp interface{}, origBase64 string) {
	buf := bytes.NewBuffer(nil)
	for i := 0; i < off; i++ {
		buf.WriteByte(0)
	}
	if err := (binaryRecordFormatV0{}).writeEmbeddedMap(rw.NewWriter(buf), off, mp); err != nil {
		t.Fatal(err)
	}
	testBase64Compare(t, buf.Bytes(), origBase64)
	r := rw.NewReadSeeker(bytes.NewReader(buf.Bytes()))
	out, err := (binaryRecordFormatV0{}).readEmbeddedMap(r, nil)
	if err != nil {
		t.Fatal(err)
	}
	if reflect.TypeOf(out) != reflect.TypeOf(mp) {
		t.Logf("types are not the same: %T -> %T", mp, out)
	}
}

func TestSerializeEmbeddedMapInt32V0(t *testing.T) {
	testSerializeEmbMap(t, 0,
		map[int32]interface{}{int32(0): int32(25)},
		"AgcCMAAAAAkBMg==",
	)
}

func TestSerializeEmbeddedMapIntV0(t *testing.T) {
	testSerializeEmbMap(t, 0,
		map[int]interface{}{0: 25},
		"AgcCMAAAAAkDMg==",
	)
}

func TestSerializeEmbeddedMapIntOffsV0(t *testing.T) {
	testSerializeEmbMap(t, 4,
		map[int]interface{}{0: 25},
		"AAAAAAIHAjAAAAANAzI=",
	)
}

func TestSerializeEmbeddedMapStringV0(t *testing.T) {
	testSerializeEmbMap(t, 0,
		map[string]interface{}{"one": "two"},
		"AgcGb25lAAAACwcGdHdv",
	)
}

func TestSerializeEmbeddedMapEmptyV0(t *testing.T) {
	testSerializeEmbMap(t, 0,
		map[string]interface{}{},
		"AA==",
	)
}

func testSerializeEmbCol(t *testing.T, off int, col interface{}, origBase64 string) {
	buf := bytes.NewBuffer(nil)
	for i := 0; i < off; i++ {
		buf.WriteByte(0)
	}
	if err := (binaryRecordFormatV0{}).writeEmbeddedCollection(rw.NewWriter(buf), off, col, UNKNOWN); err != nil {
		t.Fatal(err)
	}
	testBase64Compare(t, buf.Bytes(), origBase64)
}

func TestSerializeEmbeddedColStringV0(t *testing.T) {
	testSerializeEmbCol(t, 0,
		[]string{"a", "b", "c"},
		"BhcHAmEHAmIHAmM=",
	)
}

func TestSerializeEmbeddedColStringOffsV0(t *testing.T) {
	testSerializeEmbCol(t, 4,
		[]string{"a", "b", "c"},
		"AAAAAAYXBwJhBwJiBwJj",
	)
}

func testSerializeDoc(t *testing.T, doc *Document, origBase64 string) {
	buf := bytes.NewBuffer(nil)
	GetDefaultRecordSerializer().ToStream(buf, doc)
	testBase64Compare(t, buf.Bytes(), origBase64)
}

func TestSerializeDocumentEmpty(t *testing.T) {
	doc := NewEmptyDocument()
	doc.SetField("parameters", map[string]interface{}{})
	testSerializeDoc(t,
		doc,
		"AAAUcGFyYW1ldGVycwAAABMMAAA=",
	)
}

func TestSerializeDocumentFieldStringMap(t *testing.T) {
	doc := NewEmptyDocument()
	doc.SetField("parameters", map[string]string{"one": "two"})
	testSerializeDoc(t,
		doc,
		"AAAUcGFyYW1ldGVycwAAABMMAAIHBm9uZQAAAB4HBnR3bw==",
	)
}

func TestSerializeDocumentFieldMapAndArr(t *testing.T) {
	doc := NewEmptyDocument()
	doc.SetField("map", map[string]string{"one": "two"})
	doc.SetField("arr", []string{"a", "b", "c"})
	testSerializeDoc(t,
		doc,
		"AAAGbWFwAAAAFQwGYXJyAAAAJAoAAgcGb25lAAAAIAcGdHdvBhcHAmEHAmIHAmM=",
	)
}

func TestSerializeDecimalV0(t *testing.T) {
	buf := bytes.NewBuffer(nil)
	val := big.NewInt(123456789)
	if err := (binaryRecordFormatV0{}).writeSingleValue(rw.NewWriter(buf), 0, val, DECIMAL, UNKNOWN); err != nil {
		t.Fatal(err)
	}
	testBase64Compare(t, buf.Bytes(), "AAAAAAAAAAQHW80V")

	r := rw.NewReadSeeker(bytes.NewReader(buf.Bytes()))
	out, err := (binaryRecordFormatV0{}).readSingleValue(r, DECIMAL, nil)
	if err != nil {
		t.Fatal(err)
	}
	if val2, ok := out.(Decimal); !ok {
		t.Fatalf("expected Decimal, got: %T", out)
	} else if val.Cmp(val2.Value) != 0 {
		t.Fatalf("values differs: %v != %v", val, val2)
	}
}

func TestSerializeDatetimeV0(t *testing.T) {
	buf := bytes.NewBuffer(nil)
	val := time.Now()
	val = time.Unix(val.Unix(), int64(val.Nanosecond()/1e6)*1e6) // precise to milliseconds
	if err := (binaryRecordFormatV0{}).writeSingleValue(rw.NewWriter(buf), 0, val, DATETIME, UNKNOWN); err != nil {
		t.Fatal(err)
	}

	r := rw.NewReadSeeker(bytes.NewReader(buf.Bytes()))
	out, err := (binaryRecordFormatV0{}).readSingleValue(r, DATETIME, nil)
	if err != nil {
		t.Fatal(err)
	}
	if val2, ok := out.(time.Time); !ok {
		t.Fatalf("expected Time, got: %T", out)
	} else if !val.Equal(val2) {
		t.Fatalf("values differs: %v != %v", val, val2)
	}
}

func testDocumentToStruct(t *testing.T, dataBase64 string) {
	data, err := base64.StdEncoding.DecodeString(dataBase64)
	if err != nil {
		t.Fatal(err)
	}
	doc := NewEmptyDocument()
	err = doc.Fill(NewEmptyRID(), 0, data)
	if err != nil {
		t.Fatal(err)
	}

	type Inner struct {
		Name string
	}
	type Item struct {
		One   Inner
		Inner []Inner
	}

	one, two := Inner{Name: "one"}, Inner{Name: "two"}

	var obj *Item
	if err = doc.ToStruct(&obj); err != nil {
		t.Fatal(err)
	} else if obj.One != one {
		t.Fatal("item is wrong")
	} else if len(obj.Inner) != 2 || obj.Inner[0] != one || obj.Inner[1] != two {
		t.Fatal("list is wrong")
	}
}

func TestDocumentInnerStruct(t *testing.T) {
	testDocumentToStruct(t, "AAJWBk9uZQAAABgJCklubmVyAAAAKAoAAAhOYW1lAAAAJAcABm9uZQQXCQAITmFtZQAAADcHAAZvbmUJAAhOYW1lAAAASAcABnR3bw==")
}

func TestDocumentInnerMapToStruct(t *testing.T) {
	testDocumentToStruct(t, "AAJWBk9uZQAAABgMCklubmVyAAAAKAoAAgcITmFtZQAAACQHBm9uZQQXDAIHCE5hbWUAAAA3BwZvbmUMAgcITmFtZQAAAEgHBnR3bw==")
}
