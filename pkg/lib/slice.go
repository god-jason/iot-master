package lib

func HasIntersection[T comparable](aa []T, bb []T) bool {
	for _, a := range aa {
		for _, b := range bb {
			if a == b {
				return true
			}
		}
	}
	return false
}

func Intersection[T comparable](aa []T, bb []T) []T {
	var ret []T
	for _, a := range aa {
		for _, b := range bb {
			if a == b {
				ret = append(ret, a)
			}
		}
	}
	return ret
}

func Union[T comparable](aa []T, bb []T) []T {
	var ret []T
	ret = append(ret, aa...)
	for _, b := range bb {
		for _, a := range ret {
			if a != b {
				ret = append(ret, b)
			}
		}
	}
	return ret
}
