package bench_test

import (
	"bytes"
	"testing"

	"github.com/santhosh-tekuri/jsonschema/v2"
	_ "github.com/santhosh-tekuri/jsonschema/v2/httploader"
)

type santhoshValidator struct {
	schema *jsonschema.Schema
}

func (s *santhoshValidator) LoadSchema(jsonSchema []byte) error {
	compiler := jsonschema.NewCompiler()

	err := compiler.AddResource("schema.json", bytes.NewBuffer(jsonSchema))
	if err != nil {
		return err
	}

	s.schema, err = compiler.Compile("schema.json")
	if err != nil {
		return err
	}

	return nil
}

func (s *santhoshValidator) ValidJSON(d []byte) bool {
	return s.schema.Validate(bytes.NewBuffer(d)) == nil
}

func (s *santhoshValidator) ValidValue(d interface{}) bool {
	return s.schema.ValidateInterface(d) == nil
}

func TestSanthoshAjv(t *testing.T) {
	dir := ajvPath
	testDir(t, dir, &santhoshValidator{})
}

func TestSanthoshDraft7(t *testing.T) {
	testDir(t, "spec/JSON-Schema-Test-Suite/tests/draft7/", &santhoshValidator{})
}

func TestSanthoshDraft7Opt(t *testing.T) {
	testDir(t, "spec/JSON-Schema-Test-Suite/tests/draft7/optional/", &santhoshValidator{})
	testDir(t, "spec/JSON-Schema-Test-Suite/tests/draft7/optional/format/", &santhoshValidator{})
}

func BenchmarkSanthoshAjv(b *testing.B) {
	dir := ajvPath
	benchDir(b, dir, &santhoshValidator{}, false)
}
