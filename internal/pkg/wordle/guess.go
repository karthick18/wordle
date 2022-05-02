package wordle

import (
	"errors"
	"fmt"
	"math/rand"
	"os"
	"time"
)

type guessWork struct {
	handle        *HandleImplementor
	wordLen       int
	shuffle       []byte
	usedMap       map[byte]State
	currentStatus []State
	cache         map[byte][]int
}

type State int

const (
	alphabets    = 26
	alphabetBase = int('a')
)

const (
	Present State = iota + 1
	Locked
	Deleted
)

var (
	mostUsedLetters  = []byte{'a', 'e', 't', 's', 'd', 'n', 'p', 'h', 'i', 'o', 'r', 'w'}
	ErrInvalidStatus = errors.New("invalid status value. allowed (0/1/2)")
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func newGuess(handle *HandleImplementor, wordLen int) *guessWork {
	shuffle := make([]byte, len(mostUsedLetters))
	copy(shuffle, mostUsedLetters)

	usedMap := make(map[byte]State, alphabets)

	for {
		rand.Shuffle(len(shuffle), func(i, j int) {
			shuffle[i], shuffle[j] = shuffle[j], shuffle[i]
		})

		res := handle.Wordle(string(shuffle[:wordLen]), nil, 10)
		if len(res) == 0 {
			continue
		}

		shuffle = []byte(res[rand.Intn(len(res))])
		break
	}

	cache := make(map[byte][]int)

	// note down the positions
	for i, b := range shuffle {
		cache[b] = append(cache[b], i)
	}

	g := &guessWork{
		handle:  handle,
		wordLen: wordLen,
		shuffle: shuffle,
		usedMap: usedMap,
		cache:   cache,
	}

	g.mark(shuffle)

	return g
}

func ToState(status []int) ([]State, error) {
	state := make([]State, len(status))

	for i, v := range status {
		switch v {
		case 0:
			state[i] = Deleted
		case 1:
			state[i] = Present
		case 2:
			state[i] = Locked
		default:
			return nil, ErrInvalidStatus
		}
	}

	return state, nil
}

func (g *guessWork) get() string {
	return string(g.shuffle)
}

func (g *guessWork) accept(word string, curCountMap map[byte]int) bool {
	for i, v := range word {
		if byte(v) == g.shuffle[i] {
			if g.currentStatus[i] == Locked {
				continue
			}

			return false
		}

		if byte(v) != g.shuffle[i] && g.currentStatus[i] == Locked {
			return false
		}

		if g.usedMap[byte(v)] == Deleted {
			if curCountMap[byte(v)] == 0 {
				return false
			}

			count := 0
			for _, b := range word {
				if byte(b) == byte(v) {
					count++
				}
			}

			if count != curCountMap[byte(v)] {
				return false
			}
		}

		// if found in cache, return
		if g.findCache(byte(v), i) {
			return false
		}

		// if there is a position mismatch, this has to be found in the remaining
		if g.currentStatus[i] == Present {
			matched := false

			for j, b := range word {
				if j == i {
					continue
				}

				if byte(b) == g.shuffle[i] {
					if !g.findCache(byte(b), j) {
						matched = true
					} else {
						matched = false
						break
					}
				}
			}

			if !matched {
				return false
			}
		}
	}

	return true

}

func (g *guessWork) addCache(guess []byte) {
	for i, b := range g.shuffle {
		if g.currentStatus[i] != Present {
			continue
		}

		// this has to be in the new guess. note down the slot
		for j, b2 := range guess {
			if i == j {
				continue
			}

			if b2 == b {
				g.cache[b] = append(g.cache[b], j)
			}
		}
	}
}

func (g *guessWork) findCache(b byte, index int) bool {
	for _, v := range g.cache[b] {
		if v == index {
			return true
		}
	}

	return false
}

func (g *guessWork) next(status []State) string {
	g.currentStatus = status
	shuffle := make([]byte, g.wordLen)
	emptySlots := []int{}
	reusableSlots := []int{}
	currentEmptySlots := []int{}
	countMap := make(map[byte]int)
	reusableSlotMap := make(map[byte][]int)
	guesses := make(map[string]int)
	maxFailures := 10
	matches := 0

	for i := 0; i < g.wordLen; i++ {
		v := g.currentStatus[i]
		switch v {
		case Deleted:
			// mark entry as removed from eligible slot
			g.usedMap[g.shuffle[i]] = Deleted
			emptySlots = append(emptySlots, i)
		case Present:
			shuffle[i] = g.shuffle[i]
			if g.usedMap[g.shuffle[i]] != Deleted {
				g.usedMap[g.shuffle[i]] = Present
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
		case Locked:
			matches++
			countMap[g.shuffle[i]] += 1
			shuffle[i] = g.shuffle[i]
			if g.usedMap[g.shuffle[i]] != Deleted {
				g.usedMap[g.shuffle[i]] = Locked
			}

			if len(reusableSlotMap[g.shuffle[i]]) > 0 {
				reusableSlots = append(reusableSlots, reusableSlotMap[g.shuffle[i]]...)
				reusableSlotMap[g.shuffle[i]] = []int{}
			}
		}

	}

	if matches >= g.wordLen {
		return string(shuffle)
	}

	for _, slot := range emptySlots {
		if countMap[g.shuffle[slot]] == 0 {
			delete(g.cache, g.shuffle[slot])
		}
	}

	currentEmptySlots = append(currentEmptySlots, emptySlots...)

	// check for consecutive matches and try autocomplete with filter
	// check for the first consecutive match start and end location
	startLocation, endLocation := -1, -1

	for i := 0; i < len(g.shuffle)-1; i++ {
		if g.currentStatus[i] == Locked && g.currentStatus[i+1] == Locked {
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
				return g.accept(word, countMap)
			})

		switch {
		case len(completions) > 1:
			res := rand.Intn(len(completions))

			return g.update([]byte(completions[res]))

		case len(completions) == 1:
			return g.update([]byte(completions[0]))

		default:
			break
		}
	}

	for {
		frequentlyUsed := false
		countMap := make(map[byte]int)

		// move the reusable slot bytes to empty slots if available
		// and make reusable slot empty
		for _, reusableSlot := range reusableSlots {
			if len(emptySlots) > 0 {
				emptySlot := emptySlots[0]
				emptySlots = emptySlots[1:]
				shuffle[emptySlot] = g.shuffle[reusableSlot]
				emptySlots = append(emptySlots, reusableSlot)
			}
		}

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
				eligibleIndex := rand.Intn(alphabets)
				b := byte(alphabetBase + eligibleIndex)

				// skip if marked deleted
				if g.usedMap[b] == Deleted {
					continue
				}

				if countMap[b] >= 2 {
					continue
				}

				// already 2 same chars exist
				if frequentlyUsed && countMap[b]+1 >= 2 {
					continue
				}

				// already has a position mismatch
				if g.currentStatus[slot] == Present && b == g.shuffle[slot] {
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
			if v == Locked {
				allowedIndicesMap[i] = append(allowedIndicesMap[i], i)
			} else {
				for j := range shuffle {
					if i == j {
						continue
					}

					if g.currentStatus[j] == Locked {
						continue
					}

					// check if this is a duplicate byte
					// in which case it cannot be assigned to another duplicate byte/pinned location
					if v == Present && byte(shuffle[i]) == byte(shuffle[j]) {
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

			if g.currentStatus[shuffleIndex] == Locked {
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
			return g.accept(word, countMap)
		}, 5)

		// if no valid permutations are found, retry if applicable
		if len(matches) == 0 {
			if len(emptySlots) > 0 {
				if guesses[string(shuffleBuffer)] >= maxFailures {
					fmt.Fprintln(os.Stderr, "Exhausted shuffle buffer retries")
					break
				}

				guesses[string(shuffleBuffer)]++
				emptySlots = append([]int{}, currentEmptySlots...)
				continue
			}

			// exhausted retries to find a word
			fmt.Fprintln(os.Stderr, "No empty slots to retry")
			break
		}

		matchedBuffer := matches[rand.Intn(len(matches))]

		shuffleBuffer = []byte(matchedBuffer)
		return g.update(shuffleBuffer)
	}

	return ""
}

func (g *guessWork) mark(buffer []byte) {
	for _, b := range buffer {
		if g.usedMap[b] != Deleted {
			g.usedMap[b] = Present
		}
	}
}

// update adds the letter positions to the cache
// and also updates the shuffle buffer
func (g *guessWork) update(buffer []byte) string {
	g.mark(buffer)
	g.addCache(buffer)
	g.shuffle = buffer

	return string(g.shuffle)
}
