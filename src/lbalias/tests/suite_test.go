package lbalias

import (
	"lbalias"
	"testing"
)

func Suite(t *testing.T, impl lbalias.CLI) {
	expected := false
	res, err := 
	if err != nil {
		t.Errorf("Unexpected result from the `Check.Run` implementation [%s]", err.Error())
	} else if res != expected {
		t.Errorf("Expected [%v] but got [%v]", expected, res)
	}
}
