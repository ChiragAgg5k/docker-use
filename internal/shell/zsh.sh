docker-use() {
  if [ "$1" = "add" ] || [ "$1" = "remove" ] || [ "$1" = "rm" ] || [ "$1" = "list" ] || [ "$1" = "ls" ] || [ "$1" = "whoami" ] || [ "$1" = "init" ] || [ "$1" = "completion" ] || [ "$1" = "help" ] || [ "$1" = "--help" ] || [ "$1" = "-h" ] || [ $# -eq 0 ]; then
    "{{ .Binary }}" "$@"
  else
    docker_use_config="$("{{ .Binary }}" __path "$@")" || return
    export DOCKER_CONFIG="$docker_use_config"
    printf 'Switched Docker account to %s\n' "$1"
  fi
}
