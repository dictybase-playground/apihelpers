package query

import (
	"fmt"
	"regexp"
	"strings"
)

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
	var filters []*APIFilter
	m := re.FindAllStringSubmatch(fstr, -1)
	if len(m) == 0 {
		return filters, nil
	}
	omap := getOperatorMap()
	for _, n := range m {
		if ok := omap[n[2]]; !ok {
			return filters, fmt.Errorf("filter operator %s not allowed", n[2])
		}
		f := Filter{
			Field:    n[1],
			Operator: n[2],
			Value:    n[3],
		}
		if len(n) == 5 {
			f.Logic = n[4]
		}
		filters = append(filters, f)
	}
	return filters, nil
}

// GenAQLFilterStatement generates an AQL(arangodb query language) compatible
// filter query statement
func GenAQLFilterStatement(fmap map[string]string, filters []*Filter) string {
	lmap := map[string]string{",": "OR", ";": "AND"}
	omap := getOperatorMap()
	var clause strings.Builder
	clause.WriteString("FILTER ")
	for i, f := range filters {
		clause.WriteString(
			fmt.Sprintf(
				"%s %s %s",
				fmap[f.Attribute],
				omap[f.Operator],
				checkAndQuote(f.Operator, f.Value),
			),
		)
		if len(f.Logic) != 0 {
			clause.WriteString(fmt.Sprintf(" %s ", lmap[f.Logic]))
		}
	}
	return clause.String()
}

func checkAndQuote(op, value string) string {
	if op == "===" || op == "!==" || op == "=~" || op == "!~" {
		return fmt.Sprintf("'%s'", value)
	}
	return value
}
