# GoTTY - Share your terminal as a web application

[![GitHub release](http://img.shields.io/github/release/yudai/gotty.svg?style=flat-square)][release]
[![Wercker](http://img.shields.io/wercker/ci/55d0eeff7331453f0801982c.svg?style=flat-square)][wercker]
[![MIT License](http://img.shields.io/badge/license-MIT-blue.svg?style=flat-square)][license]

[release]: https://github.com/yudai/gotty/releases
[wercker]: https://app.wercker.com/project/bykey/03b91f441bebeda34f80e09a9f14126f
[license]: https://github.com/yudai/gotty/blob/master/LICENSE


GoTTY is a simple command line tool that turns your CLI tools into web applications.

![Screenshot](https://raw.githubusercontent.com/yudai/gotty/master/screenshot.gif)

# Installation

Download the latest binary file from the [Releases](https://github.com/yudai/gotty/releases) page.

(`darwin_amd64.tar.gz` is for Mac OS X users)

## Homebrew Installation

You can install gotty with [Homebrew](http://brew.sh/) as well.

```sh
$ brew tap yudai/gotty
$ brew install gotty
```

## `go get` Installation

If you have a Go language environment, you can install gotty with the `go get` command.

```sh
$ go get github.com/yudai/gotty
```

# Usage

```
Usage: gotty [options] <command> [<arguments...>]
```

Run `gotty` with your preferred command as its arguments (e.g. `gotty top`).

By default, gotty starts a web server at port 8080. Open the URL on your web browser and you can see the running command as if it's running on your terminal.

## Options

```
--addr, -a                                                   IP address to listen [$GOTTY_ADDR]
--port, -p "8080"                                            Port number to listen [$GOTTY_PORT]
--permit-write, -w                                           Permit clients to write to the TTY (BE CAREFUL) [$GOTTY_PERMIT_WRITE]
--credential, -c                                             Credential for Basic Authentication (ex: user:pass) [$GOTTY_CREDENTIAL]
--random-url, -r                                             Add a random string to the URL [$GOTTY_RANDOM_URL]
--profile-file, -f "~/.gotty"                                Path to profile file [$GOTTY_PROFILE_FILE]
--enable-tls, -t                                             Enable TLS/SSL [$GOTTY_ENABLE_TLS]
--tls-cert "~/.gotty.crt"                                    TLS/SSL cert [$GOTTY_TLS_CERT]
--tls-key "~/.gotty.key"                                     TLS/SSL key [$GOTTY_TLS_KEY]
--title-format "GoTTY - {{ .Command }} ({{ .Hostname }})"    Title format of browser window [$GOTTY_TITLE_FORMAT]
--auto-reconnect "-1"                                        Seconds to automatically reconnect to the server when the connection is closed (default: disabled) [$GOTTY_AUTO_RECONNECT]
```

### Profile File

You can customize your terminal (hterm) by providing a profile file to the `gotty` command, which is a JSON file that has a map of preference keys and values. Gotty loads a profile file at `~/.gotty` by default when it exists.

The following example makes the font size smaller and the background color a little bit blue.

```json
{
    "font-size": 5,
    "background-color": "rgb(16, 16, 32)"
}
```

Available preferences are listed in [the hterm source code](https://chromium.googlesource.com/apps/libapps/+/master/hterm/js/hterm_preference_manager.js)

### Security Options

By default, gotty doesn't allow clients to send any keystrokes or commands except terminal window resizing. When you want to permit clients to write input to the PTY, add the `-w` option. However, accepting input from remote clients is dangerous for most commands. When you need interaction with the PTY for some reasons, consider starting gotty with tmux or GNU Screen and run your main command on it (see "Sharing with Multiple Clients" section for detail).

To restrict client access, you can use the `-c` option to enable the basic authentication. With option, clients need to input the specified username and passwords to connect to the gotty server. The `-r` option is a little bit casualer way to restrict access. With this option, gotty generates a random URL so that only people who know the URL can access to the server.

All traffic between servers and clients are NOT encrypted by default. When you send secret information through gotty, we strongly recommend you use the `-t` option which enables TLS/SSL on the session. By default, gotty loads the cert and key files placed at `~/.gotty.cert` and `~/.gotty.key`. You can overwrite these file paths with the `--tls-cert` and `--tls-key` options. When you need to generate a self-sined certification file, you can use the `openssl` command.

```sh
openssl req -x509 -nodes -days 9999 -newkey rsa:2048 -keyout ~/.gotty.key -out ~/.gotty.crt
```

## Sharing with Multiple Clients

Gotty starts a new process when a new client connects to the server. This means users cannot share a single terminal with others by default. However, you can use terminal multiplexers for sharing a single process with multiple clients.

For example, you can start a new tmux session named `gotty` with `top` command by the command below.

```sh
$ gotty tmux new -A -s gotty top
```

This command doesn't allow clients to send keystrokes, however, you can attach the session from your local terminal and run operations like switching the mode of the `top` command. To connect to the tmux session from your terminal, you can use following command.

```sh
$ tmux new -A -s gotty
```

By using terminal multiplexers, you can have the control of your terminal and allow clients to just see your screen.

### Quick Sharing on tmux

To share your current session with others by a shortcut key, you can add a line like below to your `.tmux.conf`.

```
# Start gotty in a new window with C-t
bind-key C-t new-window "gotty tmux attach -t `tmux display -p '#S'`"
```

## Playing with Docker

When you want to create a jailed environment for each client, you can use Docker containers like following:

```sh
$ gotty -w docker run -it --rm busybox
```

## Development

You can build a binary yourself using following commands. Windows is not supported now.

```sh
# Install tools
go get github.com/jteeuwen/go-bindata/...
go get github.com/tools/godep

# Checkout hterm
git submodule sync && git submodule update --init --recursive

# Restore libraries in Godeps
godep restore

# Build
make
```

## Architecture

Gotty uses [hterm](https://groups.google.com/a/chromium.org/forum/#!forum/chromium-hterm) to run a JavaScript based terminal on web browsers. Gotty itself provides a websocket server that simply relays output from the PTY to clients and receives input from clients and forwards to the PTY. This hterm + websocket idea is highly inspired by [Wetty](https://github.com/krishnasrinivas/wetty).

# License

The MIT License
