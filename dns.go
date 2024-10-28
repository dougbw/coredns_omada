package coredns_omada

import (
	"regexp"
	"strings"
)

func makeDNSSafe(input string) string {
	// Replace any whitespace with hyphens
	input = strings.Replace(input, " ", "-", -1)

	// Remove any characters that are not letters, numbers, or hyphens
	re := regexp.MustCompile("[^a-zA-Z0-9-]")
	input = re.ReplaceAllString(input, "")

	// Convert the string to lower case
	input = strings.ToLower(input)

	return input
}
