package param

import logger "github.com/sirupsen/logrus"

type Parameterized interface {
	Run(contextLogger *logger.Entry, metrics []string, valueList *map[string]interface{}) error
	Name() string
}
