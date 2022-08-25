function waitfor() {
  for i in {1..30}; do
    if [ ! -z "$(${@})" ]; then
      break
    fi
    sleep 1
  done
  if [ -z "$(${@})" ]; then
    echo "No results for '${1}' after 30 attempts"
  fi
}