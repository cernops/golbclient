package ut

import (
	"lbalias/checks"
)

/**********************************************************************
						UNIT-TESTS :: COLLECTD
**********************************************************************/

type UTNologin struct{}

func (tc UTCNologin) Run(args ...interface{}) interface{} {
	touch file
	run // no negative || (positive && load constant)
	rm
	run // return .code of nologin
}
