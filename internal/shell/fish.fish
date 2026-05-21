function docker-use
  if test "$argv[1]" = use
    set -e argv[1]
    set -l path ("{{ .Binary }}" use $argv)
    or return
    set -gx DOCKER_CONFIG "$path"
  else
    "{{ .Binary }}" $argv
  end
end
