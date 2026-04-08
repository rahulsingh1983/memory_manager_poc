# memory_manager_poc

Library-first virtual memory manager over a byte-array-backed simulated disk.

The manager supports:
- Alloc(size)
- Free(handle)
- Read(handle, off, n)
- Write(handle, off, data)

One logical allocation may map to multiple physical extents.

## Prerequisites

This project requires Go **1.25.0**, managed via [goenv](https://github.com/go-env/goenv).

The setup steps below are for **macOS** and assume you are using **zsh**.

### 1. Install Homebrew

The goenv installation command below uses Homebrew. If `brew` is not installed, install it first:

```zsh
/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"
```

After installation, if the Homebrew installer asks you to add Homebrew to your `PATH`, run the commands it prints. If `which brew` already returns `/opt/homebrew/bin/brew`, you can skip this and continue.

On most Apple Silicon Macs with `zsh`, the PATH setup commands will look like:

```zsh
echo 'eval "$(/opt/homebrew/bin/brew shellenv)"' >> ~/.zprofile
eval "$(/opt/homebrew/bin/brew shellenv)"
```

If you are not on macOS, or you prefer not to use Homebrew, install goenv using the method documented in the goenv repository and then continue with the shell setup below.

### 2. Install goenv

```zsh
brew install goenv
```

Add the following to your shell profile (`~/.zshrc` or `~/.bash_profile`):

```zsh
export GOENV_ROOT="$HOME/.goenv"
export PATH="$GOENV_ROOT/bin:$PATH"
eval "$(goenv init -)"
```

Then reload your shell:

```zsh
source ~/.zshrc
```

### 3. Install the required Go version

From the project root:

```zsh
goenv install 1.25.0
```

goenv will automatically use this version inside the project directory (via `.go-version`).

### 4. Verify

```zsh
go version
# should print: go version go1.25.0 ...
```

## Usage

```zsh
make build   # compile to bin/app
make test    # run all tests
```