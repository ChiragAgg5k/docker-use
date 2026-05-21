docker-use() {
  if [ "$1" = "use" ]; then
    shift
    eval "$("{{ .Binary }}" use "$@")"
  else
    "{{ .Binary }}" "$@"
  fi
}
