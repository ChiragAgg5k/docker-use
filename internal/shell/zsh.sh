docker-use() {
  local docker_use_config
  if [ $# -eq 0 ]; then
    "{{ .Binary }}"
    return $?
  fi
  case "$1" in
    add|completion|help|init|list|ls|remove|rm|whoami|--help|-h|--version)
      "{{ .Binary }}" "$@"
      return $?
      ;;
  esac
  docker_use_config="$("{{ .Binary }}" __switch -- "$@" 2>/dev/null)"
  if [ $? -eq 0 ]; then
    export DOCKER_CONFIG="$docker_use_config"
    printf 'Switched Docker account to %s\n' "$1"
  else
    "{{ .Binary }}" "$@"
  fi
}
if [ -z "${DOCKER_CONFIG:-}" ]; then
  docker_use_config="$("{{ .Binary }}" __current 2>/dev/null)" && [ -n "$docker_use_config" ] && export DOCKER_CONFIG="$docker_use_config"
fi
unset docker_use_config
