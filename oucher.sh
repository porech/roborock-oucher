#!/bin/bash

# Set the expressions language to have a better pronunce
language="en"

# List the expressions to say
expressions=(
'Ouch!'
'Argh!'
'Ehy, it hurts!'
)

# Don't touch below if you don't know what you're doing ;)
while true; do
	tail -n 0 -F /run/shm/NAV_normal.log 2>/dev/null | grep --line-buffered ':Bumper' | grep -v --line-buffered 'Curr:(0, 0, 0)' | (
		read line;
		while [ "$line" != "" ]; do
			echo "$line";
			selectedexpression=${expressions[$RANDOM % ${#expressions[@]} ]}
			espeak -v "$language" "$selectedexpression" --stdout | aplay 2> /dev/null > /dev/null;
			exit 0;
		done
	);
	sleep 1;
done
