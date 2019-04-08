package ci

import (
	"testing"
)

func TestCommand(t *testing.T) {
	var myTests [3]lbTest

	myTests[0] = lbTest{"Command", "../test/lbclient_command.conf", true, 53, nil, nil}
	myTests[1] = lbTest{"CommandDoesNotExist", "../test/lbclient_failed_command.conf", false, -14, nil, nil}
	myTests[2] = lbTest{"CommandFail", "../test/lbclient_command_failed.conf", false, -14, nil, nil}

	runMultipleTests(t, myTests[:])
}
