#!/bin/bash
PYTHON_SCRIPT_PATH=$1

while true; do
  python2 "$PYTHON_SCRIPT_PATH"
  EXIT_CODE=$?
  if [ $EXIT_CODE -ne 0 ]; then
    echo "Script crashed with exit code $EXIT_CODE. Restarting..." >&2
    sleep 1
  fi
done

