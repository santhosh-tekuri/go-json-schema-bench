# Benchmarking correctness and performance of Go JSON schema validators

[JSON schema](https://json-schema.org/) is a popular technology to describe structured data in a programming language agnostic way. 

One of the most useful applications of JSON schema is data validation. 

Constraints provided by schema are stateless and allow predictable and high performance check of data conformance.

For example, given this schema:
```json
{
    "type": "object",
    "properties": {
        "foo": {
            "type": "string",
            "minLength": 5
        }
    }
}
```

This value would be valid:
```json
{"foo":"abcdef"}
```

And this would be not:
```json
{"foo":"abc"}
```

As [implementations](https://json-schema.org/implementations.html#validator-go) list suggests, these libraries can validate JSON schema in Go:
* https://github.com/xeipuuv/gojsonschema `v1.2.0`
* https://github.com/santhosh-tekuri/jsonschema `v2.2.0`
* https://github.com/qri-io/jsonschema `v0.2.0`

Let's see how good are they.

Great thing about JSON schema is that a [test suite](https://github.com/json-schema-org/JSON-Schema-Test-Suite/tree/master/tests/) is available to check the correctness of implementation. Some test cases are optional, it means they may fail and user should not rely on their behavior unless implementation specifically declares such support. Test cases in `format.json` are also considered as best effort implementation and may fail.

Another source of test cases is [`ajv-validator`](https://github.com/ajv-validator/ajv/tree/master/spec/tests/schemas) - one of the best implementations available across all languages.

## Correctness

[This repo](https://github.com/swaggest/go-json-schema-bench) includes test case sources as git submodules and implements test case adapters for validators.

[Test report](https://github.com/swaggest/go-json-schema-bench/actions?query=workflow%3Atest)

```
santhosh-tekuri/jsonschema failed tests count
draft 7          | 1
draft 7 format   | 90
draft 7 optional | 48
ajv              | 0
qri-io/jsonschema failed tests count
draft 7          | 51
draft 7 format   | 0
draft 7 optional | 37
ajv              | 6
xeipuuv/gojsonschema failed tests count
draft 7          | 0
draft 7 format   | 32
draft 7 optional | 49
ajv              | 0
```

We can see that both `santhosh-tekuri/jsonschema` and `xeipuuv/gojsonschema` are good in general schema compliance.

The [failing edge case](https://github.com/json-schema-org/JSON-Schema-Test-Suite/blob/fadab68f9c153d2b6c8b6a75487fdc265db1fb42/tests/draft7/type.json#L11-L15) for `santhosh-tekuri/jsonschema` was recently added to JSON Schema Test Suite:

```json
            {
                "description": "a float with zero fractional part is an integer",
                "data": 1.0,
                "valid": true
            }
```

Unfortunately `qri-io/jsonschema` fails quite a few tests in general suite and also in ajv suite. The correctness of this library is questionable. Though this library has remarkable results in validating `format`.

## Performance

Tests that were used for correctness can also be used to benchmark performance.

Where possible schema is initialized before benchmark, so that every iteration performs only validation. Benchmark result is then aggregated with [benchstat](https://pkg.go.dev/golang.org/x/perf/cmd/benchstat).

[Performance Benchmark Reports](https://github.com/swaggest/go-json-schema-bench/actions?query=workflow%3Abench).

### AJV Suite Time/Op

Let's start with `ajv` suite, it contains complex schemas that especially fit for performance benchmarks.

```
name \ time/op    santhosh-ajv.txt  qri-ajv.txt   xeipuuv-ajv.txt
[Geo mean]        13.0µs            9.0µs         207.3µs
```

This result shows how much time does it take to validate a value. Geo mean is aggregated from a full result table. 

The value for `qri` is not very relevant because validation failed for several slow tests and did not contribute to mean value. Please check full report below to compare performance of `qri-io/jsonschema` with its rivals.

Results for `santhosh-ajv.txt` and `xeipuuv-ajv.txt` are comparable since both have successfully passed a full benchmark.

Lower performance of `xeipuuv/gojsonschema` is caused by unfortunate design of the validator that prevents schemas from being reused. For every iteration schema has to be unmarshaled. In contrast both `santhosh-tekuri/jsonschema` and `qri-io/jsonschema` validate using a "compiled" (unmarshaled) schema instance which makes more sense in long-running Go application where validation generally happens much more often than schema initialization.

<details><summary>full benchstat report</summary>
```
name \ time/op                                                                                                                          santhosh-ajv.txt  qri-ajv.txt   xeipuuv-ajv.txt
Ajv/advanced.json:advanced_schema_from_z-schema_benchmark_(https://github.com/zaggino/z-schema):valid_object_from_z-schema_benchmark-2       89.0µs ± 1%                   609.1µs ± 1%
Ajv/advanced.json:advanced_schema_from_z-schema_benchmark_(https://github.com/zaggino/z-schema):not_object-2                                 1.64µs ± 2%   2.33µs ± 2%    399.79µs ± 1%
Ajv/advanced.json:advanced_schema_from_z-schema_benchmark_(https://github.com/zaggino/z-schema):root_only_is_valid-2                         21.6µs ± 1%                   454.2µs ± 1%
Ajv/advanced.json:advanced_schema_from_z-schema_benchmark_(https://github.com/zaggino/z-schema):missing_root_entry-2                         8.10µs ± 1%   7.94µs ± 1%    419.72µs ± 1%
Ajv/advanced.json:advanced_schema_from_z-schema_benchmark_(https://github.com/zaggino/z-schema):invalid_entry_key-2                          27.7µs ± 1%   16.1µs ± 1%     464.4µs ± 0%
Ajv/advanced.json:advanced_schema_from_z-schema_benchmark_(https://github.com/zaggino/z-schema):missing_storage_in_entry-2                   6.34µs ± 1%   8.71µs ± 1%    414.92µs ± 1%
Ajv/advanced.json:advanced_schema_from_z-schema_benchmark_(https://github.com/zaggino/z-schema):missing_storage_type-2                       26.6µs ± 0%   10.4µs ± 1%     446.3µs ± 1%
Ajv/advanced.json:advanced_schema_from_z-schema_benchmark_(https://github.com/zaggino/z-schema):storage_type_should_be_a_string-2            30.4µs ± 1%   10.8µs ± 1%     458.3µs ± 1%
Ajv/advanced.json:advanced_schema_from_z-schema_benchmark_(https://github.com/zaggino/z-schema):storage_device_should_match_pattern-2        32.5µs ± 1%   11.1µs ± 2%     469.2µs ± 2%
Ajv/basic.json:basic_schema_from_z-schema_benchmark_(https://github.com/zaggino/z-schema):valid_array_from_z-schema_benchmark-2              29.5µs ± 1%   41.6µs ± 1%      97.3µs ± 1%
Ajv/basic.json:basic_schema_from_z-schema_benchmark_(https://github.com/zaggino/z-schema):not_array-2                                        1.61µs ± 0%   2.30µs ± 1%     55.48µs ± 1%
Ajv/basic.json:basic_schema_from_z-schema_benchmark_(https://github.com/zaggino/z-schema):array_of_not_onjects-2                             4.04µs ± 1%   6.02µs ± 1%     61.82µs ± 1%
Ajv/basic.json:basic_schema_from_z-schema_benchmark_(https://github.com/zaggino/z-schema):missing_required_properties-2                      3.00µs ±12%   4.73µs ± 0%     58.19µs ± 1%
Ajv/basic.json:basic_schema_from_z-schema_benchmark_(https://github.com/zaggino/z-schema):required_property_of_wrong_type-2                  4.52µs ± 2%   7.83µs ± 1%     59.88µs ± 2%
Ajv/basic.json:basic_schema_from_z-schema_benchmark_(https://github.com/zaggino/z-schema):smallest_valid_product-2                           3.88µs ± 1%   7.04µs ± 1%     58.09µs ± 2%
Ajv/basic.json:basic_schema_from_z-schema_benchmark_(https://github.com/zaggino/z-schema):tags_should_be_array-2                             5.53µs ± 1%   9.11µs ± 1%     61.97µs ± 1%
Ajv/basic.json:basic_schema_from_z-schema_benchmark_(https://github.com/zaggino/z-schema):dimensions_should_be_object-2                      5.64µs ± 2%   9.04µs ± 1%     61.33µs ± 1%
Ajv/basic.json:basic_schema_from_z-schema_benchmark_(https://github.com/zaggino/z-schema):valid_product_with_tag-2                           4.65µs ± 2%   9.13µs ± 0%     61.25µs ± 2%
Ajv/basic.json:basic_schema_from_z-schema_benchmark_(https://github.com/zaggino/z-schema):dimensions_miss_required_properties-2              7.49µs ± 2%  12.75µs ± 1%     69.05µs ± 2%
Ajv/basic.json:basic_schema_from_z-schema_benchmark_(https://github.com/zaggino/z-schema):valid_product_with_tag_and_dimensions-2            8.47µs ± 2%  14.64µs ± 1%     67.15µs ± 1%
Ajv/complex.json:complex_schema_from_jsck_benchmark_(https://github.com/pandastrike/jsck):valid_array_from_jsck_benchmark-2                   254µs ± 0%                     732µs ± 1%
Ajv/complex.json:complex_schema_from_jsck_benchmark_(https://github.com/pandastrike/jsck):not_array-2                                        1.62µs ± 2%   2.27µs ± 1%    399.45µs ± 1%
Ajv/complex2.json:complex_schema_from_jsck_benchmark_without_IDs_in_definitions:valid_array_from_jsck_benchmark-2                             257µs ± 1%                     721µs ± 1%
Ajv/complex2.json:complex_schema_from_jsck_benchmark_without_IDs_in_definitions:not_array-2                                                  1.63µs ± 1%   2.26µs ± 1%    386.57µs ± 0%
Ajv/complex2.json:complex_schema_from_jsck_benchmark_without_IDs_in_definitions:one_valid_item-2                                             75.2µs ± 1%                   502.1µs ± 2%
Ajv/complex2.json:complex_schema_from_jsck_benchmark_without_IDs_in_definitions:one_invalid_item-2                                           81.4µs ± 1%   32.9µs ± 1%     510.1µs ± 1%
Ajv/complex3.json:complex_schema_from_jsck_benchmark_(https://github.com/pandastrike/jsck):valid_array_from_jsck_benchmark-2                  253µs ± 1%                     822µs ± 1%
Ajv/complex3.json:complex_schema_from_jsck_benchmark_(https://github.com/pandastrike/jsck):not_array-2                                       1.64µs ± 1%   2.40µs ± 1%    495.68µs ± 1%
Ajv/cosmicrealms.json:schema_from_cosmicrealms_benchmark:valid_data_from_cosmicrealms_benchmark-2                                            44.9µs ± 2%   54.0µs ± 2%     165.4µs ± 2%
Ajv/cosmicrealms.json:schema_from_cosmicrealms_benchmark:invalid_data-2                                                                      68.8µs ± 1%   59.6µs ± 1%     224.4µs ± 2%
Ajv/medium.json:medium_schema_from_jsck_benchmark_(https://github.com/pandastrike/jsck):valid_object_from_jsck_benchmark-2                   17.2µs ± 2%   33.2µs ± 1%      92.9µs ± 1%
Ajv/medium.json:medium_schema_from_jsck_benchmark_(https://github.com/pandastrike/jsck):not_object-2                                         1.61µs ± 1%   2.34µs ± 1%     74.07µs ± 1%
[Geo mean]                                                                                                                                   13.0µs         9.0µs          207.3µs
```
</details>

### Draft 7 Time/Op

```
name \ time/op   santhosh-draft7.txt  qri-draft7.txt  xeipuuv-draft7.txt
[Geo mean]       2.23µs               3.43µs          11.54µs     
```

Test cases in draft 7 suite are smaller and less complex, so perfomance is better.

<details><summary>full benchstat report</summary>
```
name \ time/op                                                                                                                                       santhosh-draft7.txt  qri-draft7.txt  xeipuuv-draft7.txt
Draft7/additionalItems.json:additionalItems_as_schema:additional_items_match_schema-2                                                                        4.70µs ± 2%     5.12µs ± 3%        12.50µs ± 0%
Draft7/additionalItems.json:additionalItems_as_schema:additional_items_do_not_match_schema-2                                                                 4.39µs ± 1%     5.67µs ± 1%        14.61µs ± 1%
Draft7/additionalItems.json:items_is_schema,_no_additionalItems:all_items_match_schema-2                                                                     4.56µs ± 1%     4.66µs ± 1%        12.87µs ± 1%
Draft7/additionalItems.json:array_of_items_with_no_additionalItems:empty_array-2                                                                              850ns ± 1%     1979ns ± 1%         8120ns ± 0%
Draft7/additionalItems.json:array_of_items_with_no_additionalItems:fewer_number_of_items_present_(1)-2                                                       1.69µs ± 1%     2.57µs ± 1%         9.76µs ± 1%
Draft7/additionalItems.json:array_of_items_with_no_additionalItems:fewer_number_of_items_present_(2)-2                                                       2.47µs ± 2%     3.21µs ± 1%        11.28µs ± 2%
Draft7/additionalItems.json:array_of_items_with_no_additionalItems:equal_number_of_items_present-2                                                           3.22µs ± 1%     3.80µs ± 1%        13.07µs ±13%
Draft7/additionalItems.json:array_of_items_with_no_additionalItems:additional_items_are_not_permitted-2                                                      3.91µs ± 2%     4.25µs ± 1%        13.87µs ± 1%
Draft7/additionalItems.json:additionalItems_as_false_without_items:items_defaults_to_empty_schema_so_everything_is_valid-2                                   1.80µs ± 1%     2.88µs ± 0%         5.47µs ± 1%
Draft7/additionalItems.json:additionalItems_as_false_without_items:ignores_non-arrays-2                                                                      1.31µs ± 1%     2.17µs ± 1%         4.76µs ± 1%
Draft7/additionalItems.json:additionalItems_are_allowed_by_default:only_the_first_item_is_validated-2                                                        2.63µs ± 2%     3.06µs ± 0%         8.44µs ± 0%
Draft7/additionalItems.json:additionalItems_should_not_look_in_applicators,_valid_case:items_defined_in_allOf_are_not_examined-2                             2.38µs ± 1%                        12.80µs ± 1%
Draft7/additionalItems.json:additionalItems_should_not_look_in_applicators,_invalid_case:items_defined_in_allOf_are_not_examined-2                           4.14µs ± 0%                        22.13µs ± 1%
Draft7/additionalProperties.json:additionalProperties_being_false_does_not_allow_other_properties:no_additional_properties_is_valid-2                        2.42µs ± 2%     4.11µs ± 2%        15.57µs ± 1%
Draft7/additionalProperties.json:additionalProperties_being_false_does_not_allow_other_properties:an_additional_property_is_invalid-2                        4.72µs ± 0%     5.90µs ± 0%        23.38µs ± 1%
Draft7/additionalProperties.json:additionalProperties_being_false_does_not_allow_other_properties:ignores_arrays-2                                           1.41µs ± 2%     2.48µs ± 0%        12.40µs ± 1%
Draft7/additionalProperties.json:additionalProperties_being_false_does_not_allow_other_properties:ignores_strings-2                                          1.25µs ± 1%     1.84µs ± 1%        12.41µs ± 1%
Draft7/additionalProperties.json:additionalProperties_being_false_does_not_allow_other_properties:ignores_other_non-objects-2                                1.73µs ± 0%     1.81µs ± 1%        13.39µs ± 1%
Draft7/additionalProperties.json:additionalProperties_being_false_does_not_allow_other_properties:patternProperties_are_not_additional_properties-2          3.51µs ± 1%     5.39µs ± 0%        18.86µs ± 1%
Draft7/additionalProperties.json:non-ASCII_pattern_with_additionalProperties:matching_the_pattern_is_valid-2                                                 2.48µs ± 1%     3.99µs ± 1%        12.92µs ± 1%
Draft7/additionalProperties.json:non-ASCII_pattern_with_additionalProperties:not_matching_the_pattern_is_invalid-2                                           2.59µs ± 1%     3.22µs ± 1%        13.45µs ± 0%
Draft7/additionalProperties.json:additionalProperties_allows_a_schema_which_should_validate:no_additional_properties_is_valid-2                              2.15µs ± 0%     3.79µs ± 1%        11.09µs ± 1%
Draft7/additionalProperties.json:additionalProperties_allows_a_schema_which_should_validate:an_additional_valid_property_is_valid-2                          3.49µs ± 1%     5.45µs ± 1%        13.33µs ± 0%
Draft7/additionalProperties.json:additionalProperties_allows_a_schema_which_should_validate:an_additional_invalid_property_is_invalid-2                      4.39µs ± 1%     6.27µs ± 1%        16.52µs ± 1%
Draft7/additionalProperties.json:additionalProperties_can_exist_by_itself:an_additional_valid_property_is_valid-2                                            1.47µs ± 0%     3.27µs ± 0%         7.09µs ± 1%
Draft7/additionalProperties.json:additionalProperties_can_exist_by_itself:an_additional_invalid_property_is_invalid-2                                        2.26µs ± 1%     3.99µs ± 0%        10.25µs ± 1%
Draft7/additionalProperties.json:additionalProperties_are_allowed_by_default:additional_properties_are_allowed-2                                             3.01µs ± 1%     4.34µs ± 1%        10.51µs ± 1%
Draft7/additionalProperties.json:additionalProperties_should_not_look_in_applicators:properties_defined_in_allOf_are_not_examined-2                          3.40µs ± 0%                        16.61µs ± 0%
Draft7/allOf.json:allOf:allOf-2                                                                                                                              3.05µs ± 1%     7.19µs ± 1%        19.70µs ± 0%
Draft7/allOf.json:allOf:mismatch_second-2                                                                                                                    2.89µs ± 1%     5.94µs ± 0%        21.15µs ± 1%
Draft7/allOf.json:allOf:mismatch_first-2                                                                                                                     3.95µs ± 1%     5.96µs ± 2%        22.22µs ± 1%
Draft7/allOf.json:allOf:wrong_type-2                                                                                                                         3.52µs ± 1%     7.78µs ± 1%        22.49µs ± 1%
Draft7/allOf.json:allOf_with_base_schema:valid-2                                                                                                             3.41µs ± 1%     8.64µs ± 1%        24.19µs ± 1%
Draft7/allOf.json:allOf_with_base_schema:mismatch_base_schema-2                                                                                              2.71µs ± 1%     8.04µs ± 1%        24.41µs ± 1%
Draft7/allOf.json:allOf_with_base_schema:mismatch_first_allOf-2                                                                                              4.42µs ± 2%     7.63µs ± 1%        26.61µs ± 0%
Draft7/allOf.json:allOf_with_base_schema:mismatch_second_allOf-2                                                                                             4.57µs ± 2%     7.68µs ± 1%        26.76µs ± 1%
Draft7/allOf.json:allOf_with_base_schema:mismatch_both-2                                                                                                     5.64µs ± 1%     6.52µs ± 0%        27.86µs ± 1%
Draft7/allOf.json:allOf_simple_types:valid-2                                                                                                                 2.94µs ± 1%     3.32µs ± 1%        13.68µs ± 1%
Draft7/allOf.json:allOf_simple_types:mismatch_one-2                                                                                                          6.43µs ± 1%     4.21µs ± 2%        22.75µs ± 1%
Draft7/allOf.json:allOf_with_boolean_schemas,_all_true:any_value_is_valid-2                                                                                  1.21µs ± 1%     3.10µs ± 1%         6.16µs ± 1%
Draft7/allOf.json:allOf_with_boolean_schemas,_some_false:any_value_is_invalid-2                                                                              1.95µs ± 2%     3.54µs ± 1%         8.52µs ± 2%
Draft7/allOf.json:allOf_with_boolean_schemas,_all_false:any_value_is_invalid-2                                                                               3.08µs ± 1%     3.95µs ± 1%         9.56µs ± 3%
Draft7/allOf.json:allOf_with_one_empty_schema:any_data_is_valid-2                                                                                            2.27µs ± 1%     2.53µs ± 0%         8.51µs ± 1%
Draft7/allOf.json:allOf_with_two_empty_schemas:any_data_is_valid-2                                                                                           2.81µs ± 1%     3.22µs ± 1%        10.72µs ± 1%
Draft7/allOf.json:allOf_with_the_first_empty_schema:number_is_valid-2                                                                                        2.84µs ± 1%     3.25µs ± 1%        11.82µs ± 1%
Draft7/allOf.json:allOf_with_the_first_empty_schema:string_is_invalid-2                                                                                      2.29µs ± 0%     3.89µs ± 1%        12.78µs ± 1%
Draft7/allOf.json:allOf_with_the_last_empty_schema:number_is_valid-2                                                                                         2.82µs ± 2%     3.27µs ± 2%        11.76µs ± 0%
Draft7/allOf.json:allOf_with_the_last_empty_schema:string_is_invalid-2                                                                                       2.31µs ± 0%     3.85µs ± 0%        13.03µs ± 1%
Draft7/allOf.json:nested_allOf,_to_check_validation_semantics:null_is_valid-2                                                                                1.19µs ± 1%     3.39µs ± 1%        10.91µs ± 1%
Draft7/allOf.json:nested_allOf,_to_check_validation_semantics:anything_non-null_is_invalid-2                                                                 3.99µs ± 1%     4.22µs ± 1%        19.05µs ± 3%
Draft7/allOf.json:allOf_combined_with_anyOf,_oneOf:allOf:_false,_anyOf:_false,_oneOf:_false-2                                                                16.1µs ± 1%      6.6µs ± 2%         46.2µs ± 1%
Draft7/allOf.json:allOf_combined_with_anyOf,_oneOf:allOf:_false,_anyOf:_false,_oneOf:_true-2                                                                 12.2µs ± 2%      5.8µs ± 0%         37.1µs ± 1%
Draft7/allOf.json:allOf_combined_with_anyOf,_oneOf:allOf:_false,_anyOf:_true,_oneOf:_false-2                                                                 12.1µs ± 1%      5.8µs ± 2%         37.2µs ± 1%
Draft7/allOf.json:allOf_combined_with_anyOf,_oneOf:allOf:_false,_anyOf:_true,_oneOf:_true-2                                                                  8.08µs ± 1%     5.00µs ± 1%        28.66µs ± 1%
Draft7/allOf.json:allOf_combined_with_anyOf,_oneOf:allOf:_true,_anyOf:_false,_oneOf:_false-2                                                                 12.4µs ± 2%      5.8µs ± 1%         37.3µs ± 1%
Draft7/allOf.json:allOf_combined_with_anyOf,_oneOf:allOf:_true,_anyOf:_false,_oneOf:_true-2                                                                  8.29µs ± 1%     5.02µs ± 1%        29.10µs ± 2%
Draft7/allOf.json:allOf_combined_with_anyOf,_oneOf:allOf:_true,_anyOf:_true,_oneOf:_false-2                                                                  8.03µs ± 1%     4.94µs ± 1%        29.15µs ± 2%
Draft7/allOf.json:allOf_combined_with_anyOf,_oneOf:allOf:_true,_anyOf:_true,_oneOf:_true-2                                                                   4.16µs ± 1%     4.17µs ± 1%        18.99µs ± 1%
Draft7/anyOf.json:anyOf:first_anyOf_valid-2                                                                                                                  2.70µs ± 1%     2.39µs ± 1%        12.85µs ± 0%
Draft7/anyOf.json:anyOf:second_anyOf_valid-2                                                                                                                 4.03µs ± 0%     3.50µs ± 2%        17.42µs ± 0%
Draft7/anyOf.json:anyOf:both_anyOf_valid-2                                                                                                                   2.70µs ± 1%     2.37µs ± 2%        12.70µs ± 0%
Draft7/anyOf.json:anyOf:neither_anyOf_valid-2                                                                                                                8.30µs ± 2%     4.27µs ± 1%        27.08µs ± 1%
Draft7/anyOf.json:anyOf_with_base_schema:mismatch_base_schema-2                                                                                              1.62µs ± 1%     2.94µs ± 2%        14.46µs ± 1%
Draft7/anyOf.json:anyOf_with_base_schema:one_anyOf_valid-2                                                                                                   1.96µs ± 1%     3.59µs ± 1%        14.50µs ± 1%
Draft7/anyOf.json:anyOf_with_base_schema:both_anyOf_invalid-2                                                                                                3.23µs ± 1%     4.16µs ± 1%        17.75µs ± 0%
Draft7/anyOf.json:anyOf_with_boolean_schemas,_all_true:any_value_is_valid-2                                                                                  1.21µs ± 1%     2.25µs ± 1%         6.13µs ± 1%
Draft7/anyOf.json:anyOf_with_boolean_schemas,_some_true:any_value_is_valid-2                                                                                 1.21µs ± 1%     2.26µs ± 1%         6.03µs ± 1%
Draft7/anyOf.json:anyOf_with_boolean_schemas,_all_false:any_value_is_invalid-2                                                                               2.55µs ± 1%     3.45µs ± 1%         9.61µs ± 0%
Draft7/anyOf.json:anyOf_complex_types:first_anyOf_valid_(complex)-2                                                                                          2.58µs ± 1%     4.25µs ± 0%        18.87µs ± 1%
Draft7/anyOf.json:anyOf_complex_types:second_anyOf_valid_(complex)-2                                                                                         2.39µs ± 1%     5.68µs ± 1%        20.26µs ± 1%
Draft7/anyOf.json:anyOf_complex_types:both_anyOf_valid_(complex)-2                                                                                           2.86µs ± 3%     4.60µs ± 1%        19.21µs ± 2%
Draft7/anyOf.json:anyOf_complex_types:neither_anyOf_valid_(complex)-2                                                                                        4.72µs ± 1%     7.67µs ± 1%        26.04µs ± 2%
Draft7/anyOf.json:anyOf_with_one_empty_schema:string_is_valid-2                                                                                              1.80µs ± 1%     3.44µs ± 1%        11.69µs ± 1%
Draft7/anyOf.json:anyOf_with_one_empty_schema:number_is_valid-2                                                                                              2.39µs ± 1%     2.42µs ± 1%        11.03µs ± 1%
Draft7/anyOf.json:nested_anyOf,_to_check_validation_semantics:null_is_valid-2                                                                                1.18µs ± 1%     2.92µs ± 1%        10.86µs ± 1%
Draft7/anyOf.json:nested_anyOf,_to_check_validation_semantics:anything_non-null_is_invalid-2                                                                 4.27µs ± 0%     3.74µs ± 2%        19.19µs ± 1%
Draft7/anyOf.json:nested_anyOf,_to_check_validation_semantics:null_is_valid#01-2                                                                             1.19µs ± 2%     2.91µs ± 1%        10.93µs ± 1%
Draft7/anyOf.json:nested_anyOf,_to_check_validation_semantics:anything_non-null_is_invalid#01-2                                                              4.28µs ± 2%     3.74µs ± 1%        19.20µs ± 1%
Draft7/boolean_schema.json:boolean_schema_'true':number_is_valid-2                                                                                           1.12µs ± 1%     1.68µs ± 1%         3.68µs ± 1%
Draft7/boolean_schema.json:boolean_schema_'true':string_is_valid-2                                                                                           1.17µs ± 1%     1.68µs ± 1%         3.68µs ± 1%
Draft7/boolean_schema.json:boolean_schema_'true':boolean_true_is_valid-2                                                                                     1.09µs ± 2%     1.60µs ± 1%         3.58µs ± 0%
Draft7/boolean_schema.json:boolean_schema_'true':boolean_false_is_valid-2                                                                                    1.08µs ± 0%     1.58µs ± 0%         3.64µs ± 1%
Draft7/boolean_schema.json:boolean_schema_'true':null_is_valid-2                                                                                             1.10µs ± 2%     1.59µs ± 1%         3.67µs ± 1%
Draft7/boolean_schema.json:boolean_schema_'true':object_is_valid-2                                                                                           1.26µs ± 1%     2.09µs ± 1%         3.84µs ± 2%
Draft7/boolean_schema.json:boolean_schema_'true':empty_object_is_valid-2                                                                                      811ns ± 1%     1677ns ± 1%         3393ns ± 1%
Draft7/boolean_schema.json:boolean_schema_'true':array_is_valid-2                                                                                            1.03µs ± 0%     1.88µs ± 1%         3.58µs ± 0%
Draft7/boolean_schema.json:boolean_schema_'true':empty_array_is_valid-2                                                                                       815ns ± 1%     1701ns ± 2%         3289ns ± 1%
Draft7/boolean_schema.json:boolean_schema_'false':number_is_invalid-2                                                                                        1.33µs ± 1%     1.95µs ± 1%         4.77µs ± 1%
Draft7/boolean_schema.json:boolean_schema_'false':string_is_invalid-2                                                                                        1.38µs ± 1%     1.95µs ± 1%         4.85µs ± 0%
Draft7/boolean_schema.json:boolean_schema_'false':boolean_true_is_invalid-2                                                                                  1.31µs ± 1%     1.86µs ± 1%         4.72µs ± 1%
Draft7/boolean_schema.json:boolean_schema_'false':boolean_false_is_invalid-2                                                                                 1.32µs ± 4%     1.86µs ± 1%         4.72µs ± 1%
Draft7/boolean_schema.json:boolean_schema_'false':null_is_invalid-2                                                                                          1.33µs ± 1%     1.86µs ± 1%         4.69µs ± 1%
Draft7/boolean_schema.json:boolean_schema_'false':object_is_invalid-2                                                                                        1.46µs ± 1%     2.38µs ± 1%         4.94µs ± 1%
Draft7/boolean_schema.json:boolean_schema_'false':empty_object_is_invalid-2                                                                                  1.01µs ± 1%     1.99µs ± 1%         4.48µs ± 2%
Draft7/boolean_schema.json:boolean_schema_'false':array_is_invalid-2                                                                                         1.21µs ± 0%     2.20µs ± 2%         4.72µs ± 0%
Draft7/boolean_schema.json:boolean_schema_'false':empty_array_is_invalid-2                                                                                   1.01µs ± 1%     1.99µs ± 1%         4.47µs ± 1%
Draft7/const.json:const_validation:same_value_is_valid-2                                                                                                     2.21µs ± 1%     2.28µs ± 1%         8.31µs ± 2%
Draft7/const.json:const_validation:another_value_is_invalid-2                                                                                                2.71µs ± 1%     3.00µs ± 1%        10.91µs ± 1%
Draft7/const.json:const_validation:another_type_is_invalid-2                                                                                                 1.68µs ± 1%     2.94µs ± 1%         9.78µs ± 1%
Draft7/const.json:const_with_object:same_object_is_valid-2                                                                                                   1.81µs ± 1%     5.09µs ± 1%        15.92µs ± 1%
Draft7/const.json:const_with_object:same_object_with_different_property_order_is_valid-2                                                                     1.81µs ± 2%     5.15µs ± 2%        15.66µs ± 2%
Draft7/const.json:const_with_object:another_object_is_invalid-2                                                                                              1.65µs ± 1%     5.93µs ± 1%        16.17µs ± 1%
Draft7/const.json:const_with_object:another_type_is_invalid-2                                                                                                1.61µs ± 2%     5.65µs ± 1%        15.74µs ± 1%
Draft7/const.json:const_with_array:same_array_is_valid-2                                                                                                     1.74µs ± 1%     4.86µs ± 1%        13.49µs ± 1%
Draft7/const.json:const_with_array:another_array_item_is_invalid-2                                                                                           1.39µs ± 0%     5.21µs ± 3%        13.61µs ± 1%
Draft7/const.json:const_with_array:array_with_additional_items_is_invalid-2                                                                                  1.80µs ± 1%     5.48µs ± 1%        15.17µs ± 1%
Draft7/const.json:const_with_null:null_is_valid-2                                                                                                            1.15µs ± 1%     2.05µs ± 2%         6.33µs ± 0%
Draft7/const.json:const_with_null:not_null_is_invalid-2                                                                                                      2.02µs ± 1%     2.71µs ± 1%        10.00µs ± 1%
Draft7/const.json:const_with_false_does_not_match_0:false_is_valid-2                                                                                         1.15µs ± 1%     2.10µs ± 1%         6.72µs ± 0%
Draft7/const.json:const_with_false_does_not_match_0:integer_zero_is_invalid-2                                                                                2.05µs ± 0%     2.79µs ± 1%        10.08µs ± 1%
Draft7/const.json:const_with_false_does_not_match_0:float_zero_is_invalid-2                                                                                  2.14µs ± 1%     2.87µs ± 2%        10.36µs ± 2%
Draft7/const.json:const_with_true_does_not_match_1:true_is_valid-2                                                                                           1.14µs ± 0%     2.12µs ± 2%         6.75µs ± 1%
Draft7/const.json:const_with_true_does_not_match_1:integer_one_is_invalid-2                                                                                  2.14µs ± 1%     2.78µs ± 1%        10.72µs ± 0%
Draft7/const.json:const_with_true_does_not_match_1:float_one_is_invalid-2                                                                                    2.48µs ± 1%     2.78µs ± 0%        11.39µs ± 1%
Draft7/const.json:const_with_[false]_does_not_match_[0]:[false]_is_valid-2                                                                                   1.02µs ± 1%     2.83µs ± 1%         7.79µs ± 1%
Draft7/const.json:const_with_[false]_does_not_match_[0]:[0]_is_invalid-2                                                                                     1.39µs ± 2%     3.72µs ± 1%        10.72µs ± 2%
Draft7/const.json:const_with_[false]_does_not_match_[0]:[0.0]_is_invalid-2                                                                                   1.41µs ± 1%     3.74µs ± 1%        10.79µs ± 1%
Draft7/const.json:const_with_[true]_does_not_match_[1]:[true]_is_valid-2                                                                                     1.02µs ± 1%     2.85µs ± 1%         7.88µs ± 1%
Draft7/const.json:const_with_[true]_does_not_match_[1]:[1]_is_invalid-2                                                                                      1.40µs ± 0%     3.75µs ± 2%        10.73µs ± 1%
Draft7/const.json:const_with_[true]_does_not_match_[1]:[1.0]_is_invalid-2                                                                                    1.43µs ± 1%     3.83µs ± 1%        10.87µs ± 1%
Draft7/const.json:const_with_{"a":_false}_does_not_match_{"a":_0}:{"a":_false}_is_valid-2                                                                    1.32µs ± 3%     3.75µs ± 1%        11.32µs ± 1%
Draft7/const.json:const_with_{"a":_false}_does_not_match_{"a":_0}:{"a":_0}_is_invalid-2                                                                      1.69µs ± 1%     5.30µs ± 0%        14.48µs ± 1%
Draft7/const.json:const_with_{"a":_false}_does_not_match_{"a":_0}:{"a":_0.0}_is_invalid-2                                                                    1.70µs ± 1%     5.41µs ± 0%        14.46µs ± 0%
Draft7/const.json:const_with_{"a":_true}_does_not_match_{"a":_1}:{"a":_true}_is_valid-2                                                                      1.30µs ± 1%     3.76µs ± 1%
Draft7/const.json:const_with_{"a":_true}_does_not_match_{"a":_1}:{"a":_1}_is_invalid-2                                                                       1.71µs ± 1%     5.40µs ± 0%
Draft7/const.json:const_with_{"a":_true}_does_not_match_{"a":_1}:{"a":_1.0}_is_invalid-2                                                                     1.72µs ± 1%     5.42µs ± 1%
Draft7/const.json:const_with_0_does_not_match_other_zero-like_types:false_is_invalid-2                                                                       1.65µs ± 0%     2.82µs ± 2%
Draft7/const.json:const_with_0_does_not_match_other_zero-like_types:integer_zero_is_valid-2                                                                  2.02µs ± 2%     2.19µs ± 0%
Draft7/const.json:const_with_0_does_not_match_other_zero-like_types:float_zero_is_valid-2                                                                    2.16µs ± 3%     2.23µs ± 0%
Draft7/const.json:const_with_0_does_not_match_other_zero-like_types:empty_object_is_invalid-2                                                                1.40µs ± 1%     2.99µs ± 2%
Draft7/const.json:const_with_0_does_not_match_other_zero-like_types:empty_array_is_invalid-2                                                                 1.36µs ± 1%     2.99µs ± 3%
Draft7/const.json:const_with_0_does_not_match_other_zero-like_types:empty_string_is_invalid-2                                                                1.67µs ± 0%     2.84µs ± 1%
Draft7/const.json:const_with_1_does_not_match_true:true_is_invalid-2                                                                                                         2.90µs ± 2%
Draft7/const.json:const_with_1_does_not_match_true:integer_one_is_valid-2                                                                                                    2.30µs ± 1%
Draft7/const.json:const_with_1_does_not_match_true:float_one_is_valid-2                                                                                                      2.35µs ± 0%
[Geo mean]                                                                                                                                                   2.23µs          3.43µs             11.54µs     
```
</details>

### AJV Suite Alloc/Op

```
name \ alloc/op   santhosh-ajv.txt  qri-ajv.txt   xeipuuv-ajv.txt
[Geo mean]        5.00kB            5.16kB        77.47kB
```

Amount of memory needed for validation on average, as in previous latency benchmark, `xeipuuv/gojsonschema` suffers here from disability to reuse the schema between validation iterations.

<details><summary>full benchstat report</summary>
```
name \ alloc/op                                                                                                                         santhosh-ajv.txt  qri-ajv.txt   xeipuuv-ajv.txt
Ajv/advanced.json:advanced_schema_from_z-schema_benchmark_(https://github.com/zaggino/z-schema):valid_object_from_z-schema_benchmark-2       21.6kB ± 0%                   242.4kB ± 0%
Ajv/advanced.json:advanced_schema_from_z-schema_benchmark_(https://github.com/zaggino/z-schema):not_object-2                                 2.65kB ± 0%   2.31kB ± 0%    177.93kB ± 0%
Ajv/advanced.json:advanced_schema_from_z-schema_benchmark_(https://github.com/zaggino/z-schema):root_only_is_valid-2                         5.75kB ± 0%                  190.42kB ± 0%
Ajv/advanced.json:advanced_schema_from_z-schema_benchmark_(https://github.com/zaggino/z-schema):missing_root_entry-2                         2.79kB ± 0%   4.10kB ± 0%    182.34kB ± 0%
Ajv/advanced.json:advanced_schema_from_z-schema_benchmark_(https://github.com/zaggino/z-schema):invalid_entry_key-2                          8.45kB ± 0%   6.58kB ± 0%    197.69kB ± 0%
Ajv/advanced.json:advanced_schema_from_z-schema_benchmark_(https://github.com/zaggino/z-schema):missing_storage_in_entry-2                   2.20kB ± 0%   5.01kB ± 0%    181.54kB ± 0%
Ajv/advanced.json:advanced_schema_from_z-schema_benchmark_(https://github.com/zaggino/z-schema):missing_storage_type-2                       7.26kB ± 0%   5.42kB ± 0%    190.90kB ± 0%
Ajv/advanced.json:advanced_schema_from_z-schema_benchmark_(https://github.com/zaggino/z-schema):storage_type_should_be_a_string-2            8.30kB ± 0%   5.42kB ± 0%    192.73kB ± 0%
Ajv/advanced.json:advanced_schema_from_z-schema_benchmark_(https://github.com/zaggino/z-schema):storage_device_should_match_pattern-2        8.90kB ± 0%   5.44kB ± 0%    193.61kB ± 0%
Ajv/basic.json:basic_schema_from_z-schema_benchmark_(https://github.com/zaggino/z-schema):valid_array_from_z-schema_benchmark-2              6.66kB ± 0%  13.68kB ± 0%     31.79kB ± 0%
Ajv/basic.json:basic_schema_from_z-schema_benchmark_(https://github.com/zaggino/z-schema):not_array-2                                        2.65kB ± 0%   2.31kB ± 0%     24.68kB ± 0%
Ajv/basic.json:basic_schema_from_z-schema_benchmark_(https://github.com/zaggino/z-schema):array_of_not_onjects-2                             2.02kB ± 0%   3.64kB ± 0%     25.43kB ± 0%
Ajv/basic.json:basic_schema_from_z-schema_benchmark_(https://github.com/zaggino/z-schema):missing_required_properties-2                      1.50kB ± 0%   3.38kB ± 0%     24.83kB ± 0%
Ajv/basic.json:basic_schema_from_z-schema_benchmark_(https://github.com/zaggino/z-schema):required_property_of_wrong_type-2                  1.81kB ± 0%   4.78kB ± 0%     24.15kB ± 0%
Ajv/basic.json:basic_schema_from_z-schema_benchmark_(https://github.com/zaggino/z-schema):smallest_valid_product-2                           1.52kB ± 0%   4.55kB ± 0%     23.58kB ± 0%
Ajv/basic.json:basic_schema_from_z-schema_benchmark_(https://github.com/zaggino/z-schema):tags_should_be_array-2                             1.96kB ± 0%   5.08kB ± 0%     24.55kB ± 0%
Ajv/basic.json:basic_schema_from_z-schema_benchmark_(https://github.com/zaggino/z-schema):dimensions_should_be_object-2                      2.00kB ± 0%   5.09kB ± 0%     24.60kB ± 0%
Ajv/basic.json:basic_schema_from_z-schema_benchmark_(https://github.com/zaggino/z-schema):valid_product_with_tag-2                           1.66kB ± 0%   5.30kB ± 0%     24.26kB ± 0%
Ajv/basic.json:basic_schema_from_z-schema_benchmark_(https://github.com/zaggino/z-schema):dimensions_miss_required_properties-2              2.32kB ± 0%   6.56kB ± 0%     26.97kB ± 0%
Ajv/basic.json:basic_schema_from_z-schema_benchmark_(https://github.com/zaggino/z-schema):valid_product_with_tag_and_dimensions-2            2.32kB ± 0%   7.15kB ± 0%     25.67kB ± 0%
Ajv/complex.json:complex_schema_from_jsck_benchmark_(https://github.com/pandastrike/jsck):valid_array_from_jsck_benchmark-2                  37.0kB ± 0%                   202.3kB ± 0%
Ajv/complex.json:complex_schema_from_jsck_benchmark_(https://github.com/pandastrike/jsck):not_array-2                                        2.65kB ± 0%   2.31kB ± 0%    145.37kB ± 0%
Ajv/complex2.json:complex_schema_from_jsck_benchmark_without_IDs_in_definitions:valid_array_from_jsck_benchmark-2                            37.0kB ± 0%                   193.2kB ± 0%
Ajv/complex2.json:complex_schema_from_jsck_benchmark_without_IDs_in_definitions:not_array-2                                                  2.65kB ± 0%   2.31kB ± 0%    136.32kB ± 0%
Ajv/complex2.json:complex_schema_from_jsck_benchmark_without_IDs_in_definitions:one_valid_item-2                                             13.6kB ± 0%                   153.9kB ± 0%
Ajv/complex2.json:complex_schema_from_jsck_benchmark_without_IDs_in_definitions:one_invalid_item-2                                           15.3kB ± 0%    8.6kB ± 0%     156.2kB ± 0%
Ajv/complex3.json:complex_schema_from_jsck_benchmark_(https://github.com/pandastrike/jsck):valid_array_from_jsck_benchmark-2                 37.0kB ± 0%                   218.4kB ± 0%
Ajv/complex3.json:complex_schema_from_jsck_benchmark_(https://github.com/pandastrike/jsck):not_array-2                                       2.65kB ± 0%   2.33kB ± 0%    161.53kB ± 0%
Ajv/cosmicrealms.json:schema_from_cosmicrealms_benchmark:valid_data_from_cosmicrealms_benchmark-2                                            9.63kB ± 0%  16.76kB ± 0%     51.38kB ± 0%
Ajv/cosmicrealms.json:schema_from_cosmicrealms_benchmark:invalid_data-2                                                                      18.2kB ± 0%   20.5kB ± 0%      67.8kB ± 0%
Ajv/medium.json:medium_schema_from_jsck_benchmark_(https://github.com/pandastrike/jsck):valid_object_from_jsck_benchmark-2                   5.34kB ± 0%  14.50kB ± 0%     35.35kB ± 0%
Ajv/medium.json:medium_schema_from_jsck_benchmark_(https://github.com/pandastrike/jsck):not_object-2                                         2.65kB ± 0%   2.31kB ± 0%     31.96kB ± 0%
[Geo mean]                                                                                                                                   5.00kB        5.16kB          77.47kB
```
</details>

### Draft 7 Alloc/Op

```
name \ alloc/op  santhosh-draft7.txt  qri-draft7.txt  xeipuuv-draft7.txt
[Geo mean]       1.94kB               3.04kB          7.45kB     
```

<details><summary>full benchstat report</summary>
```
name \ alloc/op                                                                                                                                      santhosh-draft7.txt  qri-draft7.txt  xeipuuv-draft7.txt
Draft7/additionalItems.json:additionalItems_as_schema:additional_items_match_schema-2                                                                        1.38kB ± 0%     3.87kB ± 0%         6.46kB ± 0%
Draft7/additionalItems.json:additionalItems_as_schema:additional_items_do_not_match_schema-2                                                                 1.55kB ± 0%     4.03kB ± 0%         7.02kB ± 0%
Draft7/additionalItems.json:items_is_schema,_no_additionalItems:all_items_match_schema-2                                                                     1.50kB ± 0%     3.44kB ± 0%         6.16kB ± 0%
Draft7/additionalItems.json:array_of_items_with_no_additionalItems:empty_array-2                                                                               976B ± 0%      2408B ± 0%          5680B ± 0%
Draft7/additionalItems.json:array_of_items_with_no_additionalItems:fewer_number_of_items_present_(1)-2                                                       1.05kB ± 0%     2.61kB ± 0%         6.03kB ± 0%
Draft7/additionalItems.json:array_of_items_with_no_additionalItems:fewer_number_of_items_present_(2)-2                                                       1.14kB ± 0%     2.82kB ± 0%         6.40kB ± 0%
Draft7/additionalItems.json:array_of_items_with_no_additionalItems:equal_number_of_items_present-2                                                           1.26kB ± 0%     3.07kB ± 0%         6.80kB ± 0%
Draft7/additionalItems.json:array_of_items_with_no_additionalItems:additional_items_are_not_permitted-2                                                      1.46kB ± 0%     3.14kB ± 0%         7.55kB ± 0%
Draft7/additionalItems.json:additionalItems_as_false_without_items:items_defaults_to_empty_schema_so_everything_is_valid-2                                   1.30kB ± 0%     2.48kB ± 0%         3.92kB ± 0%
Draft7/additionalItems.json:additionalItems_as_false_without_items:ignores_non-arrays-2                                                                      1.30kB ± 0%     2.53kB ± 0%         3.86kB ± 0%
Draft7/additionalItems.json:additionalItems_are_allowed_by_default:only_the_first_item_is_validated-2                                                        1.21kB ± 0%     2.73kB ± 0%         5.14kB ± 0%
Draft7/additionalItems.json:additionalItems_should_not_look_in_applicators,_valid_case:items_defined_in_allOf_are_not_examined-2                             1.12kB ± 0%                         7.16kB ± 0%
Draft7/additionalItems.json:additionalItems_should_not_look_in_applicators,_invalid_case:items_defined_in_allOf_are_not_examined-2                           1.47kB ± 0%                        10.46kB ± 0%
Draft7/additionalProperties.json:additionalProperties_being_false_does_not_allow_other_properties:no_additional_properties_is_valid-2                        1.34kB ± 0%     3.62kB ± 0%         9.05kB ± 0%
Draft7/additionalProperties.json:additionalProperties_being_false_does_not_allow_other_properties:an_additional_property_is_invalid-2                        1.67kB ± 0%     3.98kB ± 0%        12.11kB ± 0%
Draft7/additionalProperties.json:additionalProperties_being_false_does_not_allow_other_properties:ignores_arrays-2                                           1.14kB ± 0%     2.34kB ± 0%         7.64kB ± 0%
Draft7/additionalProperties.json:additionalProperties_being_false_does_not_allow_other_properties:ignores_strings-2                                          2.50kB ± 0%     2.19kB ± 0%         9.00kB ± 0%
Draft7/additionalProperties.json:additionalProperties_being_false_does_not_allow_other_properties:ignores_other_non-objects-2                                2.53kB ± 0%     2.18kB ± 0%         9.18kB ± 0%
Draft7/additionalProperties.json:additionalProperties_being_false_does_not_allow_other_properties:patternProperties_are_not_additional_properties-2          1.40kB ± 0%     3.95kB ± 0%        10.37kB ± 0%
Draft7/additionalProperties.json:non-ASCII_pattern_with_additionalProperties:matching_the_pattern_is_valid-2                                                 1.35kB ± 0%     3.52kB ± 0%         7.53kB ± 0%
Draft7/additionalProperties.json:non-ASCII_pattern_with_additionalProperties:not_matching_the_pattern_is_invalid-2                                           1.56kB ± 0%     2.95kB ± 0%         7.96kB ± 0%
Draft7/additionalProperties.json:additionalProperties_allows_a_schema_which_should_validate:no_additional_properties_is_valid-2                              1.34kB ± 0%     3.61kB ± 0%         6.75kB ± 0%
Draft7/additionalProperties.json:additionalProperties_allows_a_schema_which_should_validate:an_additional_valid_property_is_valid-2                          1.41kB ± 0%     4.02kB ± 0%         7.14kB ± 0%
Draft7/additionalProperties.json:additionalProperties_allows_a_schema_which_should_validate:an_additional_invalid_property_is_invalid-2                      1.67kB ± 0%     4.18kB ± 0%         8.06kB ± 0%
Draft7/additionalProperties.json:additionalProperties_can_exist_by_itself:an_additional_valid_property_is_valid-2                                            1.29kB ± 0%     3.33kB ± 0%         4.89kB ± 0%
Draft7/additionalProperties.json:additionalProperties_can_exist_by_itself:an_additional_invalid_property_is_invalid-2                                        1.54kB ± 0%     3.49kB ± 0%         5.82kB ± 0%
Draft7/additionalProperties.json:additionalProperties_are_allowed_by_default:additional_properties_are_allowed-2                                             1.41kB ± 0%     3.52kB ± 0%         6.11kB ± 0%
Draft7/additionalProperties.json:additionalProperties_should_not_look_in_applicators:properties_defined_in_allOf_are_not_examined-2                          1.59kB ± 0%                         8.28kB ± 0%
Draft7/allOf.json:allOf:allOf-2                                                                                                                              1.41kB ± 0%     5.54kB ± 0%         9.18kB ± 0%
Draft7/allOf.json:allOf:mismatch_second-2                                                                                                                    1.70kB ± 0%     4.82kB ± 0%        10.32kB ± 0%
Draft7/allOf.json:allOf:mismatch_first-2                                                                                                                     1.78kB ± 0%     4.81kB ± 0%        10.51kB ± 0%
Draft7/allOf.json:allOf:wrong_type-2                                                                                                                         1.78kB ± 0%     5.50kB ± 0%        10.48kB ± 0%
Draft7/allOf.json:allOf_with_base_schema:valid-2                                                                                                             1.41kB ± 0%     5.96kB ± 0%        10.58kB ± 0%
Draft7/allOf.json:allOf_with_base_schema:mismatch_base_schema-2                                                                                              1.52kB ± 0%     5.82kB ± 0%        11.00kB ± 0%
Draft7/allOf.json:allOf_with_base_schema:mismatch_first_allOf-2                                                                                              1.79kB ± 0%     5.55kB ± 0%        11.92kB ± 0%
Draft7/allOf.json:allOf_with_base_schema:mismatch_second_allOf-2                                                                                             1.81kB ± 0%     5.57kB ± 0%        12.01kB ± 0%
Draft7/allOf.json:allOf_with_base_schema:mismatch_both-2                                                                                                     2.36kB ± 0%     4.89kB ± 0%        12.66kB ± 0%
Draft7/allOf.json:allOf_simple_types:valid-2                                                                                                                 2.62kB ± 0%     3.06kB ± 0%         8.34kB ± 0%
Draft7/allOf.json:allOf_simple_types:mismatch_one-2                                                                                                          3.64kB ± 0%     3.24kB ± 0%        10.53kB ± 0%
Draft7/allOf.json:allOf_with_boolean_schemas,_all_true:any_value_is_valid-2                                                                                  2.48kB ± 0%     3.06kB ± 0%         6.33kB ± 0%
Draft7/allOf.json:allOf_with_boolean_schemas,_some_false:any_value_is_invalid-2                                                                              2.76kB ± 0%     3.22kB ± 0%         7.75kB ± 0%
Draft7/allOf.json:allOf_with_boolean_schemas,_all_false:any_value_is_invalid-2                                                                               3.22kB ± 0%     3.43kB ± 0%         8.50kB ± 0%
Draft7/allOf.json:allOf_with_one_empty_schema:any_data_is_valid-2                                                                                            2.56kB ± 0%     2.66kB ± 0%         6.32kB ± 0%
Draft7/allOf.json:allOf_with_two_empty_schemas:any_data_is_valid-2                                                                                           2.60kB ± 0%     3.05kB ± 0%         7.25kB ± 0%
Draft7/allOf.json:allOf_with_the_first_empty_schema:number_is_valid-2                                                                                        2.60kB ± 0%     3.05kB ± 0%         7.58kB ± 0%
Draft7/allOf.json:allOf_with_the_first_empty_schema:string_is_invalid-2                                                                                      2.83kB ± 0%     3.27kB ± 0%         8.46kB ± 0%
Draft7/allOf.json:allOf_with_the_last_empty_schema:number_is_valid-2                                                                                         2.60kB ± 0%     3.05kB ± 0%         7.58kB ± 0%
Draft7/allOf.json:allOf_with_the_last_empty_schema:string_is_invalid-2                                                                                       2.83kB ± 0%     3.27kB ± 0%         8.46kB ± 0%
Draft7/allOf.json:nested_allOf,_to_check_validation_semantics:null_is_valid-2                                                                                2.48kB ± 0%     3.17kB ± 0%         7.24kB ± 0%
Draft7/allOf.json:nested_allOf,_to_check_validation_semantics:anything_non-null_is_invalid-2                                                                 3.17kB ± 0%     3.42kB ± 0%        10.14kB ± 0%
Draft7/allOf.json:allOf_combined_with_anyOf,_oneOf:allOf:_false,_anyOf:_false,_oneOf:_false-2                                                                6.24kB ± 0%     4.24kB ± 0%        16.62kB ± 0%
Draft7/allOf.json:allOf_combined_with_anyOf,_oneOf:allOf:_false,_anyOf:_false,_oneOf:_true-2                                                                 5.10kB ± 0%     3.93kB ± 0%        14.35kB ± 0%
Draft7/allOf.json:allOf_combined_with_anyOf,_oneOf:allOf:_false,_anyOf:_true,_oneOf:_false-2                                                                 5.10kB ± 0%     3.93kB ± 0%        14.35kB ± 0%
Draft7/allOf.json:allOf_combined_with_anyOf,_oneOf:allOf:_false,_anyOf:_true,_oneOf:_true-2                                                                  3.91kB ± 0%     3.72kB ± 0%        12.16kB ± 0%
Draft7/allOf.json:allOf_combined_with_anyOf,_oneOf:allOf:_true,_anyOf:_false,_oneOf:_false-2                                                                 5.21kB ± 0%     3.93kB ± 0%        14.43kB ± 0%
Draft7/allOf.json:allOf_combined_with_anyOf,_oneOf:allOf:_true,_anyOf:_false,_oneOf:_true-2                                                                  4.01kB ± 0%     3.72kB ± 0%        12.24kB ± 0%
Draft7/allOf.json:allOf_combined_with_anyOf,_oneOf:allOf:_true,_anyOf:_true,_oneOf:_false-2                                                                  3.98kB ± 0%     3.71kB ± 0%        12.22kB ± 0%
Draft7/allOf.json:allOf_combined_with_anyOf,_oneOf:allOf:_true,_anyOf:_true,_oneOf:_true-2                                                                   2.96kB ± 0%     3.55kB ± 0%        10.05kB ± 0%
Draft7/anyOf.json:anyOf:first_anyOf_valid-2                                                                                                                  2.60kB ± 0%     2.55kB ± 0%         7.74kB ± 0%
Draft7/anyOf.json:anyOf:second_anyOf_valid-2                                                                                                                 3.02kB ± 0%     3.09kB ± 0%         8.98kB ± 0%
Draft7/anyOf.json:anyOf:both_anyOf_valid-2                                                                                                                   2.60kB ± 0%     2.55kB ± 0%         7.74kB ± 0%
Draft7/anyOf.json:anyOf:neither_anyOf_valid-2                                                                                                                4.02kB ± 0%     3.27kB ± 0%        11.11kB ± 0%
Draft7/anyOf.json:anyOf_with_base_schema:mismatch_base_schema-2                                                                                              2.65kB ± 0%     2.70kB ± 0%         8.11kB ± 0%
Draft7/anyOf.json:anyOf_with_base_schema:one_anyOf_valid-2                                                                                                   2.66kB ± 0%     3.08kB ± 0%         8.24kB ± 0%
Draft7/anyOf.json:anyOf_with_base_schema:both_anyOf_invalid-2                                                                                                3.14kB ± 0%     3.27kB ± 0%         9.71kB ± 0%
Draft7/anyOf.json:anyOf_with_boolean_schemas,_all_true:any_value_is_valid-2                                                                                  2.48kB ± 0%     2.56kB ± 0%         6.30kB ± 0%
Draft7/anyOf.json:anyOf_with_boolean_schemas,_some_true:any_value_is_valid-2                                                                                 2.48kB ± 0%     2.56kB ± 0%         6.30kB ± 0%
Draft7/anyOf.json:anyOf_with_boolean_schemas,_all_false:any_value_is_invalid-2                                                                               2.95kB ± 0%     3.20kB ± 0%         8.44kB ± 0%
Draft7/anyOf.json:anyOf_complex_types:first_anyOf_valid_(complex)-2                                                                                          1.38kB ± 0%     3.96kB ± 0%         9.02kB ± 0%
Draft7/anyOf.json:anyOf_complex_types:second_anyOf_valid_(complex)-2                                                                                         1.54kB ± 0%     4.64kB ± 0%         9.58kB ± 0%
Draft7/anyOf.json:anyOf_complex_types:both_anyOf_valid_(complex)-2                                                                                           1.41kB ± 0%     3.98kB ± 0%         9.06kB ± 0%
Draft7/anyOf.json:anyOf_complex_types:neither_anyOf_valid_(complex)-2                                                                                        2.18kB ± 0%     5.27kB ± 0%        11.30kB ± 0%
Draft7/anyOf.json:anyOf_with_one_empty_schema:string_is_valid-2                                                                                              2.67kB ± 0%     3.09kB ± 0%         7.73kB ± 0%
Draft7/anyOf.json:anyOf_with_one_empty_schema:number_is_valid-2                                                                                              2.58kB ± 0%     2.56kB ± 0%         7.30kB ± 0%
Draft7/anyOf.json:nested_anyOf,_to_check_validation_semantics:null_is_valid-2                                                                                2.48kB ± 0%     2.94kB ± 0%         7.24kB ± 0%
Draft7/anyOf.json:nested_anyOf,_to_check_validation_semantics:anything_non-null_is_invalid-2                                                                 3.18kB ± 0%     3.20kB ± 0%        10.11kB ± 0%
Draft7/anyOf.json:nested_anyOf,_to_check_validation_semantics:null_is_valid#01-2                                                                             2.48kB ± 0%     2.94kB ± 0%         7.24kB ± 0%
Draft7/anyOf.json:nested_anyOf,_to_check_validation_semantics:anything_non-null_is_invalid#01-2                                                              3.18kB ± 0%     3.20kB ± 0%        10.11kB ± 0%
Draft7/boolean_schema.json:boolean_schema_'true':number_is_valid-2                                                                                           2.48kB ± 0%     2.17kB ± 0%         6.18kB ± 0%
Draft7/boolean_schema.json:boolean_schema_'true':string_is_valid-2                                                                                           2.48kB ± 0%     2.18kB ± 0%         6.18kB ± 0%
Draft7/boolean_schema.json:boolean_schema_'true':boolean_true_is_valid-2                                                                                     2.46kB ± 0%     2.16kB ± 0%         6.17kB ± 0%
Draft7/boolean_schema.json:boolean_schema_'true':boolean_false_is_valid-2                                                                                    2.46kB ± 0%     2.16kB ± 0%         6.17kB ± 0%
Draft7/boolean_schema.json:boolean_schema_'true':null_is_valid-2                                                                                             2.48kB ± 0%     2.18kB ± 0%         6.18kB ± 0%
Draft7/boolean_schema.json:boolean_schema_'true':object_is_valid-2                                                                                           1.30kB ± 0%     2.53kB ± 0%         5.01kB ± 0%
Draft7/boolean_schema.json:boolean_schema_'true':empty_object_is_valid-2                                                                                       992B ± 0%      2216B ± 0%          4696B ± 0%
Draft7/boolean_schema.json:boolean_schema_'true':array_is_valid-2                                                                                            1.02kB ± 0%     2.24kB ± 0%         4.72kB ± 0%
Draft7/boolean_schema.json:boolean_schema_'true':empty_array_is_valid-2                                                                                        976B ± 0%      2200B ± 0%          4680B ± 0%
Draft7/boolean_schema.json:boolean_schema_'false':number_is_invalid-2                                                                                        2.59kB ± 0%     2.26kB ± 0%         6.86kB ± 0%
Draft7/boolean_schema.json:boolean_schema_'false':string_is_invalid-2                                                                                        2.59kB ± 0%     2.27kB ± 0%         6.87kB ± 0%
Draft7/boolean_schema.json:boolean_schema_'false':boolean_true_is_invalid-2                                                                                  2.58kB ± 0%     2.26kB ± 0%         6.85kB ± 0%
Draft7/boolean_schema.json:boolean_schema_'false':boolean_false_is_invalid-2                                                                                 2.58kB ± 0%     2.26kB ± 0%         6.85kB ± 0%
Draft7/boolean_schema.json:boolean_schema_'false':null_is_invalid-2                                                                                          2.59kB ± 0%     2.27kB ± 0%         6.86kB ± 0%
Draft7/boolean_schema.json:boolean_schema_'false':object_is_invalid-2                                                                                        1.42kB ± 0%     2.62kB ± 0%         5.70kB ± 0%
Draft7/boolean_schema.json:boolean_schema_'false':empty_object_is_invalid-2                                                                                  1.10kB ± 0%     2.31kB ± 0%         5.38kB ± 0%
Draft7/boolean_schema.json:boolean_schema_'false':array_is_invalid-2                                                                                         1.12kB ± 0%     2.34kB ± 0%         5.40kB ± 0%
Draft7/boolean_schema.json:boolean_schema_'false':empty_array_is_invalid-2                                                                                   1.09kB ± 0%     2.30kB ± 0%         5.37kB ± 0%
Draft7/const.json:const_validation:same_value_is_valid-2                                                                                                     2.60kB ± 0%     2.37kB ± 0%         5.82kB ± 0%
Draft7/const.json:const_validation:another_value_is_invalid-2                                                                                                2.75kB ± 0%     2.48kB ± 0%         6.58kB ± 0%
Draft7/const.json:const_validation:another_type_is_invalid-2                                                                                                 2.63kB ± 0%     2.49kB ± 0%         6.41kB ± 0%
Draft7/const.json:const_with_object:same_object_is_valid-2                                                                                                   1.33kB ± 0%     3.38kB ± 0%         7.51kB ± 0%
Draft7/const.json:const_with_object:same_object_with_different_property_order_is_valid-2                                                                     1.33kB ± 0%     3.38kB ± 0%         7.51kB ± 0%
Draft7/const.json:const_with_object:another_object_is_invalid-2                                                                                              1.44kB ± 0%     3.75kB ± 0%         7.82kB ± 0%
Draft7/const.json:const_with_object:another_type_is_invalid-2                                                                                                1.19kB ± 0%     3.49kB ± 0%         6.85kB ± 0%
Draft7/const.json:const_with_array:same_array_is_valid-2                                                                                                     1.38kB ± 0%     3.39kB ± 0%         6.94kB ± 0%
Draft7/const.json:const_with_array:another_array_item_is_invalid-2                                                                                           1.14kB ± 0%     3.25kB ± 0%         6.46kB ± 0%
Draft7/const.json:const_with_array:array_with_additional_items_is_invalid-2                                                                                  1.27kB ± 0%     3.36kB ± 0%         6.70kB ± 0%
Draft7/const.json:const_with_null:null_is_valid-2                                                                                                            2.48kB ± 0%     2.38kB ± 0%         5.59kB ± 0%
Draft7/const.json:const_with_null:not_null_is_invalid-2                                                                                                      2.66kB ± 0%     2.49kB ± 0%         6.52kB ± 0%
Draft7/const.json:const_with_false_does_not_match_0:false_is_valid-2                                                                                         2.46kB ± 0%     2.35kB ± 0%         5.54kB ± 0%
Draft7/const.json:const_with_false_does_not_match_0:integer_zero_is_invalid-2                                                                                2.66kB ± 0%     2.47kB ± 0%         6.50kB ± 0%
Draft7/const.json:const_with_false_does_not_match_0:float_zero_is_invalid-2                                                                                  2.67kB ± 0%     2.47kB ± 0%         6.52kB ± 0%
Draft7/const.json:const_with_true_does_not_match_1:true_is_valid-2                                                                                           2.46kB ± 0%     2.35kB ± 0%         5.54kB ± 0%
Draft7/const.json:const_with_true_does_not_match_1:integer_one_is_invalid-2                                                                                  2.67kB ± 0%     2.48kB ± 0%         6.58kB ± 0%
Draft7/const.json:const_with_true_does_not_match_1:float_one_is_invalid-2                                                                                    2.78kB ± 0%     2.48kB ± 0%         6.69kB ± 0%
Draft7/const.json:const_with_[false]_does_not_match_[0]:[false]_is_valid-2                                                                                     992B ± 0%      2464B ± 0%          4328B ± 0%
Draft7/const.json:const_with_[false]_does_not_match_[0]:[0]_is_invalid-2                                                                                     1.14kB ± 0%     2.60kB ± 0%         5.11kB ± 0%
Draft7/const.json:const_with_[false]_does_not_match_[0]:[0.0]_is_invalid-2                                                                                   1.15kB ± 0%     2.61kB ± 0%         5.13kB ± 0%
Draft7/const.json:const_with_[true]_does_not_match_[1]:[true]_is_valid-2                                                                                       992B ± 0%      2464B ± 0%          4328B ± 0%
Draft7/const.json:const_with_[true]_does_not_match_[1]:[1]_is_invalid-2                                                                                      1.14kB ± 0%     2.61kB ± 0%         5.11kB ± 0%
Draft7/const.json:const_with_[true]_does_not_match_[1]:[1.0]_is_invalid-2                                                                                    1.15kB ± 0%     2.62kB ± 0%         5.13kB ± 0%
Draft7/const.json:const_with_{"a":_false}_does_not_match_{"a":_0}:{"a":_false}_is_valid-2                                                                    1.28kB ± 0%     3.22kB ± 0%         6.46kB ± 0%
Draft7/const.json:const_with_{"a":_false}_does_not_match_{"a":_0}:{"a":_0}_is_invalid-2                                                                      1.43kB ± 0%     3.62kB ± 0%         7.22kB ± 0%
Draft7/const.json:const_with_{"a":_false}_does_not_match_{"a":_0}:{"a":_0.0}_is_invalid-2                                                                    1.44kB ± 0%     3.62kB ± 0%         7.26kB ± 0%
Draft7/const.json:const_with_{"a":_true}_does_not_match_{"a":_1}:{"a":_true}_is_valid-2                                                                      1.28kB ± 0%     3.22kB ± 0%
Draft7/const.json:const_with_{"a":_true}_does_not_match_{"a":_1}:{"a":_1}_is_invalid-2                                                                       1.43kB ± 0%     3.62kB ± 0%
Draft7/const.json:const_with_{"a":_true}_does_not_match_{"a":_1}:{"a":_1.0}_is_invalid-2                                                                     1.44kB ± 0%     3.63kB ± 0%
Draft7/const.json:const_with_0_does_not_match_other_zero-like_types:false_is_invalid-2                                                                       2.62kB ± 0%     2.46kB ± 0%
Draft7/const.json:const_with_0_does_not_match_other_zero-like_types:integer_zero_is_valid-2                                                                  2.58kB ± 0%     2.35kB ± 0%
Draft7/const.json:const_with_0_does_not_match_other_zero-like_types:float_zero_is_valid-2                                                                    2.58kB ± 0%     2.35kB ± 0%
Draft7/const.json:const_with_0_does_not_match_other_zero-like_types:empty_object_is_invalid-2                                                                1.14kB ± 0%     2.52kB ± 0%
Draft7/const.json:const_with_0_does_not_match_other_zero-like_types:empty_array_is_invalid-2                                                                 1.13kB ± 0%     2.50kB ± 0%
Draft7/const.json:const_with_0_does_not_match_other_zero-like_types:empty_string_is_invalid-2                                                                2.62kB ± 0%     2.46kB ± 0%
Draft7/const.json:const_with_1_does_not_match_true:true_is_invalid-2                                                                                                         2.47kB ± 0%
Draft7/const.json:const_with_1_does_not_match_true:integer_one_is_valid-2                                                                                                    2.37kB ± 0%
Draft7/const.json:const_with_1_does_not_match_true:float_one_is_valid-2                                                                                                      2.38kB ± 0%
[Geo mean]                                                                                                                                                   1.94kB          3.04kB              7.45kB     
```
</details>

### AJV Suite Allocs/Op

```
name \ allocs/op  santhosh-ajv.txt  qri-ajv.txt   xeipuuv-ajv.txt
[Geo mean]        73.2              77.7          1043.9
```

Schemas in ajv suite are rather large, that's not an issue for validators that reuse schema between validation, but for `xeipuuv/gojsonschema` unwanted unmarshaling leads to a whopping penalty of 13x more allocations per validation.

<details><summary>full benchstat report</summary>
```
name \ allocs/op                                                                                                                        santhosh-ajv.txt  qri-ajv.txt   xeipuuv-ajv.txt
Ajv/advanced.json:advanced_schema_from_z-schema_benchmark_(https://github.com/zaggino/z-schema):valid_object_from_z-schema_benchmark-2          495 ± 0%                      3376 ± 0%
Ajv/advanced.json:advanced_schema_from_z-schema_benchmark_(https://github.com/zaggino/z-schema):not_object-2                                   11.0 ± 0%     22.0 ± 0%      2044.0 ± 0%
Ajv/advanced.json:advanced_schema_from_z-schema_benchmark_(https://github.com/zaggino/z-schema):root_only_is_valid-2                            125 ± 0%                      2341 ± 0%
Ajv/advanced.json:advanced_schema_from_z-schema_benchmark_(https://github.com/zaggino/z-schema):missing_root_entry-2                           49.0 ± 0%     54.0 ± 0%      2128.0 ± 0%
Ajv/advanced.json:advanced_schema_from_z-schema_benchmark_(https://github.com/zaggino/z-schema):invalid_entry_key-2                             154 ± 0%      113 ± 0%        2432 ± 0%
Ajv/advanced.json:advanced_schema_from_z-schema_benchmark_(https://github.com/zaggino/z-schema):missing_storage_in_entry-2                     38.0 ± 0%     80.0 ± 0%      2117.0 ± 0%
Ajv/advanced.json:advanced_schema_from_z-schema_benchmark_(https://github.com/zaggino/z-schema):missing_storage_type-2                          195 ± 0%       87 ± 0%        2319 ± 0%
Ajv/advanced.json:advanced_schema_from_z-schema_benchmark_(https://github.com/zaggino/z-schema):storage_type_should_be_a_string-2               218 ± 0%       88 ± 0%        2372 ± 0%
Ajv/advanced.json:advanced_schema_from_z-schema_benchmark_(https://github.com/zaggino/z-schema):storage_device_should_match_pattern-2           234 ± 0%       88 ± 0%        2388 ± 0%
Ajv/basic.json:basic_schema_from_z-schema_benchmark_(https://github.com/zaggino/z-schema):valid_array_from_z-schema_benchmark-2                 137 ± 0%      280 ± 0%         542 ± 0%
Ajv/basic.json:basic_schema_from_z-schema_benchmark_(https://github.com/zaggino/z-schema):not_array-2                                          11.0 ± 0%     22.0 ± 0%       254.0 ± 0%
Ajv/basic.json:basic_schema_from_z-schema_benchmark_(https://github.com/zaggino/z-schema):array_of_not_onjects-2                               42.0 ± 0%     61.0 ± 0%       322.0 ± 0%
Ajv/basic.json:basic_schema_from_z-schema_benchmark_(https://github.com/zaggino/z-schema):missing_required_properties-2                        30.0 ± 0%     47.0 ± 0%       290.0 ± 0%
Ajv/basic.json:basic_schema_from_z-schema_benchmark_(https://github.com/zaggino/z-schema):required_property_of_wrong_type-2                    34.0 ± 0%     78.0 ± 0%       289.0 ± 0%
Ajv/basic.json:basic_schema_from_z-schema_benchmark_(https://github.com/zaggino/z-schema):smallest_valid_product-2                             25.0 ± 0%     69.0 ± 0%       285.0 ± 0%
Ajv/basic.json:basic_schema_from_z-schema_benchmark_(https://github.com/zaggino/z-schema):tags_should_be_array-2                               41.0 ± 0%     89.0 ± 0%       310.0 ± 0%
Ajv/basic.json:basic_schema_from_z-schema_benchmark_(https://github.com/zaggino/z-schema):dimensions_should_be_object-2                        41.0 ± 0%     89.0 ± 0%       310.0 ± 0%
Ajv/basic.json:basic_schema_from_z-schema_benchmark_(https://github.com/zaggino/z-schema):valid_product_with_tag-2                             32.0 ± 0%     93.0 ± 0%       312.0 ± 0%
Ajv/basic.json:basic_schema_from_z-schema_benchmark_(https://github.com/zaggino/z-schema):dimensions_miss_required_properties-2                56.0 ± 0%    127.0 ± 0%       367.0 ± 0%
Ajv/basic.json:basic_schema_from_z-schema_benchmark_(https://github.com/zaggino/z-schema):valid_product_with_tag_and_dimensions-2              54.0 ± 0%    140.0 ± 0%       374.0 ± 0%
Ajv/complex.json:complex_schema_from_jsck_benchmark_(https://github.com/pandastrike/jsck):valid_array_from_jsck_benchmark-2                     579 ± 0%                      3528 ± 0%
Ajv/complex.json:complex_schema_from_jsck_benchmark_(https://github.com/pandastrike/jsck):not_array-2                                          11.0 ± 0%     22.0 ± 0%      1928.0 ± 0%
Ajv/complex2.json:complex_schema_from_jsck_benchmark_without_IDs_in_definitions:valid_array_from_jsck_benchmark-2                               579 ± 0%                      3396 ± 0%
Ajv/complex2.json:complex_schema_from_jsck_benchmark_without_IDs_in_definitions:not_array-2                                                    11.0 ± 0%     22.0 ± 0%      1796.0 ± 0%
Ajv/complex2.json:complex_schema_from_jsck_benchmark_without_IDs_in_definitions:one_valid_item-2                                                195 ± 0%                      2291 ± 0%
Ajv/complex2.json:complex_schema_from_jsck_benchmark_without_IDs_in_definitions:one_invalid_item-2                                              255 ± 0%      165 ± 0%        2338 ± 0%
Ajv/complex3.json:complex_schema_from_jsck_benchmark_(https://github.com/pandastrike/jsck):valid_array_from_jsck_benchmark-2                    579 ± 0%                      3927 ± 0%
Ajv/complex3.json:complex_schema_from_jsck_benchmark_(https://github.com/pandastrike/jsck):not_array-2                                         11.0 ± 0%     23.0 ± 0%      2327.0 ± 0%
Ajv/cosmicrealms.json:schema_from_cosmicrealms_benchmark:valid_data_from_cosmicrealms_benchmark-2                                               212 ± 0%      375 ± 0%         991 ± 0%
Ajv/cosmicrealms.json:schema_from_cosmicrealms_benchmark:invalid_data-2                                                                         432 ± 0%      439 ± 0%        1238 ± 0%
Ajv/medium.json:medium_schema_from_jsck_benchmark_(https://github.com/pandastrike/jsck):valid_object_from_jsck_benchmark-2                     69.0 ± 0%    240.0 ± 0%       450.0 ± 0%
Ajv/medium.json:medium_schema_from_jsck_benchmark_(https://github.com/pandastrike/jsck):not_object-2                                           11.0 ± 0%     22.0 ± 0%       348.0 ± 0%
[Geo mean]                                                                                                                                     73.2          77.7           1043.9
```
</details>

### Draft 7 Allocs/Op

```
name \ allocs/op   santhosh-draft7.txt  qri-draft7.txt  xeipuuv-draft7.txt
[Geo mean]         14.8                 35.0            86.9
```

With smaller schemas in draft 7 suite, schema unmarshaling penalty is less pronounced. In this test `santhosh-tekuri/jsonschema` shows excellent result, making validation affordable even in hot loops.

<details><summary>full benchstat report</summary>
```
name \ allocs/op                                                                                                                                     santhosh-draft7.txt  qri-draft7.txt  xeipuuv-draft7.txt
Draft7/additionalItems.json:additionalItems_as_schema:additional_items_match_schema-2                                                                          25.0 ± 0%       60.0 ± 0%          114.0 ± 0%
Draft7/additionalItems.json:additionalItems_as_schema:additional_items_do_not_match_schema-2                                                                   30.0 ± 0%       67.0 ± 0%          120.0 ± 0%
Draft7/additionalItems.json:items_is_schema,_no_additionalItems:all_items_match_schema-2                                                                       26.0 ± 0%       56.0 ± 0%          136.0 ± 0%
Draft7/additionalItems.json:array_of_items_with_no_additionalItems:empty_array-2                                                                               7.00 ± 0%      21.00 ± 0%          58.00 ± 0%
Draft7/additionalItems.json:array_of_items_with_no_additionalItems:fewer_number_of_items_present_(1)-2                                                         11.0 ± 0%       29.0 ± 0%           77.0 ± 0%
Draft7/additionalItems.json:array_of_items_with_no_additionalItems:fewer_number_of_items_present_(2)-2                                                         15.0 ± 0%       37.0 ± 0%           96.0 ± 0%
Draft7/additionalItems.json:array_of_items_with_no_additionalItems:equal_number_of_items_present-2                                                             19.0 ± 0%       45.0 ± 0%          115.0 ± 0%
Draft7/additionalItems.json:array_of_items_with_no_additionalItems:additional_items_are_not_permitted-2                                                        24.0 ± 0%       48.0 ± 0%          131.0 ± 0%
Draft7/additionalItems.json:additionalItems_as_false_without_items:items_defaults_to_empty_schema_so_everything_is_valid-2                                     16.0 ± 0%       27.0 ± 0%           43.0 ± 0%
Draft7/additionalItems.json:additionalItems_as_false_without_items:ignores_non-arrays-2                                                                        11.0 ± 0%       22.0 ± 0%           36.0 ± 0%
Draft7/additionalItems.json:additionalItems_are_allowed_by_default:only_the_first_item_is_validated-2                                                          17.0 ± 0%       33.0 ± 0%           73.0 ± 0%
Draft7/additionalItems.json:additionalItems_should_not_look_in_applicators,_valid_case:items_defined_in_allOf_are_not_examined-2                               14.0 ± 0%                           94.0 ± 0%
Draft7/additionalItems.json:additionalItems_should_not_look_in_applicators,_invalid_case:items_defined_in_allOf_are_not_examined-2                             28.0 ± 0%                          165.0 ± 0%
Draft7/additionalProperties.json:additionalProperties_being_false_does_not_allow_other_properties:no_additional_properties_is_valid-2                          12.0 ± 0%       42.0 ± 0%          106.0 ± 0%
Draft7/additionalProperties.json:additionalProperties_being_false_does_not_allow_other_properties:an_additional_property_is_invalid-2                          27.0 ± 0%       59.0 ± 0%          170.0 ± 0%
Draft7/additionalProperties.json:additionalProperties_being_false_does_not_allow_other_properties:ignores_arrays-2                                             13.0 ± 0%       24.0 ± 0%           82.0 ± 0%
Draft7/additionalProperties.json:additionalProperties_being_false_does_not_allow_other_properties:ignores_strings-2                                            7.00 ± 0%      18.00 ± 0%          78.00 ± 0%
Draft7/additionalProperties.json:additionalProperties_being_false_does_not_allow_other_properties:ignores_other_non-objects-2                                  10.0 ± 0%       18.0 ± 0%           89.0 ± 0%
Draft7/additionalProperties.json:additionalProperties_being_false_does_not_allow_other_properties:patternProperties_are_not_additional_properties-2            16.0 ± 0%       51.0 ± 0%          138.0 ± 0%
Draft7/additionalProperties.json:non-ASCII_pattern_with_additionalProperties:matching_the_pattern_is_valid-2                                                   12.0 ± 0%       38.0 ± 0%           92.0 ± 0%
Draft7/additionalProperties.json:non-ASCII_pattern_with_additionalProperties:not_matching_the_pattern_is_invalid-2                                             18.0 ± 0%       32.0 ± 0%           92.0 ± 0%
Draft7/additionalProperties.json:additionalProperties_allows_a_schema_which_should_validate:no_additional_properties_is_valid-2                                12.0 ± 0%       41.0 ± 0%           79.0 ± 0%
Draft7/additionalProperties.json:additionalProperties_allows_a_schema_which_should_validate:an_additional_valid_property_is_valid-2                            17.0 ± 0%       57.0 ± 0%          100.0 ± 0%
Draft7/additionalProperties.json:additionalProperties_allows_a_schema_which_should_validate:an_additional_invalid_property_is_invalid-2                        27.0 ± 0%       65.0 ± 0%          128.0 ± 0%
Draft7/additionalProperties.json:additionalProperties_can_exist_by_itself:an_additional_valid_property_is_valid-2                                              9.00 ± 0%      35.00 ± 0%          47.00 ± 0%
Draft7/additionalProperties.json:additionalProperties_can_exist_by_itself:an_additional_invalid_property_is_invalid-2                                          18.0 ± 0%       42.0 ± 0%           74.0 ± 0%
Draft7/additionalProperties.json:additionalProperties_are_allowed_by_default:additional_properties_are_allowed-2                                               17.0 ± 0%       44.0 ± 0%           88.0 ± 0%
Draft7/additionalProperties.json:additionalProperties_should_not_look_in_applicators:properties_defined_in_allOf_are_not_examined-2                            21.0 ± 0%                          116.0 ± 0%
Draft7/allOf.json:allOf:allOf-2                                                                                                                                17.0 ± 0%       74.0 ± 0%          122.0 ± 0%
Draft7/allOf.json:allOf:mismatch_second-2                                                                                                                      27.0 ± 0%       66.0 ± 0%          136.0 ± 0%
Draft7/allOf.json:allOf:mismatch_first-2                                                                                                                       30.0 ± 0%       65.0 ± 0%          146.0 ± 0%
Draft7/allOf.json:allOf:wrong_type-2                                                                                                                           30.0 ± 0%       83.0 ± 0%          142.0 ± 0%
Draft7/allOf.json:allOf_with_base_schema:valid-2                                                                                                               18.0 ± 0%       85.0 ± 0%          144.0 ± 0%
Draft7/allOf.json:allOf_with_base_schema:mismatch_base_schema-2                                                                                                21.0 ± 0%       79.0 ± 0%          143.0 ± 0%
Draft7/allOf.json:allOf_with_base_schema:mismatch_first_allOf-2                                                                                                31.0 ± 0%       78.0 ± 0%          168.0 ± 0%
Draft7/allOf.json:allOf_with_base_schema:mismatch_second_allOf-2                                                                                               33.0 ± 0%       80.0 ± 0%          174.0 ± 0%
Draft7/allOf.json:allOf_with_base_schema:mismatch_both-2                                                                                                       50.0 ± 0%       71.0 ± 0%          184.0 ± 0%
Draft7/allOf.json:allOf_simple_types:valid-2                                                                                                                   16.0 ± 0%       40.0 ± 0%          120.0 ± 0%
Draft7/allOf.json:allOf_simple_types:mismatch_one-2                                                                                                            47.0 ± 0%       46.0 ± 0%          174.0 ± 0%
Draft7/allOf.json:allOf_with_boolean_schemas,_all_true:any_value_is_valid-2                                                                                    7.00 ± 0%      40.00 ± 0%          49.00 ± 0%
Draft7/allOf.json:allOf_with_boolean_schemas,_some_false:any_value_is_invalid-2                                                                                16.0 ± 0%       45.0 ± 0%           78.0 ± 0%
Draft7/allOf.json:allOf_with_boolean_schemas,_all_false:any_value_is_invalid-2                                                                                 29.0 ± 0%       50.0 ± 0%           93.0 ± 0%
Draft7/allOf.json:allOf_with_one_empty_schema:any_data_is_valid-2                                                                                              10.0 ± 0%       30.0 ± 0%           73.0 ± 0%
Draft7/allOf.json:allOf_with_two_empty_schemas:any_data_is_valid-2                                                                                             12.0 ± 0%       39.0 ± 0%           95.0 ± 0%
Draft7/allOf.json:allOf_with_the_first_empty_schema:number_is_valid-2                                                                                          12.0 ± 0%       39.0 ± 0%          100.0 ± 0%
Draft7/allOf.json:allOf_with_the_first_empty_schema:string_is_invalid-2                                                                                        19.0 ± 0%       47.0 ± 0%           98.0 ± 0%
Draft7/allOf.json:allOf_with_the_last_empty_schema:number_is_valid-2                                                                                           12.0 ± 0%       39.0 ± 0%          100.0 ± 0%
Draft7/allOf.json:allOf_with_the_last_empty_schema:string_is_invalid-2                                                                                         19.0 ± 0%       47.0 ± 0%           98.0 ± 0%
Draft7/allOf.json:nested_allOf,_to_check_validation_semantics:null_is_valid-2                                                                                  6.00 ± 0%      43.00 ± 0%          60.00 ± 0%
Draft7/allOf.json:nested_allOf,_to_check_validation_semantics:anything_non-null_is_invalid-2                                                                   33.0 ± 0%       53.0 ± 0%          146.0 ± 0%
Draft7/allOf.json:allOf_combined_with_anyOf,_oneOf:allOf:_false,_anyOf:_false,_oneOf:_false-2                                                                   122 ± 0%         70 ± 0%            326 ± 0%
Draft7/allOf.json:allOf_combined_with_anyOf,_oneOf:allOf:_false,_anyOf:_false,_oneOf:_true-2                                                                   88.0 ± 0%       64.0 ± 0%          273.0 ± 0%
Draft7/allOf.json:allOf_combined_with_anyOf,_oneOf:allOf:_false,_anyOf:_true,_oneOf:_false-2                                                                   88.0 ± 0%       64.0 ± 0%          273.0 ± 0%
Draft7/allOf.json:allOf_combined_with_anyOf,_oneOf:allOf:_false,_anyOf:_true,_oneOf:_true-2                                                                    56.0 ± 0%       59.0 ± 0%          221.0 ± 0%
Draft7/allOf.json:allOf_combined_with_anyOf,_oneOf:allOf:_true,_anyOf:_false,_oneOf:_false-2                                                                   90.0 ± 0%       64.0 ± 0%          273.0 ± 0%
Draft7/allOf.json:allOf_combined_with_anyOf,_oneOf:allOf:_true,_anyOf:_false,_oneOf:_true-2                                                                    58.0 ± 0%       59.0 ± 0%          221.0 ± 0%
Draft7/allOf.json:allOf_combined_with_anyOf,_oneOf:allOf:_true,_anyOf:_true,_oneOf:_false-2                                                                    53.0 ± 0%       58.0 ± 0%          220.0 ± 0%
Draft7/allOf.json:allOf_combined_with_anyOf,_oneOf:allOf:_true,_anyOf:_true,_oneOf:_true-2                                                                     25.0 ± 0%       53.0 ± 0%          167.0 ± 0%
Draft7/anyOf.json:anyOf:first_anyOf_valid-2                                                                                                                    12.0 ± 0%       26.0 ± 0%           93.0 ± 0%
Draft7/anyOf.json:anyOf:second_anyOf_valid-2                                                                                                                   28.0 ± 0%       41.0 ± 0%          127.0 ± 0%
Draft7/anyOf.json:anyOf:both_anyOf_valid-2                                                                                                                     12.0 ± 0%       26.0 ± 0%           93.0 ± 0%
Draft7/anyOf.json:anyOf:neither_anyOf_valid-2                                                                                                                  63.0 ± 0%       47.0 ± 0%          181.0 ± 0%
Draft7/anyOf.json:anyOf_with_base_schema:mismatch_base_schema-2                                                                                                11.0 ± 0%       31.0 ± 0%           95.0 ± 0%
Draft7/anyOf.json:anyOf_with_base_schema:one_anyOf_valid-2                                                                                                     12.0 ± 0%       40.0 ± 0%          101.0 ± 0%
Draft7/anyOf.json:anyOf_with_base_schema:both_anyOf_invalid-2                                                                                                  27.0 ± 0%       46.0 ± 0%          132.0 ± 0%
Draft7/anyOf.json:anyOf_with_boolean_schemas,_all_true:any_value_is_valid-2                                                                                    7.00 ± 0%      27.00 ± 0%          48.00 ± 0%
Draft7/anyOf.json:anyOf_with_boolean_schemas,_some_true:any_value_is_valid-2                                                                                   7.00 ± 0%      27.00 ± 0%          48.00 ± 0%
Draft7/anyOf.json:anyOf_with_boolean_schemas,_all_false:any_value_is_invalid-2                                                                                 23.0 ± 0%       44.0 ± 0%           92.0 ± 0%
Draft7/anyOf.json:anyOf_complex_types:first_anyOf_valid_(complex)-2                                                                                            14.0 ± 0%       44.0 ± 0%          113.0 ± 0%
Draft7/anyOf.json:anyOf_complex_types:second_anyOf_valid_(complex)-2                                                                                           21.0 ± 0%       60.0 ± 0%          121.0 ± 0%
Draft7/anyOf.json:anyOf_complex_types:both_anyOf_valid_(complex)-2                                                                                             17.0 ± 0%       47.0 ± 0%          116.0 ± 0%
Draft7/anyOf.json:anyOf_complex_types:neither_anyOf_valid_(complex)-2                                                                                          43.0 ± 0%       84.0 ± 0%          162.0 ± 0%
Draft7/anyOf.json:anyOf_with_one_empty_schema:string_is_valid-2                                                                                                13.0 ± 0%       41.0 ± 0%           83.0 ± 0%
Draft7/anyOf.json:anyOf_with_one_empty_schema:number_is_valid-2                                                                                                13.0 ± 0%       27.0 ± 0%           85.0 ± 0%
Draft7/anyOf.json:nested_anyOf,_to_check_validation_semantics:null_is_valid-2                                                                                  6.00 ± 0%      35.00 ± 0%          60.00 ± 0%
Draft7/anyOf.json:nested_anyOf,_to_check_validation_semantics:anything_non-null_is_invalid-2                                                                   36.0 ± 0%       45.0 ± 0%          146.0 ± 0%
Draft7/anyOf.json:nested_anyOf,_to_check_validation_semantics:null_is_valid#01-2                                                                               6.00 ± 0%      35.00 ± 0%          60.00 ± 0%
Draft7/anyOf.json:nested_anyOf,_to_check_validation_semantics:anything_non-null_is_invalid#01-2                                                                36.0 ± 0%       45.0 ± 0%          146.0 ± 0%
Draft7/boolean_schema.json:boolean_schema_'true':number_is_valid-2                                                                                             6.00 ± 0%      17.00 ± 0%          26.00 ± 0%
Draft7/boolean_schema.json:boolean_schema_'true':string_is_valid-2                                                                                             7.00 ± 0%      18.00 ± 0%          27.00 ± 0%
Draft7/boolean_schema.json:boolean_schema_'true':boolean_true_is_valid-2                                                                                       5.00 ± 0%      16.00 ± 0%          25.00 ± 0%
Draft7/boolean_schema.json:boolean_schema_'true':boolean_false_is_valid-2                                                                                      5.00 ± 0%      16.00 ± 0%          25.00 ± 0%
Draft7/boolean_schema.json:boolean_schema_'true':null_is_valid-2                                                                                               6.00 ± 0%      17.00 ± 0%          26.00 ± 0%
Draft7/boolean_schema.json:boolean_schema_'true':object_is_valid-2                                                                                             11.0 ± 0%       22.0 ± 0%           31.0 ± 0%
Draft7/boolean_schema.json:boolean_schema_'true':empty_object_is_valid-2                                                                                       7.00 ± 0%      18.00 ± 0%          27.00 ± 0%
Draft7/boolean_schema.json:boolean_schema_'true':array_is_valid-2                                                                                              10.0 ± 0%       21.0 ± 0%           30.0 ± 0%
Draft7/boolean_schema.json:boolean_schema_'true':empty_array_is_valid-2                                                                                        7.00 ± 0%      18.00 ± 0%          27.00 ± 0%
Draft7/boolean_schema.json:boolean_schema_'false':number_is_invalid-2                                                                                          8.00 ± 0%      20.00 ± 0%          40.00 ± 0%
Draft7/boolean_schema.json:boolean_schema_'false':string_is_invalid-2                                                                                          9.00 ± 0%      21.00 ± 0%          41.00 ± 0%
Draft7/boolean_schema.json:boolean_schema_'false':boolean_true_is_invalid-2                                                                                    7.00 ± 0%      19.00 ± 0%          39.00 ± 0%
Draft7/boolean_schema.json:boolean_schema_'false':boolean_false_is_invalid-2                                                                                   7.00 ± 0%      19.00 ± 0%          39.00 ± 0%
Draft7/boolean_schema.json:boolean_schema_'false':null_is_invalid-2                                                                                            8.00 ± 0%      20.00 ± 0%          40.00 ± 0%
Draft7/boolean_schema.json:boolean_schema_'false':object_is_invalid-2                                                                                          13.0 ± 0%       25.0 ± 0%           45.0 ± 0%
Draft7/boolean_schema.json:boolean_schema_'false':empty_object_is_invalid-2                                                                                    9.00 ± 0%      21.00 ± 0%          41.00 ± 0%
Draft7/boolean_schema.json:boolean_schema_'false':array_is_invalid-2                                                                                           12.0 ± 0%       24.0 ± 0%           44.0 ± 0%
Draft7/boolean_schema.json:boolean_schema_'false':empty_array_is_invalid-2                                                                                     9.00 ± 0%      21.00 ± 0%          41.00 ± 0%
Draft7/const.json:const_validation:same_value_is_valid-2                                                                                                       12.0 ± 0%       20.0 ± 0%           63.0 ± 0%
Draft7/const.json:const_validation:another_value_is_invalid-2                                                                                                  16.0 ± 0%       26.0 ± 0%           82.0 ± 0%
Draft7/const.json:const_validation:another_type_is_invalid-2                                                                                                   10.0 ± 0%       26.0 ± 0%           73.0 ± 0%
Draft7/const.json:const_with_object:same_object_is_valid-2                                                                                                     14.0 ± 0%       44.0 ± 0%          125.0 ± 0%
Draft7/const.json:const_with_object:same_object_with_different_property_order_is_valid-2                                                                       14.0 ± 0%       44.0 ± 0%          125.0 ± 0%
Draft7/const.json:const_with_object:another_object_is_invalid-2                                                                                                15.0 ± 0%       50.0 ± 0%          130.0 ± 0%
Draft7/const.json:const_with_object:another_type_is_invalid-2                                                                                                  15.0 ± 0%       50.0 ± 0%          120.0 ± 0%
Draft7/const.json:const_with_array:same_array_is_valid-2                                                                                                       15.0 ± 0%       41.0 ± 0%          113.0 ± 0%
Draft7/const.json:const_with_array:another_array_item_is_invalid-2                                                                                             13.0 ± 0%       44.0 ± 0%          109.0 ± 0%
Draft7/const.json:const_with_array:array_with_additional_items_is_invalid-2                                                                                    17.0 ± 0%       48.0 ± 0%          117.0 ± 0%
Draft7/const.json:const_with_null:null_is_valid-2                                                                                                              6.00 ± 0%      20.00 ± 0%          51.00 ± 0%
Draft7/const.json:const_with_null:not_null_is_invalid-2                                                                                                        11.0 ± 0%       26.0 ± 0%           74.0 ± 0%
Draft7/const.json:const_with_false_does_not_match_0:false_is_valid-2                                                                                           5.00 ± 0%      18.00 ± 0%          48.00 ± 0%
Draft7/const.json:const_with_false_does_not_match_0:integer_zero_is_invalid-2                                                                                  11.0 ± 0%       25.0 ± 0%           73.0 ± 0%
Draft7/const.json:const_with_false_does_not_match_0:float_zero_is_invalid-2                                                                                    13.0 ± 0%       26.0 ± 0%           76.0 ± 0%
Draft7/const.json:const_with_true_does_not_match_1:true_is_valid-2                                                                                             5.00 ± 0%      18.00 ± 0%          48.00 ± 0%
Draft7/const.json:const_with_true_does_not_match_1:integer_one_is_invalid-2                                                                                    12.0 ± 0%       26.0 ± 0%           82.0 ± 0%
Draft7/const.json:const_with_true_does_not_match_1:float_one_is_invalid-2                                                                                      17.0 ± 0%       27.0 ± 0%           87.0 ± 0%
Draft7/const.json:const_with_[false]_does_not_match_[0]:[false]_is_valid-2                                                                                     8.00 ± 0%      24.00 ± 0%          63.00 ± 0%
Draft7/const.json:const_with_[false]_does_not_match_[0]:[0]_is_invalid-2                                                                                       13.0 ± 0%       31.0 ± 0%           83.0 ± 0%
Draft7/const.json:const_with_[false]_does_not_match_[0]:[0.0]_is_invalid-2                                                                                     14.0 ± 0%       32.0 ± 0%           85.0 ± 0%
Draft7/const.json:const_with_[true]_does_not_match_[1]:[true]_is_valid-2                                                                                       8.00 ± 0%      24.00 ± 0%          63.00 ± 0%
Draft7/const.json:const_with_[true]_does_not_match_[1]:[1]_is_invalid-2                                                                                        13.0 ± 0%       32.0 ± 0%           84.0 ± 0%
Draft7/const.json:const_with_[true]_does_not_match_[1]:[1.0]_is_invalid-2                                                                                      14.0 ± 0%       33.0 ± 0%           86.0 ± 0%
Draft7/const.json:const_with_{"a":_false}_does_not_match_{"a":_0}:{"a":_false}_is_valid-2                                                                      8.00 ± 0%      29.00 ± 0%          85.00 ± 0%
Draft7/const.json:const_with_{"a":_false}_does_not_match_{"a":_0}:{"a":_0}_is_invalid-2                                                                        13.0 ± 0%       42.0 ± 0%          105.0 ± 0%
Draft7/const.json:const_with_{"a":_false}_does_not_match_{"a":_0}:{"a":_0.0}_is_invalid-2                                                                      14.0 ± 0%       43.0 ± 0%          107.0 ± 0%
Draft7/const.json:const_with_{"a":_true}_does_not_match_{"a":_1}:{"a":_true}_is_valid-2                                                                        8.00 ± 0%      29.00 ± 0%
Draft7/const.json:const_with_{"a":_true}_does_not_match_{"a":_1}:{"a":_1}_is_invalid-2                                                                         13.0 ± 0%       43.0 ± 0%
Draft7/const.json:const_with_{"a":_true}_does_not_match_{"a":_1}:{"a":_1.0}_is_invalid-2                                                                       14.0 ± 0%       44.0 ± 0%
Draft7/const.json:const_with_0_does_not_match_other_zero-like_types:false_is_invalid-2                                                                         9.00 ± 0%      24.00 ± 0%
Draft7/const.json:const_with_0_does_not_match_other_zero-like_types:integer_zero_is_valid-2                                                                    9.00 ± 0%      18.00 ± 0%
Draft7/const.json:const_with_0_does_not_match_other_zero-like_types:float_zero_is_valid-2                                                                      11.0 ± 0%       19.0 ± 0%
Draft7/const.json:const_with_0_does_not_match_other_zero-like_types:empty_object_is_invalid-2                                                                  11.0 ± 0%       26.0 ± 0%
Draft7/const.json:const_with_0_does_not_match_other_zero-like_types:empty_array_is_invalid-2                                                                   11.0 ± 0%       26.0 ± 0%
Draft7/const.json:const_with_0_does_not_match_other_zero-like_types:empty_string_is_invalid-2                                                                  9.00 ± 0%      24.00 ± 0%
Draft7/const.json:const_with_1_does_not_match_true:true_is_invalid-2                                                                                                           25.0 ± 0%
Draft7/const.json:const_with_1_does_not_match_true:integer_one_is_valid-2                                                                                                      20.0 ± 0%
Draft7/const.json:const_with_1_does_not_match_true:float_one_is_valid-2                                                                                                        21.0 ± 0%
[Geo mean]                                                                                                                                                     14.8            35.0                86.9
```
</details>

## Conclusion

If you want a fast validator, pick `santhosh-tekuri/jsonschema` or `qri-io/jsonschema`.

If you want a correct validator, pick `santhosh-tekuri/jsonschema` or `xeipuuv/gojsonschema`.

If you want a fast and correct validator, pick `santhosh-tekuri/jsonschema`.

Thank you for reading and keep your data valid! :)
