package query

import (
	"fmt"
	"regexp"
	"strings"
)

// regex to capture all variations of filter string
var qre = regexp.MustCompile(`(\w+)(\=\=|\!\=|\=\=\=|\!\=\=|\~|\!\~|>|<|>\=|\=<)(\w+)(\,|\;)?`)

// Filter is a container for filter parameters
type Filter struct {
	// Field of the object on which the filter will be applied
	Field string
	// Type of filter for matching or exclusion
	Operator string
	// The value to match or exclude
	Value string
	// Logic for combining multiple filter expressions, usually "AND" or "OR"
	Logic string
}

func getOperatorMap() map[string]string {
	return map[string]string{
		"==": "==",
		"!=": "!=",
		">":  ">",
		"<":  "<",
		">=": ">=",
		"<=": "<=",
		"~":  "=~",
		"!~": "!~",
	}
}

// ParseFilterString parses a predefined filter string to Filter
// structure. The filter string specification is defined in
// corresponding protocol buffer definition.
func ParseFilterString(fstr string) ([]*Filter, error) {
	// create slice that will contain Filter structs
	var filters []*Filter
	// get all regex matches for fstr
	m := qre.FindAllStringSubmatch(fstr, -1)
	// if no matches, return empty slice
	if len(m) == 0 {
		return filters, nil
	}
	// get map of all allowed operators
	omap := getOperatorMap()
	// loop through separate items from fstr string
	for _, n := range m {
		// if no operator found in map, return slice and throw error
		if _, ok := omap[n[2]]; !ok {
			return filters, fmt.Errorf("filter operator %s not allowed", n[2])
		}
		// initialize Filter container with appropriate data
		f := &Filter{
			Field:    n[1],
			Operator: n[2],
			Value:    n[3],
		}
		if len(n) == 5 {
			f.Logic = n[4]
		}
		// add this Filter to slice
		filters = append(filters, f)
	}
	// return slice of Filter structs
	return filters, nil
}

// GenAQLFilterStatement generates an AQL(arangodb query language) compatible
// filter query statement
func GenAQLFilterStatement(fmap map[string]string, filters []*Filter) string {
	// set map for logic
	lmap := map[string]string{",": "OR", ";": "AND"}
	// get map of all allowed operators
	omap := getOperatorMap()
	// initialize variable for a string builder
	var clause strings.Builder
	// write FILTER to this string
	clause.WriteString("FILTER ")
	// loop over items in filters slice
	for _, f := range filters {
		// write the rest of AQL statement based on data
		clause.WriteString(
			fmt.Sprintf(
				"%s %s %s",
				fmap[f.Field],
				omap[f.Operator],
				checkAndQuote(f.Operator, f.Value),
			),
		)
		// if there's logic, write that too
		if len(f.Logic) != 0 {
			clause.WriteString(fmt.Sprintf(" %s ", lmap[f.Logic]))
		}
	}
	// return the string
	return clause.String()
}

// check if operator is for a string
func checkAndQuote(op, value string) string {
	if op == "===" || op == "!==" || op == "=~" || op == "!~" {
		return fmt.Sprintf("'%s'", value)
	}
	return value
}
