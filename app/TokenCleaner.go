package main

import (
	"fmt"
	"regexp"
	"strings"
)

var allPlainTokenReplacementRules = make(map[string]plainTokenReplacementRule, 6)

type plainTokenReplacementRule struct {
	reg         regexp.Regexp
	replacement string
}

func init() {
	// allPlainTokenReplacementRules["NoRepeatedSingleQuotes"] = plainTokenReplacementRule{*regexp.MustCompile("''"), ""}
	// allPlainTokenReplacementRules["NoRepeatedDoubleQuotes"] = plainTokenReplacementRule{*regexp.MustCompile(`""`), ""}
	allPlainTokenReplacementRules["EscapedSpaceIsJustSpace"] = plainTokenReplacementRule{*regexp.MustCompile(`\\ `), " "}
	allPlainTokenReplacementRules["UnescapeSingleQuote"] = plainTokenReplacementRule{*regexp.MustCompile(`\\'`), "'"}
	allPlainTokenReplacementRules["UnescapeDoubleQuote"] = plainTokenReplacementRule{*regexp.MustCompile(`\\"`), "\""}
	allPlainTokenReplacementRules["UnescapeRegularChar"] = plainTokenReplacementRule{*regexp.MustCompile(`\\(\w)`), "$1"}
}

func sanitisePlainToken(token string) string {
	for ruleName, rule := range allPlainTokenReplacementRules {
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

func sanitiseDoubleQuoteToken(dirtyCommandSegment string) string {
	return strings.ReplaceAll(dirtyCommandSegment, "\"", "")
}

func sanitiseSingleQuoteToken(dirtyCommandSegment string) string {
	return strings.ReplaceAll(dirtyCommandSegment, "'", "")
}
