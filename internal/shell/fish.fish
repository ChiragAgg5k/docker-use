function docker-use
  set -l docker_use_config ("{{ .Binary }}" __switch -- $argv 2>/dev/null)
  if test $status -eq 0
    set -gx DOCKER_CONFIG "$docker_use_config"
    printf 'Switched Docker account to %s\n' "$argv[1]"
  else
    "{{ .Binary }}" $argv
  end
end
set -l docker_use_config ("{{ .Binary }}" __current)
if test -n "$docker_use_config"
  set -gx DOCKER_CONFIG "$docker_use_config"
end
