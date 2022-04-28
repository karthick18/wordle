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
	mostUsed = []byte{'a', 'e', 't', 's', 'd', 'm', 'p', 'c', 'i', 'o'}
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func newGuess(handle *HandleImplementor, wordLen int) *guessWork {
	shuffle := make([]byte, len(mostUsed))
	copy(shuffle, mostUsed)

	usedMap := make([]int, 26)

	for {
		rand.Shuffle(len(shuffle), func(i, j int) {
			shuffle[i], shuffle[j] = shuffle[j], shuffle[i]
		})

		res := handle.Wordle(string(shuffle[:wordLen]), nil, 10)
		if len(res) == 0 {
			//			fmt.Println("res 0 for", string(shuffle[:wordLen]))
			continue
		}

		shuffle = []byte(res[rand.Intn(len(res))])
		break
	}

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
		// has a byte that has been marked absent
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
	reusableSlots := []int{}
	currentEmptySlots := []int{}
	countMap := make(map[byte]int)
	reusableSlotMap := make(map[byte][]int)

	for i := 0; i < g.wordLen; i++ {
		v := g.currentStatus[i]
		switch v {
		case 0:
			// mark entry as removed from eligible slot
			g.usedMap[int(g.shuffle[i])-97] = 3
			emptySlots = append(emptySlots, i)
		case 1:
			shuffle[i] = g.shuffle[i]
			if g.usedMap[int(g.shuffle[i])-97] != 2 {
				g.usedMap[int(g.shuffle[i])-97] = 1
			}

			if countMap[g.shuffle[i]] > 1 {
				// mark it as a duplicate that can be reused
				if len(reusableSlotMap[g.shuffle[i]]) > 0 {
					reusableSlots = append(reusableSlots, reusableSlotMap[g.shuffle[i]]...)
					reusableSlotMap[g.shuffle[i]] = []int{}
				}

				reusableSlots = append(reusableSlots, i)
			} else {
				// mark it as a candidate for reuse
				reusableSlotMap[g.shuffle[i]] = append(reusableSlotMap[g.shuffle[i]], i)
			}

			countMap[g.shuffle[i]] += 1
		case 2:
			countMap[g.shuffle[i]] += 1
			shuffle[i] = g.shuffle[i]
			g.usedMap[int(g.shuffle[i])-97] = 2

			if len(reusableSlotMap[g.shuffle[i]]) > 0 {
				reusableSlots = append(reusableSlots, reusableSlotMap[g.shuffle[i]]...)
				reusableSlotMap[g.shuffle[i]] = []int{}
			}
		}

	}

	if string(shuffle) == string(g.shuffle) {
		return string(shuffle)
	}

	currentEmptySlots = append(currentEmptySlots, emptySlots...)

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

	for {
		frequentlyUsed := false
		countMap := make(map[byte]int)

		for _, b := range shuffle {
			if int(b) == 0 {
				continue
			}

			countMap[b] += 1
			if countMap[b] >= 2 {
				frequentlyUsed = true
			}
		}

		for _, slot := range emptySlots {
			for {
				eligibleIndex := rand.Intn(len(g.usedMap))
				// skip if marked absent
				if g.usedMap[eligibleIndex] == 3 {
					continue
				}

				b := byte(int('a') + eligibleIndex)

				if countMap[b] >= 2 {
					continue
				}

				if frequentlyUsed && countMap[b]+1 >= 2 {
					continue
				}

				countMap[b] += 1
				if countMap[b] > 1 {
					frequentlyUsed = true
				}

				shuffle[slot] = b
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

					// check if this is a duplicate byte
					// in which case it cannot be assigned to another duplicate byte/pinned location
					if v == 1 && byte(shuffle[i]) == byte(shuffle[j]) {
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

		for i := range shuffle {
			allowedIndex := rand.Intn(len(allowedIndicesMap[i]))
			shuffleIndex := allowedIndicesMap[i][allowedIndex]
			shuffleBuffer[i] = shuffle[shuffleIndex]

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

		// take permutations and get eligible words
		matches := g.handle.Wordle(string(shuffleBuffer), func(word string) bool {
			for i, b := range word {
				if byte(b) == g.shuffle[i] && g.currentStatus[i] < 2 {
					return false
				}

				if byte(b) != g.shuffle[i] && g.currentStatus[i] == 2 {
					return false
				}
			}

			return true
		}, 5)

		// if no valid permutations are found, try swapping out reusable slots to see if we can find one
		if len(matches) == 0 {

			if len(reusableSlots) > 0 {
				reusableSlot := reusableSlots[0]
				reusableSlots = reusableSlots[1:]
				currentEmptySlots = append(currentEmptySlots, reusableSlot)
				emptySlots = []int{reusableSlot}

				//retry again
				continue
			}

			emptySlots = currentEmptySlots

			if len(emptySlots) > 0 {
				continue
			}

			// exhausted retries to find a word
			break
		}

		matchedBuffer := matches[rand.Intn(len(matches))]

		shuffleBuffer = []byte(matchedBuffer)

		g.shuffle = shuffleBuffer
		g.mark(g.shuffle)

		return string(g.shuffle)
	}

	return ""
}

func (g *guessWork) mark(buffer []byte) {
	for _, b := range buffer {
		g.usedMap[int(b)-97] = 1
	}
}