package main

import (
	"fmt"
	"io/ioutil"
	"testing"
)

var expectedDescriptions = map[string]bool{
	"ACTUK4850": true,
	"ACTUK4620": false,
}

func TestGetDescription(t *testing.T) {
	for fn, hasDesc := range expectedDescriptions {
		page, err := ioutil.ReadFile(fmt.Sprintf("./test_files/%s.html", fn))
		if err != nil {
			t.Fatal(err)
		}

		t.Error(parsePage(page))
		if (parsePage(page) == "") == hasDesc {
			if hasDesc {
				t.Errorf("Expected to find a description in %s", fn)
			} else {
				t.Errorf("Did not expected to find a description in %s", fn)
			}
		}
	}
}
