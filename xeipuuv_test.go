package bench

import (
	"encoding/json"
	"github.com/xeipuuv/gojsonschema"
	"testing"
)

type xeipuuvValidator struct {
	schema      gojsonschema.JSONLoader
	schemaValue gojsonschema.JSONLoader
}

func (s *xeipuuvValidator) LoadSchema(jsonSchema []byte) error {
	s.schema = gojsonschema.NewBytesLoader(jsonSchema)

	var schemaValue interface{}
	err := json.Unmarshal(jsonSchema, &schemaValue)
	if err != nil {
		return err
	}

	s.schemaValue = gojsonschema.NewGoLoader(schemaValue)

	return nil
}

func (s *xeipuuvValidator) ValidJSON(d []byte) bool {
	documentLoader := gojsonschema.NewBytesLoader(d)
	result, err := gojsonschema.Validate(s.schema, documentLoader)
	if err != nil {
		panic(err)
	}

	return result.Valid()
}

func (s *xeipuuvValidator) ValidValue(d interface{}) bool {
	documentLoader := gojsonschema.NewGoLoader(d)
	result, err := gojsonschema.Validate(s.schemaValue, documentLoader)
	if err != nil {
		panic(err)
	}

	return result.Valid()
}

func TestXeipuuvAjv(t *testing.T) {
	dir := "spec/ajv/spec/tests/schemas/"
	testDir(t, dir, &xeipuuvValidator{})
}

func TestXeipuuvDraft7(t *testing.T) {
	testDir(t, "spec/JSON-Schema-Test-Suite/tests/draft7/", &xeipuuvValidator{})
}

func TestXeipuuvDraft7Opt(t *testing.T) {
	testDir(t, "spec/JSON-Schema-Test-Suite/tests/draft7/optional/", &xeipuuvValidator{})
	testDir(t, "spec/JSON-Schema-Test-Suite/tests/draft7/optional/format/", &xeipuuvValidator{})
}

func BenchmarkXeipuuvAjv(b *testing.B) {
	dir := "spec/ajv/spec/tests/schemas/"
	benchDir(b, dir, &xeipuuvValidator{}, false)
}
