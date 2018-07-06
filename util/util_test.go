package util

import (
	"testing"
)

var containsTests = []struct {
	strs     []string
	value    string
	contains bool
}{
	{[]string{}, "jack", false},
	{[]string{"notjack"}, "jack", false},
	{[]string{"jack"}, "jack", true},
	{[]string{"notjack", "jack"}, "jack", true},
	{[]string{"notjack", "jaxk"}, "jack", false},
}

func TestContains(t *testing.T) {
	for _, test := range containsTests {
		if contains := Contains(test.strs, test.value); contains != test.contains {
			t.Errorf("Test case failed for slice: %v and value: %v", test.strs, test.value)
		}
	}
}
