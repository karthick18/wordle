package wordle

type trie struct {
	root *node
}

type node struct {
	children map[byte]*node
	word     bool
}

type Handle interface {
	Wordle(string, func(string) bool, ...int) []string
	Start() string
	Next(state []State) string
}

type HandleImplementor struct {
	*trie
	guessRef *guessWork
}

var _ Handle = &HandleImplementor{}

func New(words []string, wordLen int) *HandleImplementor {
	handle := newHandle(wordLen)

	for _, word := range words {
		handle.Insert(word)
	}

	handle.guessRef = newGuess(handle, wordLen)

	return handle
}

func newHandle(wordLen int) *HandleImplementor {
	return &HandleImplementor{
		trie: newTrie(),
	}
}

func newTrie() *trie {
	return &trie{root: newNode()}
}

func newNode() *node {
	return &node{children: make(map[byte]*node)}
}

func (t *trie) Insert(word string) {
	nodeRef := t.root

	for i := 0; i < len(word); i++ {
		if nodeRef.children[word[i]] == nil {
			nodeRef.children[word[i]] = newNode()
		}

		nodeRef = nodeRef.children[word[i]]
	}

	nodeRef.word = true
}

func (t *trie) Lookup(word string) bool {
	nodeRef := t.root

	for i := 0; i < len(word); i++ {
		nodeRef = nodeRef.children[word[i]]
		if nodeRef == nil {
			return false
		}
	}

	return nodeRef.word
}

func (t *trie) AutoComplete(prefix string,
	wordLen int, filter func(string) bool) []string {
	var res []string
	nodeRef := t.root

	for i := 0; i < len(prefix); i++ {
		if nodeRef.children[prefix[i]] == nil {
			return res
		}

		nodeRef = nodeRef.children[prefix[i]]
	}

	t.complete(nodeRef, wordLen, filter, prefix, &res)

	return res
}

func (t *trie) complete(nodeRef *node, wordLen int,
	filter func(string) bool, prefix string, res *[]string) {
	if len(prefix) == wordLen {
		if nodeRef.word {
			if filter != nil {
				if filter(prefix) {
					*res = append(*res, prefix)
				}
			} else {
				*res = append(*res, prefix)
			}
		}

		return
	}

	for b, child := range nodeRef.children {
		newPrefix := prefix + string(b)

		t.complete(child, wordLen, filter, newPrefix, res)
	}
}

func (t *trie) AutoCompleteSubstring(substring string, startLocation, wordLen int,
	filter func(string) bool) []string {
	nodeRef := t.root
	level := 0

	if startLocation > 0 {
		res := []string{}

		for b, child := range nodeRef.children {
			prefix := string(b)
			output := t.completeSubstring(child, prefix, substring, level+1, startLocation, wordLen, filter)

			res = append(res, output...)
		}

		return res
	} else {
		return t.AutoComplete(substring, wordLen, filter)
	}
}

func (t *trie) completeSubstring(nodeRef *node, prefix, substring string, level, startLocation, wordLen int,
	filter func(string) bool) []string {
	if level == startLocation {
		for i := 0; i < len(substring); i++ {
			if nodeRef.children[substring[i]] == nil {
				return nil
			}

			nodeRef = nodeRef.children[substring[i]]
		}

		//we can autocomplete now
		var res []string
		t.complete(nodeRef, wordLen, filter, prefix+substring, &res)

		return res
	}

	if level >= wordLen {
		return nil
	}

	res := []string{}

	for b, child := range nodeRef.children {
		newPrefix := prefix + string(b)
		output := t.completeSubstring(child, newPrefix, substring, level+1, startLocation, wordLen, filter)
		res = append(res, output...)
	}

	return res
}

func (t *trie) Match(prefix string) int {
	nodeRef := t.root

	for i := 0; i < len(prefix); i++ {
		if nodeRef.children[prefix[i]] == nil {
			return i
		}

		nodeRef = nodeRef.children[prefix[i]]
	}

	return len(prefix)
}

// Wordle input is a jumbled string.
// Make permutations and check if any matches a word in a dictionary.
func (h *HandleImplementor) Wordle(wordle string, filter func(string) bool, maxResults ...int) []string {
	permutations := Permutations(wordle, nil)

	max := 1

	if len(maxResults) > 0 {
		max = maxResults[0]
	}

	if max < 1 {
		max = 1
	}

	res := make([]string, 0, max)

	for _, permutation := range permutations {
		if h.Lookup(permutation) {
			if filter != nil {
				if filter(permutation) {
					res = append(res, permutation)
				}
			} else {
				res = append(res, permutation)
			}

			if len(res) >= max {
				break
			}
		}
	}

	return res
}

func (h *HandleImplementor) Start() string {
	return h.guessRef.get()
}

func (h *HandleImplementor) Next(state []State) string {
	return h.guessRef.next(state)
}
