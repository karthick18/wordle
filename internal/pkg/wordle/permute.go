package wordle

func Permutations(word string, filter func(string) bool) []string {
	var res []string

	inc := func(idx []int) {
		for i := len(idx) - 1; i >= 0; i-- {
			if i == 0 || idx[i] < len(idx)-i-1 {
				idx[i]++

				return
			}
			idx[i] = 0
		}
	}

	for idx := make([]int, len(word)); idx[0] < len(word); inc(idx) {
		p := permutation(idx, word)
		if filter != nil && !filter(p) {
			continue
		}

		res = append(res, p)
	}

	return res
}

func permutation(idx []int, word string) string {
	res := make([]byte, len(word))
	copy(res, []byte(word))

	for o, v := range idx {
		res[o], res[o+v] = res[o+v], res[o]
	}

	return string(res)
}
