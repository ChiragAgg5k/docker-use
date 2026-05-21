docker-use() {
  docker_use_config="$("{{ .Binary }}" __switch -- "$@" 2>/dev/null)"
  if [ $? -eq 0 ]; then
    export DOCKER_CONFIG="$docker_use_config"
    printf 'Switched Docker account to %s\n' "$1"
  else
    "{{ .Binary }}" "$@"
  fi
}
docker_use_config="$("{{ .Binary }}" __current)" && [ -n "$docker_use_config" ] && export DOCKER_CONFIG="$docker_use_config"
unset docker_use_config
