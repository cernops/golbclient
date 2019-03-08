package param

type Parameterized interface {
	Run(metrics []string, valueList *map[string]interface{}) error
	Name() string
}
