package registry

import (
	"fmt"
	"regexp"
	"strings"
)

var semverRe = regexp.MustCompile(`^\d+\.\d+\.\d+$`)

// ValidationError holds all field-level errors found in a single standard.
type ValidationError struct {
	Errors []string
}

func (v *ValidationError) Error() string {
	return "validation errors:\n  - " + strings.Join(v.Errors, "\n  - ")
}

// Validate returns a ValidationError if any required field is missing or malformed.
func Validate(s *Standard) error {
	var errs []string
	if s.ID == "" {
		errs = append(errs, "missing required field: id")
	}
	if s.Title == "" {
		errs = append(errs, "missing required field: title")
	}
	if s.Version == "" {
		errs = append(errs, "missing required field: version")
	} else if !semverRe.MatchString(s.Version) {
		errs = append(errs, fmt.Sprintf("version %q is not valid semver (expected X.Y.Z)", s.Version))
	}
	if len(s.Rules) == 0 {
		errs = append(errs, "missing required field: rules (must have at least one rule)")
	}
	if len(errs) == 0 {
		return nil
	}
	return &ValidationError{Errors: errs}
}
