function docker-use
  if contains -- "$argv[1]" add remove rm list ls whoami init completion help --help -h; or test (count $argv) -eq 0
    "{{ .Binary }}" $argv
  else
    set -l docker_use_config ("{{ .Binary }}" __path $argv)
    or return
    set -gx DOCKER_CONFIG "$docker_use_config"
    printf 'Switched Docker account to %s\n' "$argv[1]"
  end
end
