# wordle
Wordle solver

It uses words.txt as the dictionary to solve 5 letter wordles.
You basically enter the guesses you see while running the program at *Wordle NewYorkTimes* website.
0 for grey, 1 for yellow, 2 for green.

The guesses are optimized by keeping track of the positions of last guesses in memory.
This ensures when bytes are shuffled, the same positions are not re-used.

The output of the permutations of the shuffled bytes are filtered based on the current status and last positions.
The matched word for the permutations are the best guesses.

If there are 2 or more consecutive hits (green), the output is autocompleted and filtered to find the best word to guess.


