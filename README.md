## Wordle solver

**Download and install [Go](https://go.dev/doc/install)**

*Build and run program using command line*.  
```
make 
bin/wordle
```

It uses words.txt as the dictionary to solve 5 letter wordles.

You basically enter the guesses the program outputs while playing [Wordle](https://www.nytimes.com/games/wordle/index.html)

When it asks for feedback about the guess, enter **0 for grey, 1 for yellow, 2 for green**.  
For example, if the website shows *grey, yellow, yellow, grey, green*     
you will type: **0 1 1 0 2** to see the next guess. 

The guesses are optimized by keeping track of the positions of last guesses in memory.  
This ensures when bytes are shuffled randomly, the same positions are not re-used.

The output of the permutations of the shuffled bytes are filtered based on the 
current status and last positions.  
The matched word for the permutations are the best guesses.

If there are 2 or more consecutive hits **(green)**, the output is autocompleted and filtered to find the best word to guess.

*Happy cheating*,

-Karthick

