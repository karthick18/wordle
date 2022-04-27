package wordle

import (
	//	"fmt"
	"math/rand"
	"time"
)

type guessWork struct {
	handle        *HandleImplementor
	wordLen       int
	shuffle       []byte
	usedMap       []int
	currentStatus []int
}

var (
	mostUsed = []byte{'a', 'e', 't', 's', 'd', 'm', 'p', 'c'}
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func newGuess(handle *HandleImplementor, wordLen int) *guessWork {
	shuffle := make([]byte, len(mostUsed))
	copy(shuffle, mostUsed)

	usedMap := make([]int, 26)

	rand.Shuffle(len(shuffle), func(i, j int) {
		shuffle[i], shuffle[j] = shuffle[j], shuffle[i]
	})

	shuffle = shuffle[:wordLen]

	g := &guessWork{
		handle:  handle,
		wordLen: wordLen,
		shuffle: shuffle,
		usedMap: usedMap,
	}

	g.mark(shuffle)

	return g
}

func (g *guessWork) get() string {
	return string(g.shuffle)
}

func (g *guessWork) accept(part, word string, offset int) bool {
	for i, v := range part {
		// has an unused byte
		if g.usedMap[byte(v)-97] == 3 {
			return false
		}

		if byte(v) == g.shuffle[offset+i] && g.currentStatus[offset+i] < 2 {
			return false
		}

		if byte(v) != g.shuffle[offset+i] && g.currentStatus[offset+i] == 2 {
			return false
		}

		// if there is a position mismatch, this has to be found in the remaining
		if g.currentStatus[offset+i] == 1 {
			matched := false

			for j, b := range word {
				if j == offset+i {
					continue
				}

				if byte(b) == g.shuffle[offset+i] {
					matched = true
					break
				}
			}

			if !matched {
				return false
			}
		}
	}

	return true

}
func (g *guessWork) next(status []int) string {
	g.currentStatus = status

	shuffle := make([]byte, g.wordLen)
	emptySlots := []int{}
	matched := 0

	for i := 0; i < g.wordLen; i++ {
		v := g.currentStatus[i]
		if v > 0 {
			g.usedMap[int(g.shuffle[i])-97] = 2
			shuffle[i] = g.shuffle[i]
			matched++

			continue
		}

		// mark entry as removed from usedmap
		// make sure we don't delete a duplicate byte whose status has been marked as 1/2 and 0
		if g.usedMap[int(g.shuffle[i])-97] != 2 {
			g.usedMap[int(g.shuffle[i])-97] = 3
		}

		emptySlots = append(emptySlots, i)
	}

	if string(shuffle) == string(g.shuffle) {
		return string(shuffle)
	}

	// check for consecutive matches and try autocomplete with filter
	// check for the first consecutive match start and end location
	startLocation, endLocation := -1, -1

	for i := 0; i < len(g.shuffle)-1; i++ {
		if g.currentStatus[i] == 2 && g.currentStatus[i+1] == 2 {
			if startLocation < 0 {
				startLocation = i
			}
			if endLocation < 0 {
				endLocation = i + 1
			} else if i == endLocation {
				endLocation = i + 1
			}
		}
	}

	if startLocation >= 0 {
		completions := g.handle.AutoCompleteSubstring(string(g.shuffle[startLocation:endLocation+1]), startLocation, g.wordLen,
			func(word string) bool {
				prefix := word[:startLocation]
				suffix := word[endLocation+1:]
				seenMap := make(map[rune]bool)

				for _, b := range word {
					if seenMap[b] {
						return false
					}

					seenMap[b] = true
				}

				if len(prefix) > 0 {
					if !g.accept(prefix, word, 0) {
						return false
					}
				}

				if len(suffix) > 0 {
					if !g.accept(suffix, word, endLocation+1) {
						return false
					}
				}

				return true
			})

		switch {
		case len(completions) > 1:
			res := rand.Intn(len(completions))
			g.shuffle = []byte(completions[res])
			g.mark(g.shuffle)

			return string(g.shuffle)

		case len(completions) == 1:
			g.shuffle = []byte(completions[0])
			g.mark(g.shuffle)

			return string(g.shuffle)

		default:
			break
		}
	}

	eligible := []byte{}
	needed := g.wordLen - matched

	if needed > 0 {
		for i, v := range g.usedMap {
			if v > 0 {
				continue
			}

			eligible = append(eligible, byte(97+i))
		}
	}

	// nothing left to try
	if len(eligible) < needed {
		//		fmt.Println("eligible set", eligible, "needed", needed)

		return ""
	}

	seenMap := make(map[int]bool)

	for i := 0; i < needed; i++ {
		slot := emptySlots[i]

		for {
			eligibleIndex := rand.Intn(len(eligible))
			if seenMap[eligibleIndex] {
				continue
			}

			seenMap[eligibleIndex] = true
			shuffle[slot] = eligible[eligibleIndex]
			break
		}
	}

	allowedIndicesMap := make(map[int][]int)

	for i := range shuffle {
		v := g.currentStatus[i]
		if v == 2 {
			allowedIndicesMap[i] = append(allowedIndicesMap[i], i)
		} else {
			for j := range shuffle {
				if i == j {
					continue
				}

				if g.currentStatus[j] == 2 {
					continue
				}

				allowedIndicesMap[i] = append(allowedIndicesMap[i], j)
			}

			if len(allowedIndicesMap[i]) == 0 {
				allowedIndicesMap[i] = append(allowedIndicesMap[i], i)
			}
		}
	}

	shuffleBuffer := make([]byte, len(shuffle))
	copy(shuffleBuffer, shuffle)

	// for i := range shuffle {
	// 	fmt.Println("allowed index map", i, allowedIndicesMap[i])
	// }

	seenMap = make(map[int]bool)

	for i := range shuffle {
		var shuffleIndex int

		for {
			allowedIndex := rand.Intn(len(allowedIndicesMap[i]))
			shuffleIndex = allowedIndicesMap[i][allowedIndex]
			if seenMap[shuffleIndex] {
				//fmt.Println("shuffle index", shuffleIndex, "seen")
				continue
			}

			seenMap[shuffleIndex] = true
			shuffleBuffer[i] = shuffle[shuffleIndex]
			break
		}

		//fmt.Println("allowed map", allowedIndicesMap[i], "shuffle index", shuffleIndex, "index", i)

		if g.currentStatus[shuffleIndex] == 2 {
			continue
		}

		//remove shuffle index from subsequent allowed indices map
		for j := i + 1; j < len(shuffle); j++ {
			for k, v := range allowedIndicesMap[j] {
				if v == shuffleIndex {
					allowedIndicesMap[j] = append(allowedIndicesMap[j][:k], allowedIndicesMap[j][k+1:]...)
					break
				}
			}

			if len(allowedIndicesMap[j]) == 0 {
				allowedIndicesMap[j] = append(allowedIndicesMap[j], j)
			}
		}
	}

	// take permutations and select longest trie match on eligible permutation
	permutations := Permutations(string(shuffleBuffer), func(word string) bool {
		for i, b := range word {
			if byte(b) == g.shuffle[i] && g.currentStatus[i] < 2 {
				return false
			}

			if byte(b) != g.shuffle[i] && g.currentStatus[i] == 2 {
				return false
			}
		}

		return true
	})

	longestMatch := 1
	matchedBuffer := ""

	for _, p := range permutations {
		match := g.handle.Match(p)

		if match > longestMatch {
			longestMatch = match
			matchedBuffer = p
		}
	}

	if matchedBuffer != "" {
		shuffleBuffer = []byte(matchedBuffer)
	}

	g.shuffle = shuffleBuffer
	g.mark(g.shuffle)

	return string(g.shuffle)
}

func (g *guessWork) mark(buffer []byte) {
	for _, b := range buffer {
		g.usedMap[int(b)-97] = 1
	}
}
