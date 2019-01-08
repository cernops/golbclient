package parser

import (
	"fmt"
	"lbalias/utils/logger"
	"math"
	"strconv"
	"strings"
)

// ParseSciNumber : parses a number in scientific notation (e.g., 1.26e19)
func ParseSciNumber(str string, logErrors bool) (float64, error) {
	var err error
	if logErrors {
		defer func() {
			if err != nil {
				logger.Error("Failed to parse value [%s] with the error [%s]", str, err.Error())
			}
		}()
	}

	val, err := strconv.ParseFloat(str, 64)
	if err == nil {
		return val, nil
	}

	// Parse with scientific notation (e.g., 1.28e+12)
	sciPos := strings.IndexAny(str, "eE")
	if sciPos >= 0 {
		leftValue, err := strconv.ParseFloat(str[0:sciPos], 64)
		if err != nil {
			return -1, err
		}

		rightValue, err := strconv.ParseFloat(str[(sciPos+1):], 64)
		if err != nil {
			return -1, err
		}

		return leftValue * math.Pow10(int(rightValue)), nil
	}

	// Normal ParseFloat error
	return -1, err
}

// ParseInterfaceAsBool : returns a boolean value from a given interface object
func ParseInterfaceAsBool(obj interface{}) bool {
	result := false
	// Prevent the panic
	defer func() {
		if r := recover(); r != nil {
			logger.Debug("Recovered from an unexpected exception when trying to parse a boolean from the value [%s]", obj)
			result = false
		}
	}()

	if b, ok := obj.(bool); ok {
		result = b
	} else if i, ok := obj.(int); ok {
		result = i > 0
	} else if i, ok := obj.(int8); ok {
		result = i > 0
	} else if i, ok := obj.(int16); ok {
		result = i > 0
	} else if i, ok := obj.(int32); ok {
		result = i > 0
	} else if i, ok := obj.(int64); ok {
		result = i > 0
	} else if f, ok := obj.(float32); ok {
		result = f > 0
	} else if f, ok := obj.(float64); ok {
		result = f > 0
	} else if s, ok := obj.(string); ok {
		parsedBool, err := strconv.ParseBool(s)
		if err == nil {
			result = parsedBool
		} else {
			parsedFloat, err := strconv.ParseFloat(s, 64)
			if err == nil {
				result = parsedFloat > 0
			} else {
				result = false
			}
		}
	} else if o, ok := obj.(byte); ok {
		parsedBool, err := strconv.ParseBool(string(o))
		if err == nil {
			result = parsedBool
		} else {
			result = false
		}
	}

	return result
}

// ParseInterfaceArrayAsBool : returns a boolean value from a given array of interface objects (only returns true if all are true)
func ParseInterfaceArrayAsBool(obj ...interface{}) bool {
	for o := range obj {
		if !ParseInterfaceAsBool(o) {
			return false
		}
	}

	return true && obj != nil
}

// ParseInterfaceAsInteger : returns an integer value from a given interface object
func ParseInterfaceAsInteger(obj interface{}) int32 {
	var result int32
	result = -1
	// Prevent the panic
	defer func() {
		if r := recover(); r != nil {
			logger.Debug("Recovered from an unexpected exception when trying to parse an integer from the value [%s]", obj)
			result = -1
		}
	}()

	if b, ok := obj.(bool); ok {
		if b {
			result = 1
		} else {
			result = -1
		}
	} else if i, ok := obj.(int); ok {
		result = int32(i)
	} else if i, ok := obj.(int8); ok {
		result = int32(i)
	} else if i, ok := obj.(int16); ok {
		result = int32(i)
	} else if i, ok := obj.(int32); ok {
		result = i
	} else if i, ok := obj.(int64); ok {
		result = int32(i)
	} else if f, ok := obj.(float32); ok {
		result = int32(f)
	} else if f, ok := obj.(float64); ok {
		result = int32(f)
	} else if s, ok := obj.(string); ok {
		parsedBool, err := strconv.ParseBool(s)
		if err == nil {
			if parsedBool {
				result = 1
			} else {
				result = -1
			}
		} else {
			parsedFloat, err := strconv.ParseFloat(s, 64)
			if err == nil {
				result = int32(parsedFloat)
			} else {
				result = -1
			}
		}
	} else if o, ok := obj.(byte); ok {
		parsedBool, err := strconv.ParseBool(string(o))
		if err == nil {
			if parsedBool {
				result = 1
			} else {
				result = -1
			}
		} else {
			result = -1
		}
	}

	return result
}

// ParseInterfaceAsFloat : returns an integer value from a given interface object
func ParseInterfaceAsFloat(obj interface{}) float32 {
	var result float32
	result = -1
	// Prevent the panic
	defer func() {
		if r := recover(); r != nil {
			logger.Debug("Recovered from an unexpected exception when trying to parse an integer from the value [%s]", obj)
			result = -1
		}
	}()

	if b, ok := obj.(bool); ok {
		if b {
			result = 1
		} else {
			result = -1
		}
	} else if i, ok := obj.(int); ok {
		result = float32(i)
	} else if i, ok := obj.(int8); ok {
		result = float32(i)
	} else if i, ok := obj.(int16); ok {
		result = float32(i)
	} else if i, ok := obj.(int32); ok {
		result = float32(i)
	} else if i, ok := obj.(int64); ok {
		result = float32(i)
	} else if f, ok := obj.(float32); ok {
		result = f
	} else if f, ok := obj.(float64); ok {
		result = float32(f)
	} else if s, ok := obj.(string); ok {
		parsedBool, err := strconv.ParseBool(s)
		if err == nil {
			if parsedBool {
				result = 1
			} else {
				result = -1
			}
		} else {
			parsedFloat, err := strconv.ParseFloat(s, 64)
			if err == nil {
				result = float32(parsedFloat)
			} else {
				result = -1
			}
		}
	} else if o, ok := obj.(byte); ok {
		parsedBool, err := strconv.ParseBool(string(o))
		if err == nil {
			if parsedBool {
				result = 1
			} else {
				result = -1
			}
		} else {
			result = -1
		}
	}

	return result
}

// ParseInterfaceAsType : returns an interface value, with the desired type, from a given interface object
func ParseInterfaceAsType(obj interface{}, t func(interface{}) interface{}) (interface{}, error) {
	// t.Call([]reflect.Value{reflect.ValueOf(i)})[0].Interface()
	var result interface{}
	var err error

	result = -1
	// Prevent the panic
	defer func() {
		if r := recover(); r != nil {
			logger.Debug("Recovered from an unexpected exception when trying to parse the type [%T] from the value [%s]", t, obj)
			result = -1
		}
	}()

	if b, ok := obj.(bool); ok {
		if b {
			result = 1
		} else {
			result = -1
		}
	} else if i, ok := obj.(int); ok {
		result = t(i)
	} else if i, ok := obj.(int8); ok {
		result = t(i)
	} else if i, ok := obj.(int16); ok {
		result = t(i)
	} else if i, ok := obj.(int32); ok {
		result = i
	} else if i, ok := obj.(int64); ok {
		result = t(i)
	} else if f, ok := obj.(float32); ok {
		result = t(f)
	} else if f, ok := obj.(float64); ok {
		result = t(f)
	} else if s, ok := obj.(string); ok {
		parsedBool, err := strconv.ParseBool(s)
		if err == nil {
			if parsedBool {
				result = 1
			} else {
				result = -1
			}
		} else {
			parsedFloat, err := strconv.ParseFloat(s, 64)
			if err == nil {
				result = t(parsedFloat)
			} else {
				result = -1
			}
		}
	} else if o, ok := obj.(byte); ok {
		parsedBool, err := strconv.ParseBool(string(o))
		if err == nil {
			if parsedBool {
				result = 1
			} else {
				result = -1
			}
		} else {
			result = -1
		}
	} else {
		// Return the same value if it could not be parsed
		logger.Error("Failed to parse the interface of type [%T] from the value [%v]", t, obj)
		err = fmt.Errorf("Failed to parse the interface of type [%T] from the value [%v]", t, obj)
		result = obj
	}
	return result, err
}
