package ci

import (
	"testing"
)

func TestLemonLoad(t *testing.T) {
	var myTests [3]lbTest
	myTests[0] = lbTest{"LemonLoadSingle", "../test/lbclient_lemon_load_single.conf", true, 1, nil, nil}
	myTests[1] = lbTest{"LemonLoad", "../test/lbclient_lemon_load.conf", true, 27, nil, nil}
	myTests[2] = lbTest{"LemonFailed", "../test/lbclient_lemon_check_fail.conf", false, -12, nil, nil}
	runMultipleTests(t, myTests[:])
}
