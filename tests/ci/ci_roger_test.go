package ci

import (
	"io/ioutil"
	"os"
	"testing"
)

func createRogerFile(t *testing.T, state string) {
	path := "/etc/roger/current.yaml"
	if _, err := os.Stat("/etc/roger"); os.IsNotExist(err) {
		err = os.Mkdir("/etc/roger", os.ModePerm)
		if err != nil {
			t.FailNow()
		}
	}
	err := ioutil.WriteFile(path, []byte("---\nappstate: "+state+"\n"), 0755)
	if err != nil {
		t.FailNow()
	}
}
func createRogerFileProduction(t *testing.T) {
	createRogerFile(t, "production")
}
func createRogerFileDraining(t *testing.T) {
	createRogerFile(t, "draining")
}

func TestRoger(t *testing.T) {
	var myTests [2]lbTest
	myTests[0] = lbTest{"LemonLoadSingle", "../test/lbclient_roger.conf", true, 42, createRogerFileProduction, nil}
	myTests[1] = lbTest{"LemonLoadSingle", "../test/lbclient_roger.conf", false, -13, createRogerFileDraining, nil}
	runMultipleTests(t, myTests[:])
}
