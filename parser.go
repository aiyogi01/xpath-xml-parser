package main

import (
	"errors"
	"fmt"
	"reflect"
	"sort"
)

// ---------------------------------------------------------------------------
//	Extractor functions
// ---------------------------------------------------------------------------

// sortUniqueString sorts a slice of strings and removes duplicates
func sortUniqueString(seq []string) []string {
	if len(seq) <= 1 {
		return seq
	}
	sort.Strings(seq)
	var lastAdded = seq[0]
	var unique = []string{lastAdded}
	for _, elem := range seq[1:] {
		if elem != lastAdded {
			lastAdded = elem
			unique = append(unique, elem)
		}
	}
	return unique
}

// extractString extracts the inner text or the raw XML string of the node
func extractString(node *Node, extract string) string {
	switch extract {
	default:
		return node.InnerText()
	case "xml":
		return node.InnerXML()
	}
}

// findOneText applies the XPath expression to the node and returns
// the inner text or the raw XML string of the first match
func findOneText(node *Node, xpath *XPath, extract string) string {
	match := xpath.FindOne(node)
	if match == nil {
		return ""
	}
	return extractString(match, extract)
}

// findText applies the XPath expression to the node and returns
// a slice of the inner texts or of the raw XML strings of all matches
func findText(node *Node, xpath *XPath, extract string) []string {
	matches := xpath.Find(node)
	var text []string
	for _, match := range matches {
		text = append(text, extractString(match, extract))
	}
	return text
}

// findOneStruct applies the XPath expression to the node,
// unmarshalls the first match into a structures of type t,
// and returns a pointer to the structure
func findOneStruct(node *Node, xpath *XPath, t reflect.Type) (interface{}, error) {
	match := xpath.FindOne(node)
	ptr := reflect.New(t).Interface()
	err := Unmarshall(match, ptr)
	if err != nil {
		return nil, err
	}
	return ptr, nil
}

// findStruct applies the XPath expression to the node,
// unmarshalls all matches into structures of type t,
// and returns a slice of pointers to the structures
func findStruct(node *Node, xpath *XPath, t reflect.Type) ([]interface{}, error) {
	matches := xpath.Find(node)
	var ptrs []interface{}
	for _, match := range matches {
		ptr := reflect.New(t).Interface()
		err := Unmarshall(match, ptr)
		if err != nil {
			return nil, err
		}
		ptrs = append(ptrs, ptr)
	}
	return ptrs, nil
}

// ---------------------------------------------------------------------------
//	Parser
// ---------------------------------------------------------------------------

// Unmarshall unmarshalls an XML node into a structure v
func Unmarshall(node *Node, v interface{}) error {

	pointer := reflect.ValueOf(v)
	if pointer.Kind() != reflect.Ptr {
		return errors.New("non-pointer passed to Unmarshall")
	}

	value := pointer.Elem()
	if value.Kind() != reflect.Struct {
		return errors.New("pointer to non-struct passed to Unmarshall")
	}

	for i := 0; i < value.NumField(); i++ {
		field := value.Field(i)
		fieldInfo := value.Type().Field(i)

		xpath, err := CompileXPath(fieldInfo.Tag.Get("xpath"))
		if err != nil {
			return fmt.Errorf("can't compile xpath expression in struct '%s', field '%s': %s",
				value.Type().Name(), fieldInfo.Name, fieldInfo.Tag.Get("xpath"))
		}

		switch field.Kind() {
		default:
			return fmt.Errorf("unsupported type in struct '%s' for field '%s': %s",
				value.Type().Name(), fieldInfo.Name, field.Type())

		case reflect.String:
			match := findOneText(node, xpath, fieldInfo.Tag.Get("extract"))
			field.Set(reflect.ValueOf(match))

		case reflect.Struct:
			match, err := findOneStruct(node, xpath, field.Type())
			if err != nil {
				return err
			}
			field.Set(reflect.ValueOf(match).Elem())

		case reflect.Slice:
			elemType := field.Type().Elem()

			switch elemType.Kind() {
			case reflect.String:
				matches := findText(node, xpath, fieldInfo.Tag.Get("extract"))
				field.Set(reflect.ValueOf(matches))

			default:
				matches, err := findStruct(node, xpath, elemType)
				if err != nil {
					return err
				}
				for _, match := range matches {
					field.Set(reflect.Append(field, reflect.ValueOf(match).Elem()))
				}
			}
		}
	}
	return nil
}
