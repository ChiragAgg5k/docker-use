function docker-use
  if test "$argv[1]" = use
    set -e argv[1]
    eval ("{{ .Binary }}" use $argv)
  else
    "{{ .Binary }}" $argv
  end
end
