# flow

- Simple personal CLI for managing `tmux` sessions
- Currently requires `tmux` and `fzf`

## Installation

```sh
go install github.com/winter-again/flow@latest
```

## Configuration

Default config file location is `$HOME/.config/flow/config.toml`. Custom path can be specified instead:

```sh
flow --config path/to/config/file
```

Config file looks like this:

```toml
[flow]
init_session_name = "0" # default

[fzf-tmux]
width = "80%" # default
length = "60%" # default
border = "rounded" # default

preview_pos = "right" # default
preview_dir_cmd = ["eza", "-lah", "--icons", "--color", "always", "--group-directories-first"] # default: ["ls"]

[find]
dirs = ["~/Documents/code"] # default is []
```
