package ut

import (
	"lbalias/checks"
)

/**********************************************************************
						UNIT-TESTS :: COLLECTD
**********************************************************************/

type UTCollectd struct{}

func (tc UTCollectd) Run(args ...interface{}) interface{} {
	a := checks.ParamCheck{}
	a.SetCode(15)
	a.SetCommand("collectd")
	return a.Run(args)
}
