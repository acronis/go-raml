package raml

func IsOverridableEnum(source []*Node, target []*Node) bool {
	// Source enum must be an intersection of target enum
	for _, v := range source {
		found := false
		for _, e := range target {
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
