#!/bin/bash

WATCH_DIR="./"
DEBOUNCE_TIME=1

run_command() {
    echo "Detected change in .go file. Running command..."

    # Kill the previous command if it's still running
    if [[ $PID ]]; then
        echo "Killing previous command with PID: $PID"
        kill "$PID"
        wait "$PID" 2>/dev/null
    fi

    ./genstatic.sh
    ./gentemplate.sh
    go build -o bin/go4ignition .

    if [ $? -eq 0 ]; then
      bin/go4ignition &
      PID=$!
      echo "New command PID: $PID"
    else
      PID=-1
    fi
}

run_command

inotifywait -m -r -e modify --format '%w%f' "$WATCH_DIR" | while read FILE
do
    if [[ ! $last_run_time || $(($(date +%s) - last_run_time)) -ge $DEBOUNCE_TIME ]]; then
        run_command
        last_run_time=$(date +%s)
    fi
done
