#!/bin/bash

# This command simplifies development by automatically killing, rebuilding, and restarting the server process when
# files are modified

# Directory to watch
WATCH_DIR="./"

# Debounce time in seconds
DEBOUNCE_TIME=1

# Function to run the command
run_command() {
    FILE=$1
    echo "Modification detected in file $(echo $FILE | sed 's|~$||')"
    # Kill the previous command if it's still running
    if [[ $PID ]]; then
        echo "Killing previous command with PID: $PID"
        kill "$PID"
        wait "$PID" 2>/dev/null
    fi
    # Run the command in the background and capture its PID
    ./genstatic.sh $FILE
    ./gentemplate.sh $FILE

    # Regenerate readme.html if pandoc is installed
    if [ -d sites/.skeleton ];then
      if command -v pandoc &> /dev/null;then
        echo "Regenerating readme.html..."
        README_DIR="sites/.skeleton/template"
        README_TMP="${README_DIR}/readmetmp.html"
        README_HTML="${README_DIR}/readme.html"
        echo '{{ template "template/fragment/header.html" . }}' > "${README_HTML}"
        pandoc README.md -o "${README_TMP}"
        cat "${README_TMP}" >> "${README_HTML}"
        rm -f "${README_TMP}"
        echo '{{ template "template/fragment/footer.html" . }}' >> "${README_HTML}"
      else
          echo "WARNING: readme.html cannot be regenerated because pandoc is not installed"
      fi
    fi

    go build -o bin/go4ignition *.go

    # Regenerate persistence code if the migrations have changed
    if [[ $FILE == *"migrations.go" ]];then
      # Create an empty database with an up to date schema
      TRANSIENT_DB=$(mktemp)
      bin/go4ignition --run-migrations "${FILE}" "${TRANSIENT_DB}"

      # Introspect the empty database to generate persistence code
      ./genpersistence.sh "${FILE}" "${TRANSIENT_DB}"

      # Compile the new persistence code into the application
      go build -v -o bin/go4ignition *.go

      # Delete the empty database
      rm -f "${TRANSIENT_DB}"
    fi

    # Don't start the old version of the server because that may obscure build errors.
    if [ $? -eq 0 ]; then
      bin/go4ignition &
      PID=$!
      echo "New command PID: $PID"
    else
      PID=-1
    fi
}

run_command

# Use inotifywait to monitor the directory for changes
inotifywait -m -r -e modify --format '%w%f' "$WATCH_DIR" | while read FILE
do
    # Only run the command after the debounce time
    if [[ ! $last_run_time || $(($(date +%s) - last_run_time)) -ge $DEBOUNCE_TIME ]]; then
        run_command $FILE
        last_run_time=$(date +%s)
    fi
done
