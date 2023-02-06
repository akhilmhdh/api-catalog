package modules

import (
	"errors"
	"regexp"

	"github.com/dop251/goja"
	pluralize "github.com/gertd/go-pluralize"
)

var snakeCaseRegex = regexp.MustCompile("^[a-z0-9]+(?:_[a-z0-9]+)*$")
var camelCaseRegex = regexp.MustCompile("^[a-z]+(?:[A-Z0-9]+[a-z0-9]+[A-Za-z0-9]*)*$")
var pascalCaseRegex = regexp.MustCompile("^(?:[A-Z][a-z0-9]+)(?:[A-Z]+[a-z0-9]*)*$")
var kebabCaseRegex = regexp.MustCompile("^[a-z0-9]+(?:-[a-z0-9]+)*$")

var errUnknownCasing = errors.New("unknown casing")

func caseChecker(casing string, val string) (bool, error) {
	switch casing {
	case "camelcase":
		return camelCaseRegex.MatchString(val), nil
	case "snakecase":
		return snakeCaseRegex.MatchString(val), nil
	case "pascalcase":
		return pascalCaseRegex.MatchString(val), nil
	case "kebabcase":
		return kebabCaseRegex.MatchString(val), nil
	default:
		return false, errUnknownCasing
	}
}

// for checking various operations in strings
// like casing, plural etc
func (m *ModuleLoader) stringModule() goja.Value {
	obj := m.runtime.NewObject()
	pluralize := pluralize.NewClient()

	obj.Set("__esModule", true)
	obj.Set("isCasing", func(casing string, val string) bool {
		truthy, err := caseChecker(casing, val)
		if err != nil {
			panic(m.runtime.NewGoError(err))
		}
		return truthy
	})
	obj.Set("isPlural", func(val string) bool {
		return pluralize.IsPlural(val)
	})
	obj.Set("isSingular", func(val string) bool {
		return pluralize.IsSingular(val)
	})

	obj.Set("pluralize", func(val string) string {
		return pluralize.Plural(val)
	})
	obj.Set("singular", func(val string) string {
		return pluralize.Singular(val)
	})

	return obj
}
