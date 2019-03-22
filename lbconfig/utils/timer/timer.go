package timer

import (
	"fmt"
	"gitlab.cern.ch/lb-experts/golbclient/helpers/logger"
	"reflect"
	"runtime"
	"time"
)

// ExecuteWithTimeoutR : Executes a function given a maximum timeout value. If the timeout value is exceeded, a error
// will be returned.
func ExecuteWithTimeoutR(timeout time.Duration, f interface{},  args ...interface{}) (ret interface{}, err error){
	now := time.Now().UnixNano() / int64(time.Millisecond)
	fnName := getFunctionName(f)
	defer func() {
		newNow := time.Now().UnixNano() / int64(time.Millisecond) - now
		logger.Debug("Function [%s] :: Runtime: %dms", fnName, newNow)
	}()

	r := make(chan interface{}, 1)
	e := make(chan error, 1)
	callFunction(r, e, f, args...)
	select {
	case res := <-r:
		return res, <-e
	case <-time.After(timeout):
		return nil, fmt.Errorf("the function [%s] has reached the timeout value of [%s]",
			fnName, timeout.String())
	}
}

func ExecuteWithTimeoutV(timeout time.Duration, f interface{}, args ...interface{}) error {
	_, err := ExecuteWithTimeoutR(timeout, f, args...)
	return err
}

// callFunction : reflection method used to call an interface as a function
func callFunction(r chan<- interface{}, e chan<- error, f interface{}, args ...interface{}) {
	var err error
	defer func() {
		if r := recover(); r != nil {
			if re, ok := r.(error); re != nil && ok {
				logger.Error("Failed to launch the function [%s]. Error [%s]", getFunctionName(f), re.Error())
				err = re
			}
		}
	}()

	v := reflect.ValueOf(f)
	rArgs := make([]reflect.Value, len(args))
	for i, a := range args {
		rArgs[i] = reflect.ValueOf(a)
	}
	out := v.Call(rArgs)
	var resL []interface{}
	for _, v := range out {
		if v.Interface() == nil {
			continue
		}

		if e, ok := v.Interface().(error); ok {
			err = e
		} else {
			resL = append(resL, v.Interface())
		}
	}
	r<-resL
	e<-err
}

// getFunctionName : retrieves the function name signature from the given interface
func getFunctionName(i interface{}) string {
	return runtime.FuncForPC(reflect.ValueOf(i).Pointer()).Name()
}