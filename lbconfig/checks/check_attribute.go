package checks

type CheckAttribute struct {
}

func (checkAttribute CheckAttribute) Run(...interface{}) (interface{}, error) {
	// This will be used later on for the default load
	return -1, nil
}
