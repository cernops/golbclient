package ci

import (
	"testing"
)

func TestCollectdLoad(t *testing.T) {
	var myTests [3]lbTest

	myTests[0] = lbTest{"CollectdLoad", "../test/lbclient_collectd_load_single.conf", "", true, 98, nil, nil}
	myTests[1] = lbTest{"LoadConfigurationFile", "../test/lbclient_collectd_load.conf", "", true, 72, nil, nil}
	myTests[2] = lbTest{"LoadFailedConfigurationFile", "../test/lbclient_collectd_load_fail.conf", "", false, -15, nil, nil}

	//	runMultipleTests(t, false, myTests[:])
}
