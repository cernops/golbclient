package timer

import (
	"fmt"
	"reflect"
	"runtime"
	"time"

	logger "github.com/sirupsen/logrus"
	"gitlab.cern.ch/lb-experts/golbclient/lbconfig/utils/parser"
)

// ExecuteWithTimeoutR : Executes a function given a maximum timeout value. If the timeout value is exceeded, a error
// will be returned.
func ExecuteWithTimeoutR(timeout time.Duration, f interface{}, args ...interface{}) (ret interface{}, err error) {
	now := time.Now().UnixNano() / int64(time.Millisecond)
	fnName := getFunctionName(f)
	logger.WithFields(logger.Fields{
		"FUNCTION_W_TIMEOUT": fnName,
		"TIMEOUT_VALUE": timeout.String()},
	).Debug("Executing function...")

	r := make(chan interface{}, 1)
	e := make(chan error, 1)
	go callFunction(r, e, f, args...)
	select {
	case res := <-r:
		newNow := time.Now().UnixNano()/int64(time.Millisecond) - now
		logger.WithField("INTERNAL", "CMD_RUNNER").Debugf("Function [%s] :: Runtime: %dms", fnName, newNow)
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

// ExecuteWithTimeoutRInt : Executes a function given a maximum timeout value and returns it's output in the format
// of [int]. If the timeout value is exceeded of the function produced an error, an error will also be returned
//
func ExecuteWithTimeoutRInt(timeout time.Duration, f interface{}, args ...interface{}) (int, error) {
	value, err := ExecuteWithTimeoutR(timeout, f, args...)
	if err != nil {
		return -1, err
	}

	if parsedOutput, ok := value.([]interface{}); ok {
		if len(parsedOutput) == 0 {
			return -1, fmt.Errorf("expected output but got nothing from the function [%s]", getFunctionName(f))
		}
		return int(parser.ParseInterfaceAsInteger(parsedOutput[0])), nil
	}
	return -1, fmt.Errorf("internal error when calling the [callFunction] helper - the output is not of type "+
		"([]interface{}) but is instead [%T]", value)
}

// callFunction : reflection method used to call an interface as a function
func callFunction(r chan<- interface{}, e chan<- error, f interface{}, args ...interface{}) {
	var err error
	defer func() {
		if r := recover(); r != nil {
			if re, ok := r.(error); re != nil && ok {
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
	r <- resL
	e <- err
}

// getFunctionName : retrieves the function name signature from the given interface
func getFunctionName(i interface{}) string {
	return runtime.FuncForPC(reflect.ValueOf(i).Pointer()).Name()
}
