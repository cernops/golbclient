package tests

import (
	"golbclient/tests/ut"
	"lbalias/utils/filehandler"
	"lbalias/utils/logger"
	"testing"
)

/**********************************************************************
							TEST-SUITE
**********************************************************************/

type UnitTest interface {
	Run(input ...interface{}) (output interface{})
}

func Suite(t *testing.T, ut UnitTest, expected interface{}, input ...interface{}) {
	dir := "configuration_files_for_testing/"

	logger.Debug("Running a unit-test for [%v]", ut)
	out := ut.Run(input)
	if out != expected {
		t.Errorf("Expected [%v] but got [%v]", expected, out)
	}
	logger.Debug("Passed for [v]", ut)
}

/**********************************************************************
							CI-TESTS (CHANGE UNIT-TEST ---> CI )

							ReadLBAliases()
							Evaluate()
**********************************************************************/

var input interface{}
var expected interface{}

func TestCollectd(t *testing.T) {
	lbconf := "lbconfig"
	input, err := filehandler.ReadFirstLineFromFile(lbconf)
	if err != nil {
		t.Errorf("An error occurred when attempting to load the configuration file [%s]", lbconf)
	}
	utest := ut.UTCollectd{}

	Suite(t, utest, expected, input)
}

func TestNologinPaulo(t *testing.T) {
	Suite(t, ut.UTNologin{}, true)
}
