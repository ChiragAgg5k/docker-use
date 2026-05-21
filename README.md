# docker-use

Manage multiple Docker Hub accounts from the command line.

## Install

```sh
brew tap chiragagg5k/tools
brew install docker-use
```

Or download a binary from the [releases page](https://github.com/chiragagg5k/docker-use/releases).

## Setup

Add the following to your shell rc file (e.g., `~/.zshrc`):

```sh
eval "$(docker-use init zsh)"
```

Supported shells: `zsh`, `bash`, `fish`.

## Usage

```sh
docker-use list                  # list accounts
docker-use whoami                # show current account
docker-use add <name> -u <user>  # add a new account (interactive docker login)
docker-use remove <name>         # remove an account (with confirmation)
docker-use use <name>            # switch to an account (requires shell wrapper)
```

## How it works

Each account lives in `~/.docker-accounts/<name>/config.json`. When you `use` an account, the shell wrapper sets `DOCKER_CONFIG` to that directory, so `docker` commands use the right credentials.

During `add`, `docker-use` shells out to `docker login`, then strips `credsStore` and `credHelpers` from the generated config so credentials stay in the isolated config file.
