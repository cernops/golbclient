package benchmarker

import (
	"gitlab.cern.ch/lb-experts/golbclient/utils/logger"
	"reflect"
	"runtime"
	"time"
)

// TimeItR : measure the runtime duration of a function by supplying the function, the precision of the measurement followed by the arguments to be passed to the function execution. This function should be used when an output is expected.
func TimeItR(precision time.Duration, f interface{},  args ...interface{}) interface{}{
	now := time.Now().UnixNano() / int64(precision)
	defer func() {
		newNow := time.Now().UnixNano() / int64(precision) - now
		logger.Info("\t Function [%s] :: Runtime: %d%s", getFunctionName(f), newNow, getNotationFromPrecision(precision))
	}()
	return callFunction(f, args...)
}

// TimeItV : measure the runtime duration of a function by supplying the function, the precision of the measurement followed by the arguments to be passed to the function execution. This function should be used when no output is expected.
func TimeItV(precision time.Duration, f interface{},  args ...interface{}) {
	now := time.Now().UnixNano() / int64(precision)
	defer func() {
		newNow := time.Now().UnixNano() / int64(precision) - now
		logger.Info("\t Function [%s] :: Runtime: %d%s", getFunctionName(f), newNow, getNotationFromPrecision(precision))
	}()
	callFunction(f, args...)
}

// callFunction : reflection method used to call an interface as a function
func callFunction(f interface{}, args ...interface{}) interface{} {
	// Account unexpected errors
	defer func() {
		if r := recover(); r != nil {
			if re, ok := r.(error); re != nil && ok {
				logger.Error("Failed to benchmark the function [%s]. Error [%s]", getFunctionName(f), re.Error())
			}
		}
	}()

	v := reflect.ValueOf(f)
	rargs := make([]reflect.Value, len(args))
	for i, a := range args {
		rargs[i] = reflect.ValueOf(a)
	}
	out := v.Call(rargs)
	var resL []interface{}
	for _, v := range out {
		resL = append(resL, v.Interface())
	}
	return resL
}

// getFunctionName : retrieves the function name signature from the given interface
func getFunctionName(i interface{}) string {
	return runtime.FuncForPC(reflect.ValueOf(i).Pointer()).Name()
}

// getNotationFromPrecision : returns a string notation based on which type of `time.duration` is given as parameter
func getNotationFromPrecision(precision time.Duration) string {
	switch precision {
	case time.Nanosecond:
		return "ns"
	case time.Second:
		return "s"
	case time.Microsecond:
		return "mu"
	case time.Hour:
		return "h"
	case time.Millisecond:
		return "ms"
	case time.Minute:
		return "min"
	default:
		return ""
	}
}