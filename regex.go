package enola

import (
	"regexp"

	"github.com/dlclark/regexp2"
)

// matchUsernameRegex checks username against a site regexCheck pattern.
// Patterns from data.json may use PCRE features (lookahead/lookbehind) unsupported
// by Go's regexp package, so regexp2 is used as a fallback.
func matchUsernameRegex(pattern, username string) (matched bool, err error) {
	if pattern == "" {
		return true, nil
	}

	if re, compileErr := regexp.Compile(pattern); compileErr == nil {
		return re.MatchString(username), nil
	}

	re, compileErr := regexp2.Compile(pattern, 0)
	if compileErr != nil {
		return false, compileErr
	}
	return re.MatchString(username)
}
