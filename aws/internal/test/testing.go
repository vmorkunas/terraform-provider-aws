package test

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

// TestCheckTypeSetElemNestedAttrs is a resource.TestCheckFunc that accepts a resource
// name and flatmap style key to a schema.TypeSet attribute. The function checks
// if it appears to be a schema.TypeSet and then verifies that an element in
// the set matches all nested attribute/value pairs.
//
// Use this function over SDK provided TestCheckFunctions when validating a
// TypeSet where its elements are a nested object with their own attrs/values.
//
// Please note, if the provided value map is not granular enough, there exists
// the possibility you match an element you were not intending to, in the TypeSet.
// Provide a full mapping of attributes to be sure the unique element exists.
func TestCheckTypeSetElemNestedAttrs(name, key string, values map[string]string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		ms := s.RootModule()
		rs, ok := ms.Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s in %s", name, ms.Path)
		}

		is := rs.Primary
		if is == nil {
			return fmt.Errorf("No primary instance: %s in %s", name, ms.Path)
		}

		// A TypeSet should have a special count attribute
		countStr, ok := is.Attributes[key+".#"]
		if !ok {
			return fmt.Errorf("%q %q does not appear to be a TypeSet", name, key)
		}
		count, err := strconv.ParseInt(countStr, 10, 64)
		if err != nil {
			return err
		}

		// unflatten the TypeSet from State
		passedKeyParts := strings.Split(key, ".")
		elements := make(map[string]map[string]string, count)
		for stateKey, stateValue := range is.Attributes {
			stateKeyParts := strings.Split(stateKey, ".")
			if strings.HasPrefix(stateKey, key) {
				id := stateKeyParts[len(passedKeyParts)]
				if id != "#" {
					element, ok := elements[id]
					if !ok {
						elements[id] = make(map[string]string)
						element = elements[id]
					}
					element[strings.Join(stateKeyParts[len(passedKeyParts)+1:], ".")] = stateValue
					elements[id] = element
				}
			}
		}

		// sanity check
		if len(elements) != int(count) {
			return fmt.Errorf("Expected the number of set items to be %d, got %d.\nState: %#v", count, len(elements), is.Attributes)
		}

		// check if an element is a full match with the passed values map
		for _, element := range elements {
			var matches int
			for k, v := range values {
				if stateValue, keyExists := element[k]; keyExists && stateValue == v {
					matches++
				}
			}
			if matches == len(values) {
				return nil
			}
		}

		return fmt.Errorf("%q %q has no element with attr/value pairs: %#v in state: %#v", name, key, values, is.Attributes)
	}
}

// TestCheckTypeSetElemAttr is a resource.TestCheckFunc that accepts a resource
// name and flatmap style key to a schema.TypeSet attribute. The function checks
// if it appears to be a schema.TypeSet and then verifies that an element in
// the set matches the passed value.
//
// Use this function over SDK provided TestCheckFunctions when validating a
// TypeSet where its elements are a simple value
func TestCheckTypeSetElemAttr(name, key, value string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		ms := s.RootModule()
		rs, ok := ms.Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s in %s", name, ms.Path)
		}

		is := rs.Primary
		if is == nil {
			return fmt.Errorf("No primary instance: %s in %s", name, ms.Path)
		}

		// A TypeSet should have a special count attribute
		if _, ok := is.Attributes[key+".#"]; !ok {
			return fmt.Errorf("%q %q does not appear to be a TypeSet", name, key)
		}

		for stateKey, stateValue := range is.Attributes {
			parts := strings.Split(stateKey, ".")
			// ensure the passed key is in fact the direct path to the supposed
			// TypeSet and the values match
			if stateValue == value && key == strings.Join(parts[:len(parts)-1], ".") {
				return nil
			}
		}

		return fmt.Errorf("%q %q has no element with value: %q in state: %#v", name, key, value, is.Attributes)
	}
}
