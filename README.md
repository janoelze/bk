# bk

Directory bookmarks for your terminal.

## Install

```
./install.sh
```

Add the shell function to your `.zshrc` or `.bashrc`:

```sh
bk() {
  if [[ "$1" == "add" ]]; then
    command bk add
  else
    local dir
    dir=$(command bk)
    if [[ -n "$dir" && -d "$dir" ]]; then
      cd "$dir"
    fi
  fi
}
```

## Usage

```
bk add    # bookmark current directory
bk        # open selector
```

## Keys

```
↑/↓       navigate
enter     go to directory
e         rename bookmark
d         delete bookmark
q/esc     quit
<text>    filter list
```

## Storage

Bookmarks are stored in `~/.config/bk/bookmarks.json`.

Frequently used bookmarks rise to the top.
