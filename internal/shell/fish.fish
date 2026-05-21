function docker-use
  if test (count $argv) -eq 0
    "{{ .Binary }}"
    return $status
  end
  switch $argv[1]
    case add completion help init list ls remove rm whoami --help -h --version
      "{{ .Binary }}" $argv
      return $status
  end
  set -l docker_use_config ("{{ .Binary }}" __switch -- $argv 2>/dev/null)
  if test $status -eq 0
    set -gx DOCKER_CONFIG "$docker_use_config"
    printf 'Switched Docker account to %s\n' "$argv[1]"
  else
    "{{ .Binary }}" $argv
  end
end
if not set -q DOCKER_CONFIG
  set -l docker_use_config ("{{ .Binary }}" __current 2>/dev/null)
  if test -n "$docker_use_config"
    set -gx DOCKER_CONFIG "$docker_use_config"
  end
  set --erase docker_use_config
end
