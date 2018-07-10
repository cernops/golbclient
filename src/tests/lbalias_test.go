package tests

import (
	"fmt"
	"golbclient/utils"
	"lbalias/checks"
	"testing"
)

var options utils.Options

func TestMissingLBClientFile(t *testing.T) {
	options.LbAliasFile = "/path/to/nonexistent/file"
	lbAliases := utils.ReadLBAliases(options)
	fmt.Println("Got the aliases ", lbAliases)
	if lbAliases != nil {
		t.Error("I suppose I should fail")
	}
}

func TestSingleLBAliasFile(t *testing.T) {
	options.LbAliasFile = "test/lbalias_single.conf"
	lbAliases := utils.ReadLBAliases(options)
	fmt.Println("Got the aliases", lbAliases)
	if len(lbAliases) != 1 {
		t.Error("We were expecting a single lbalias definition")
	}
}

func TestDoubleLBAliasFile(t *testing.T) {
	options.LbAliasFile = "test/lbalias_double.conf"
	lbAliases := utils.ReadLBAliases(options)
	fmt.Println("Got the aliases", lbAliases)
	if len(lbAliases) != 2 {
		t.Error("We were expecting a single lbalias definition")
	}
}

// Parsing the configuration file
func TestSimpleAlias(t *testing.T) {
	lbalias := checks.LBalias{Name: "myTest",
		NoLogin:        true,
		Syslog:         true,
		ConfigFile:     "test/lbclient_single.conf",
		CheckXsessions: 0}
	lbalias.Evaluate()
	fmt.Println("Got an alias")
	fmt.Println(lbalias)
}

//Let's check that it panics if there is no configuration file
func TestMissingConfigurationFile(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("The code did not panic")
		}
	}()
	lbalias := checks.LBalias{Name: "myTest",
		NoLogin:        true,
		Syslog:         true,
		ConfigFile:     "test/lbtest.conf_does_not_exist",
		CheckXsessions: 0}

	lbalias.Evaluate()
	fmt.Println("WE MANAGED TO EVALUATE??")
}

/*
//Let's check that, given a constant load, we get that number
func ExampleConstantLoad() {
	lbalias := checks.LBalias{Name: "myTest",
		ConfigFile: "test/lbclient_constant.conf"}

	lbalias.Evaluate()
	// Output: constant
	// [add_constant] value= 249
}
*/
