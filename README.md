# ![](https://raw.githubusercontent.com/yudai/gotty/master/resources/favicon.png) GoTTY - Share your terminal as a web application

[![GitHub release](http://img.shields.io/github/release/yudai/gotty.svg?style=flat-square)][release]
[![Wercker](http://img.shields.io/wercker/ci/55d0eeff7331453f0801982c.svg?style=flat-square)][wercker]
[![MIT License](http://img.shields.io/badge/license-MIT-blue.svg?style=flat-square)][license]
[![Join the chat at https://gitter.im/yudai/gotty](http://img.shields.io/badge/gitter-join%20chat%20%E2%86%92-brightgreen.svg?style=flat-square)][gitter]

[release]: https://github.com/yudai/gotty/releases
[wercker]: https://app.wercker.com/project/bykey/03b91f441bebeda34f80e09a9f14126f
[license]: https://github.com/yudai/gotty/blob/master/LICENSE
[gitter]: https://gitter.im/yudai/gotty?utm_source=badge&utm_medium=badge&utm_campaign=pr-badge&utm_content=badge

GoTTY is a simple command line tool that turns your CLI tools into web applications.

![Screenshot](https://raw.githubusercontent.com/yudai/gotty/master/screenshot.gif)

# Installation

Download the latest binary file from the [Releases](https://github.com/yudai/gotty/releases) page.

(`darwin_amd64.tar.gz` is for Mac OS X users)

## Homebrew Installation

You can install GoTTY with [Homebrew](http://brew.sh/) as well.

```sh
$ brew install yudai/gotty/gotty
```

## `go get` Installation

If you have a Go language environment, you can install GoTTY with the `go get` command.

```sh
$ go get github.com/yudai/gotty
```

# Usage

```
Usage: gotty [options] <command> [<arguments...>]
```

Run `gotty` with your preferred command as its arguments (e.g. `gotty top`).

By default, GoTTY starts a web server at port 8080. Open the URL on your web browser and you can see the running command as if it were running on your terminal.

## Options

```
--address, -a                                                IP address to listen [$GOTTY_ADDRESS]
--port, -p "8080"                                            Port number to listen [$GOTTY_PORT]
--permit-write, -w                                           Permit clients to write to the TTY (BE CAREFUL) [$GOTTY_PERMIT_WRITE]
--credential, -c                                             Credential for Basic Authentication (ex: user:pass, default disabled) [$GOTTY_CREDENTIAL]
--random-url, -r                                             Add a random string to the URL [$GOTTY_RANDOM_URL]
--random-url-length "8"                                      Random URL length [$GOTTY_RANDOM_URL_LENGTH]
--tls, -t                                                    Enable TLS/SSL [$GOTTY_TLS]
--tls-crt "~/.gotty.crt"                                     TLS/SSL certificate file path [$GOTTY_TLS_CRT]
--tls-key "~/.gotty.key"                                     TLS/SSL key file path [$GOTTY_TLS_KEY]
--tls-ca-crt "~/.gotty.ca.crt"                               TLS/SSL CA certificate file for client certifications [$GOTTY_TLS_CA_CRT]
--index                                                      Custom index.html file [$GOTTY_INDEX]
--title-format "GoTTY - {{ .Command }} ({{ .Hostname }})"    Title format of browser window [$GOTTY_TITLE_FORMAT]
--reconnect                                                  Enable reconnection [$GOTTY_RECONNECT]
--reconnect-time "10"                                        Time to reconnect [$GOTTY_RECONNECT_TIME]
--timeout "0"                                                Timeout seconds for waiting a client (0 to disable) [$GOTTY_TIMEOUT]
--once                                                       Accept only one client and exit on disconnection [$GOTTY_ONCE]
--permit-arguments                                           Permit clients to send command line arguments in URL (e.g. http://example.com:8080/?arg=AAA&arg=BBB) [$GOTTY_PERMIT_ARGUMENTS]
--close-signal "1"                                           Signal sent to the command process when gotty close it (default: SIGHUP) [$GOTTY_CLOSE_SIGNAL]
--config "~/.gotty"                                          Config file path [$GOTTY_CONFIG]
--version, -v                                                print the version
```

### Config File

You can customize default options and your terminal (hterm) by providing a config file to the `gotty` command. GoTTY loads a profile file at `~/.gotty` by default when it exists.

```
// Listen at port 9000 by default
port = "9000"

// Enable TSL/SSL by default
enable_tls = true

// hterm preferences
// Smaller font and a little bit bluer background color
preferences {
    font_size = 5
    background_color = "rgb(16, 16, 32)"
}
```

See the [`.gotty`](https://github.com/yudai/gotty/blob/master/.gotty) file in this repository for the list of configuration options.

### Security Options

By default, GoTTY doesn't allow clients to send any keystrokes or commands except terminal window resizing. When you want to permit clients to write input to the TTY, add the `-w` option. However, accepting input from remote clients is dangerous for most commands. When you need interaction with the TTY for some reasons, consider starting GoTTY with tmux or GNU Screen and run your command on it (see "Sharing with Multiple Clients" section for detail).

To restrict client access, you can use the `-c` option to enable the basic authentication. With this option, clients need to input the specified username and password to connect to the GoTTY server. Note that the credentical will be transmitted between the server and clients in plain text. For more strict authentication, consider the SSL/TLS client certificate authentication described below.

The `-r` option is a little bit casualer way to restrict access. With this option, GoTTY generates a random URL so that only people who know the URL can get access to the server.  

All traffic between the server and clients are NOT encrypted by default. When you send secret information through GoTTY, we strongly recommend you use the `-t` option which enables TLS/SSL on the session. By default, GoTTY loads the crt and key files placed at `~/.gotty.crt` and `~/.gotty.key`. You can overwrite these file paths with the `--tls-crt` and `--tls-key` options. When you need to generate a self-signed certification file, you can use the `openssl` command.

```sh
openssl req -x509 -nodes -days 9999 -newkey rsa:2048 -keyout ~/.gotty.key -out ~/.gotty.crt
```

(NOTE: For Safari uses, see [how to enable self-signed certificates for WebSockets](http://blog.marcon.me/post/24874118286/secure-websockets-safari) when use self-signed certificates)

For additional security, you can use the SSL/TLS client certificate authentication by providing a CA certificate file to the `--tls-ca-crt` option (this option requires the `-t` or `--tls` to be set). This option requires all clients to send valid client certificates that are signed by the specified certification authority.

## Sharing with Multiple Clients

GoTTY starts a new process with the given command when a new client connects to the server. This means users cannot share a single terminal with others by default. However, you can use terminal multiplexers for sharing a single process with multiple clients.

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
# Start GoTTY in a new window with C-t
bind-key C-t new-window "gotty tmux attach -t `tmux display -p '#S'`"
```

## Playing with Docker

When you want to create a jailed environment for each client, you can use Docker containers like following:

```sh
$ gotty -w docker run -it --rm busybox
```

## Development

You can build a binary using the following commands. Windows is not supported now.

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

GoTTY uses [hterm](https://groups.google.com/a/chromium.org/forum/#!forum/chromium-hterm) to run a JavaScript based terminal on web browsers. GoTTY itself provides a websocket server that simply relays output from the TTY to clients and receives input from clients and forwards it to the TTY. This hterm + websocket idea is inspired by [Wetty](https://github.com/krishnasrinivas/wetty).

## Alternatives

### Command line client

* [gotty-client](https://github.com/moul/gotty-client): If you want to connect to GoTTY server from your terminal

### Terminal/SSH on Web Browsers

* [Secure Shell (Chrome App)](https://chrome.google.com/webstore/detail/secure-shell/pnhechapfaindjhompbnflcldabbghjo): If you are a chrome user and need a "real" SSH client on your web browser, perhaps the Secure Shell app is what you want
* [Wetty](https://github.com/krishnasrinivas/wetty): Node based web terminal (SSH/login)
* [ttyd](https://tsl0922.github.io/ttyd): C port of GoTTY with CJK and IME support

### Terminal Sharing

* [tmate](http://tmate.io/): Forked-Tmux based Terminal-Terminal sharing
* [termshare](https://termsha.re): Terminal-Terminal sharing through a HTTP server
* [tmux](https://tmux.github.io/): Tmux itself also supports TTY sharing through SSH)

# License

The MIT License
