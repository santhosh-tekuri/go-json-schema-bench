package bench

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"testing"
)

func init() {
	go func() {
		err := http.ListenAndServe(":1234", http.FileServer(http.Dir("./spec/JSON-Schema-Test-Suite/remotes/")))
		if err != nil {
			log.Fatal(err)
		}
	}()
}

type testCase struct {
	Description string          `json:"description"`
	Schema      json.RawMessage `json:"schema"`
	Tests       []test          `json:"tests"`
}

type test struct {
	Description string          `json:"description"`
	Data        json.RawMessage `json:"data"`
	Valid       bool            `json:"valid"`
}

type validator interface {
	LoadSchema([]byte) error
	ValidJSON([]byte) bool
	ValidValue(interface{}) bool
}

func testDir(t *testing.T, dir string, v validator) {
	t.Helper()

	files, err := ioutil.ReadDir(dir)
	require.NoError(t, err)

	for _, f := range files {
		if f.IsDir() {
			continue
		}

		j, err := ioutil.ReadFile(dir + f.Name())
		require.NoError(t, err)

		var testCases []testCase
		require.NoError(t, json.Unmarshal(j, &testCases))

		for _, testCase := range testCases {
			assert.NotPanics(t, func() {
				err = v.LoadSchema(testCase.Schema)
			})

			if !assert.NoError(t, err) {
				continue
			}

			for _, test := range testCase.Tests {
				t.Run(f.Name()+":"+testCase.Description+":"+test.Description, func(t *testing.T) {
					assert.NotPanics(t, func() {
						if test.Valid != v.ValidJSON(test.Data) {
							t.Fail()
						}
					})
				})
			}
		}
	}
}

func benchDir(b *testing.B, dir string, v validator, validateValue bool) {
	b.Helper()

	files, err := ioutil.ReadDir(dir)
	require.NoError(b, err)

	for _, f := range files {
		j, err := ioutil.ReadFile(dir + f.Name())
		require.NoError(b, err)

		var testCases []testCase
		require.NoError(b, json.Unmarshal(j, &testCases))

		for _, testCase := range testCases {
			err := v.LoadSchema(testCase.Schema)
			assert.NoError(b, err)

			for _, test := range testCase.Tests {
				b.Run(f.Name()+":"+testCase.Description+":"+test.Description, func(b *testing.B) {
					if validateValue {
						var val interface{}
						err := json.Unmarshal(test.Data, &val)
						if err != nil {
							b.FailNow()
						}

						b.ReportAllocs()
						b.ResetTimer()

						for i := 0; i < b.N; i++ {
							if test.Valid != v.ValidValue(val) {
								b.FailNow()
							}
						}
					} else {
						b.ReportAllocs()
						b.ResetTimer()

						for i := 0; i < b.N; i++ {
							if test.Valid != v.ValidJSON(test.Data) {
								b.FailNow()
							}
						}
					}
				})
			}
		}
	}
}

func validatorFromEnv() validator {
	var v validator
	switch os.Getenv("VALIDATOR") {
	case "santhosh":
		v = &santhoshValidator{}
	case "qri":
		v = &qriValidator{}
	case "xeipuuv":
		v = &xeipuuvValidator{}
	default:
		panic("unknown or missing VALIDATOR env var")
	}

	return v
}

func BenchmarkAjv(b *testing.B) {
	dir := "spec/ajv/spec/tests/schemas/"
	benchDir(b, dir, validatorFromEnv(), false)
}

func BenchmarkAjvVal(b *testing.B) {
	dir := "spec/ajv/spec/tests/schemas/"
	benchDir(b, dir, validatorFromEnv(), true)
}

func BenchmarkDraft7(b *testing.B) {
	dir := "spec/JSON-Schema-Test-Suite/tests/draft7/"
	benchDir(b, dir, validatorFromEnv(), false)
}
