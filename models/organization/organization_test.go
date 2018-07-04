package organization

import (
	"testing"
)

var validateNameTests = []struct{
	name string
	valid bool
}{
	{"organization", true},
	{"organizationorganizationorganizationorganizationorganizationorganization"+
		"organizationorganizationorganizationorganization", false},
	{"spaces org", false},
	{"", false},
}

func TestValidateName(t * testing.T) {
	for _, test := range validateNameTests {
		if (validateName(test.name) == nil) != test.valid {
			t.Errorf("Received incorrect result for organization name: %v", test.name)
		}
	}
}
