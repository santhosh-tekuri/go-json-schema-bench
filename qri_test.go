package bench

import (
	"context"
	"encoding/json"
	"github.com/qri-io/jsonschema"
	"testing"
)

type qriValidator struct {
	schema *jsonschema.Schema
}

func (s *qriValidator) LoadSchema(jsonSchema []byte) error {
	schema := jsonschema.Schema{}

	err := json.Unmarshal(jsonSchema, &schema)
	if err != nil {
		return err
	}

	s.schema = &schema

	return nil
}

func (s *qriValidator) ValidJSON(d []byte) bool {
	ke, err := s.schema.ValidateBytes(context.Background(), d)
	if err != nil {
		panic(err)
	}

	return len(ke) == 0
}

func (s *qriValidator) ValidValue(d interface{}) bool {
	ke := s.schema.Validate(context.Background(), d)

	return len(*ke.Errs) == 0
}

func TestQriAjv(t *testing.T) {
	dir := "spec/ajv/spec/tests/schemas/"
	testDir(t, dir, &qriValidator{})
}

func TestQriDraft7(t *testing.T) {
	testDir(t, "spec/JSON-Schema-Test-Suite/tests/draft7/", &qriValidator{})
}

func TestQriDraft7Opt(t *testing.T) {
	testDir(t, "spec/JSON-Schema-Test-Suite/tests/draft7/optional/", &qriValidator{})
	testDir(t, "spec/JSON-Schema-Test-Suite/tests/draft7/optional/format/", &qriValidator{})
}

func BenchmarkQriAjv(b *testing.B) {
	dir := "spec/ajv/spec/tests/schemas/"
	benchDir(b, dir, &qriValidator{}, false)
}
