package utils

func Contains[T comparable](data []T, item T) bool {
	for _, element := range data {
		if item == element {
			return true
		}
	}

	return false
}

func FilterM[K comparable, V any](data map[K]V, f func(K, V) bool) map[K]V {
	out := make(map[K]V)
	for key, value := range data {
		if f(key, value) {
			out[key] = value
		}
	}

	return out
}

func Mapf[TI any, TO any](in []TI, f func(TI) TO) []TO {
	out := make([]TO, 0)
	for _, element := range in {
		out = append(out, f(element))
	}

	return out
}

func TransformKeys[KI, KO comparable, V any](m map[KI]V, f func(KI) KO) map[KO]V {
	mNew := make(map[KO]V)
	for k, v := range m {
		newKey := f(k)
		mNew[newKey] = v
	}

	return mNew
}

func TransformValues[K comparable, VI any, VO any](m map[K]VI, f func(VI) VO) map[K]VO {
	mNew := make(map[K]VO)
	for k, v := range m {
		mNew[k] = f(v)
	}

	return mNew
}

func TransformValuesArray[K comparable, VI any, VO any](m map[K][]VI, f func(VI) VO) map[K][]VO {
	mNew := make(map[K][]VO)
	for k, v := range m {
		values := make([]VO, 0)
		for _, value := range v {
			values = append(values, f(value))
		}
		mNew[k] = values
	}

	return mNew
}
