package main

import (
	"fmt"
	"regexp"
	"strings"
)

var (
	plainTokenReplacementRules       = make(map[string]plainTokenReplacementRule, 4)
	doubleQuoteTokenReplacementRules = make(map[string]plainTokenReplacementRule, 4)
)

type plainTokenReplacementRule struct {
	reg         regexp.Regexp
	replacement string
}

func init() {
	// Plain Token Rules
	plainTokenReplacementRules["EscapedSpaceIsJustSpace"] = plainTokenReplacementRule{*regexp.MustCompile(`\\ `), " "}
	plainTokenReplacementRules["UnescapeSingleQuote"] = plainTokenReplacementRule{*regexp.MustCompile(`\\'`), "'"}
	plainTokenReplacementRules["UnescapeDoubleQuote"] = plainTokenReplacementRule{*regexp.MustCompile(`\\"`), "\""}
	plainTokenReplacementRules["UnescapeRegularChar"] = plainTokenReplacementRule{*regexp.MustCompile(`\\(\w)`), "$1"}

	// DoubleQuote Rules
	doubleQuoteTokenReplacementRules["UnescapeBackslash"] = plainTokenReplacementRule{*regexp.MustCompile(`\\\\`), `\`}
	doubleQuoteTokenReplacementRules["RemoveBeginningQuote"] = plainTokenReplacementRule{*regexp.MustCompile(`^"`), ``}
	doubleQuoteTokenReplacementRules["RemoveEndingQuote"] = plainTokenReplacementRule{*regexp.MustCompile(`"$`), ``}
	doubleQuoteTokenReplacementRules["UnescapeDoubleQuote"] = plainTokenReplacementRule{*regexp.MustCompile(`\\"`), `"`}
}

func sanitisePlainToken(token string) string {
	for ruleName, rule := range plainTokenReplacementRules {
		DbgPrintf("running rule: %s\n", ruleName)
		DbgPrintf("before: %s\n", token)
		token = rule.reg.ReplaceAllString(token, rule.replacement)
		DbgPrintf("after: %s\n", token)
	}
	return token
}

func GetSanitisedCommandSegment(inputCommand string, endToken Token) string {
	dirtyCommandSegment := inputCommand[:endToken.Position]
	var cleanCommandSegment string
	switch endToken.Type {
	case Plain:
		cleanCommandSegment = sanitisePlainToken(dirtyCommandSegment)
	case SingleQuote:
		cleanCommandSegment = sanitiseSingleQuoteToken(dirtyCommandSegment)
	case DoubleQuote:
		cleanCommandSegment = sanitiseDoubleQuoteToken(dirtyCommandSegment)
	case Termination:
	default:
		fmt.Printf("!! Don't know how to sanitise token: %v !!\n", endToken)
	}
	return cleanCommandSegment
}

func sanitiseDoubleQuoteToken(token string) string {
	for ruleName, rule := range doubleQuoteTokenReplacementRules {
		DbgPrintf("running rule: %s\n", ruleName)
		DbgPrintf("before: %s\n", token)
		token = rule.reg.ReplaceAllString(token, rule.replacement)
		DbgPrintf("after: %s\n", token)
	}
	return token
}

func sanitiseSingleQuoteToken(token string) string {
	return strings.ReplaceAll(token, "'", "")
}
