docker-use() {
  if [ "$1" = "use" ]; then
    shift
    path="$("{{ .Binary }}" use "$@")" || return
    export DOCKER_CONFIG="$path"
  else
    "{{ .Binary }}" "$@"
  fi
}
