<p align="center">
  <img src="assets/banner.png" alt="docker-use banner">
</p>

# docker-use

Manage multiple Docker Hub accounts from the command line.

## Why

Docker only reads credentials from the active `DOCKER_CONFIG`, which makes switching between personal, work, and automation accounts awkward. You either keep logging in and out, manually juggle config directories, or risk pushing and pulling with the wrong account.

`docker-use` keeps each account in its own isolated Docker config and lets your shell switch between them by name.

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

The generated wrapper calls `docker-use <name>` and assigns the returned path to `DOCKER_CONFIG` without shell `eval` during account switching.

Supported shells: `zsh`, `bash`, `fish`.

## Usage

```sh
docker-use list                  # list accounts
docker-use whoami                # show current account
docker-use add <name> -u <user>  # add a new account (interactive docker login)
docker-use add <name> -u <user> --force  # replace an existing account
docker-use remove <name>         # remove an account (with confirmation)
docker-use <name>                # switch to an account (requires shell wrapper)
```

## How it works

Each account lives in `~/.docker-accounts/<name>/config.json`. Account names must match `^[a-zA-Z0-9][a-zA-Z0-9._-]{0,63}$`. When you `use` an account, the shell wrapper sets `DOCKER_CONFIG` to that directory, so `docker` commands use the right credentials.

During `add`, `docker-use` shells out to `docker login`, then strips `credsStore` and `credHelpers` from the generated config so credentials stay in the isolated config file.
