package test

import (
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

// TestCheckTypeSetElemNestedAttrs is a resource.TestCheckFunc that accepts a resource
// name, an attribute name and depth which targets a TypeSet, as well as a value
// map to verify. The function verifies that the TypeSet attribute exists, and that
// an element matches all the values in the map.
//
// Use this function over SDK provided TestCheckFunctions when validating a
// TypeSet where its elements are a nested object with their own attrs/values.
//
// Please note, if the provided value map is not granular enough, there exists
// the possibility you match an element you were not intending to, in the TypeSet.
// Provide a full mapping of attributes to be sure the unique element exists.
func TestCheckTypeSetElemNestedAttrs(resourceName, attrName string, depth int, values map[string]string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		ms := s.RootModule()
		rs, ok := ms.Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s in %s", resourceName, ms.Path)
		}

		is := rs.Primary
		if is == nil {
			return fmt.Errorf("No primary instance: %s in %s", resourceName, ms.Path)
		}

		matches := make(map[string]int)

		for stateKey, stateValue := range is.Attributes {
			parts := strings.Split(stateKey, ".")
			// a Set/List item with nested attrs would have a flatmap address of
			// at least length 3
			// foo.0.name = "bar"
			d := len(parts) - 3
			if d < 0 {
				continue
			}
			attr := parts[d]
			if attr == attrName && d == depth-1 {
				// ensure this is a Set/List
				if _, exists := is.Attributes[strings.Join(parts[:d+1], ".")+".#"]; !exists {
					return fmt.Errorf("%q attr %q is not TypeSet", resourceName, attrName)
				}
				elementId := parts[d+1]
				nestedAttr := strings.Join(parts[d+2:], ".")
				// check if the nestedAttr exists in the passed values map
				// if it does, and matches, increment the matches count
				if v, exists := values[nestedAttr]; exists && stateValue == v {
					matches[elementId] = matches[elementId] + 1
					// exit if there is an element that is a full match
					if matches[elementId] == len(values) {
						return nil
					}
				}
			}
		}

		return fmt.Errorf("No TypeSet element in %q with attr name %q at depth %d, with nested attrs %#v in state: %#v", resourceName, attrName, depth, values, is.Attributes)
	}
}

// TestCheckTypeSetElemAttr is a resource.TestCheckFunc that accepts a resource
// name, an attribute name and depth which targets a TypeSet, as well as a value
// to verify. The function verifies that the TypeSet attribute exists, and that
// an element matches the passed value.
//
// Use this function over SDK provided TestCheckFunctions when validating a
// TypeSet where its elements are a simple value
func TestCheckTypeSetElemAttr(resourceName, attrName string, depth int, value string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		ms := s.RootModule()
		rs, ok := ms.Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s in %s", resourceName, ms.Path)
		}

		is := rs.Primary
		if is == nil {
			return fmt.Errorf("No primary instance: %s in %s", resourceName, ms.Path)
		}

		for stateKey, stateValue := range is.Attributes {
			parts := strings.Split(stateKey, ".")
			// a Set/List item would have a flatmap address of at least length 2
			// foo.0 = "bar"
			d := len(parts) - 2
			if d < 0 {
				continue
			}
			attr := parts[d]
			if attr == attrName && d == depth-1 && stateValue == value {
				// ensure this is a Set/List
				if _, exists := is.Attributes[strings.Join(parts[:d+1], ".")+".#"]; !exists {
					return fmt.Errorf("%q attr %q is not TypeSet", resourceName, attrName)
				} else {
					return nil
				}
			}
		}

		return fmt.Errorf("No TypeSet element in %q with attr name %q at depth %d, with value %q in state: %#v", resourceName, attrName, depth, value, is.Attributes)
	}
}
