package main

import (
	"regexp"
)

var initDone = false

var allPlainTokenReplacementRules = make(map[string]plainTokenReplacementRule, 4)

type plainTokenReplacementRule struct {
	reg         regexp.Regexp
	replacement string
}

func _init() {
	allPlainTokenReplacementRules["NoRepeatedSingleQuotes"] = plainTokenReplacementRule{*regexp.MustCompile("''"), ""}
	allPlainTokenReplacementRules["NoRepeatedDoubleQuotes"] = plainTokenReplacementRule{*regexp.MustCompile(`""`), ""}
	allPlainTokenReplacementRules["EscapedSpaceIsJustSpace"] = plainTokenReplacementRule{*regexp.MustCompile(`\\ `), " "}
	allPlainTokenReplacementRules["UnescapeSingleQuote"] = plainTokenReplacementRule{*regexp.MustCompile(`\\'`), "'"}
	allPlainTokenReplacementRules["UnescapeDoubleQuote"] = plainTokenReplacementRule{*regexp.MustCompile(`\\"`), "\""}
	allPlainTokenReplacementRules["UnescapeRegularChar"] = plainTokenReplacementRule{*regexp.MustCompile(`\\(\w)`), "$1"}
	initDone = true
}

func sanitizePlainToken(token string) string {
	if !initDone {
		_init()
	}
	for ruleName, rule := range allPlainTokenReplacementRules {
		DbgPrintf("running rule: %s\n", ruleName)
		DbgPrintf("before: %s\n", token)
		token = rule.reg.ReplaceAllString(token, rule.replacement)
		DbgPrintf("after: %s\n", token)
	}
	return token
}
