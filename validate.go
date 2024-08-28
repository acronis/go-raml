package raml

func IsCompatibleEnum(source Nodes, target Nodes) bool {
	// Target enum must be a subset of source enum
	for _, v := range target {
		found := false
		for _, e := range source {
			if v.Value == e.Value {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}
