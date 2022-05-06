package main

import (
	"bufio"
	"flag"
	"fmt"
	"github.com/karthick18/wordle/internal/pkg/wordle"
	"os"
	"strings"
)

func main() {
	dict := "words_alpha.txt"
	wordLen := 6
	maxAttempts := 12

	flag.StringVar(&dict, "dict", dict, "Words dictionary file separated by newlines")
	flag.IntVar(&maxAttempts, "max", maxAttempts, "Max attempts to guess the word")
	flag.IntVar(&wordLen, "len", wordLen, "Wordlen to solve")
	flag.Parse()

	if wordLen != 5 && wordLen != 6 {
		fmt.Fprintln(os.Stderr, "Supported wordlens (5, 6)")
		os.Exit(1)
	}

	f, err := os.Open(dict)
	if err != nil {
		panic(err.Error())
	}

	defer f.Close()

	scanner := bufio.NewScanner(f)

	words := []string{}

	for scanner.Scan() {
		word := strings.TrimSpace(scanner.Text())
		if len(word) != wordLen {
			continue
		}

		words = append(words, strings.TrimSpace(scanner.Text()))
	}

	if err = scanner.Err(); err != nil {
		panic(err.Error())
	}

	wh := wordle.New(words, wordLen)

	var prev, guess string

	attempts := 0
	status := make([]int, wordLen)
	state := make([]wordle.State, len(status))

	for prev, guess = "", wh.Start(); attempts < maxAttempts && guess != "" && prev != guess; prev, guess = guess, wh.Next(state) {
		attempts++
		fmt.Println("Guess", guess)
		fmt.Println("Enter status (0 - mismatch, 1 - position mismatch, 2 - matched). Attempt", attempts)
		switch {
		case wordLen == 5:
			fmt.Scanf("%d %d %d %d %d", &status[0], &status[1], &status[2], &status[3], &status[4])
		case wordLen == 6:
			fmt.Scanf("%d %d %d %d %d %d", &status[0], &status[1], &status[2], &status[3], &status[4], &status[5])
		default:
			fmt.Fprintf(os.Stderr, "Unsupported wordlen: %d\n", wordLen)
			os.Exit(1)
		}

		if state, err = wordle.ToState(status); err != nil {
			panic(err.Error())
		}
	}

	if guess == "" {
		fmt.Fprintf(os.Stderr, "Exhausted guesses for wordle. Attempts %d\n", attempts)
		os.Exit(1)
	}

	if guess == prev {
		fmt.Println("Wordle word", guess, "guessed in", attempts, "attempts")
	} else {
		fmt.Fprintf(os.Stderr, "Exhausted guesses for wordle. Attempts %d\n", attempts)
	}
}
