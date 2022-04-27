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
	wordLen := 5
	maxAttempts := 6
	input := "cause"

	flag.StringVar(&dict, "dict", dict, "Words dictionary file separated by newlines")
	flag.StringVar(&input, "input", input, "Word to guess")
	flag.IntVar(&maxAttempts, "max", maxAttempts, "Max attempts to guess the word")
	flag.IntVar(&wordLen, "len", wordLen, "Wordlen to solve")
	flag.Parse()

	if len(input) != wordLen {
		fmt.Fprintf(os.Stderr, "Wordlen %d and input %s length should be same\n", wordLen, input)
		os.Exit(1)
	}

	seenMap := make(map[rune]bool)
	for _, b := range input {
		seenMap[b] = true
	}

	if len(seenMap) != len(input) {
		fmt.Fprintf(os.Stderr, "Input %s cannot have letters repeated\n", input)
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

	for prev, guess = "", wh.Start(); attempts < maxAttempts && guess != "" && prev != guess; prev, guess = guess, wh.Next(status) {
		attempts++
		fmt.Println("Guess", guess)
		fmt.Println("Enter status (0 - mismatch, 1 - position mismatch, 2 - matched). Attempt", attempts)
		fmt.Scanf("%d %d %d %d %d", &status[0], &status[1], &status[2], &status[3], &status[4])
		fmt.Println("status", status)
	}

	if guess != input {
		fmt.Fprintf(os.Stderr, "Exhausted guesses for wordle %s\n", input)
		os.Exit(1)
	}

	fmt.Println("Worldle word", input, "guessed in", attempts, "attempts")
}
