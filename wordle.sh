#!/usr/bin/env bash
echo -e "Ctrl^C to quit\n"

while read input; do
      ./bin/wordle -wordle $input
done
