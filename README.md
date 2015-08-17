# GoTTY - Share your terminal as a web application

[![wercker status](https://app.wercker.com/status/03b91f441bebeda34f80e09a9f14126f/s/master "wercker status")](https://app.wercker.com/project/bykey/03b91f441bebeda34f80e09a9f14126f)

GoTTY is a simple command line tool that turns your CLI tools into web applications.

![Screenshot](https://raw.githubusercontent.com/yudai/gotty/master/screenshot.gif)

# Installation

Download the latest binary file from the [Releases](https://github.com/yudai/gotty/releases) page.

## `go get` Instalation

If you have a Go language environment, you can install gotty with the `go get` command.

```sh
$ go get github.com/yudai/gotty
```

# Usage

```
Usage: gotty [options] <command> [<arguments...>]
```

Run `gotty` with your prefered command as its arguments (e.g. `gotty top`).

By default, gotty starts a web server at port 8080. Open the URL on your web browser and you can see the running command as if it's running on your terminal.

## Options

```
--addr, -a           IP address to listen [$GOTTY_ADDR]
--port, -p "8080"    Port number to listen [$GOTTY_PORT]
--permit-write, -w   Permit clients to write to the TTY (BE CAREFUL) [$GOTTY_PERMIT_WRITE]
```

By default, gotty doesn't allow clients to send any keystrokes or commands except terminal window resizing. When you want to permmit clients to write input to the PTY, add the `-w` option. However, accepting input from remote clients is dangerous for most commands. Make sure that only trusted clients can connect to your gotty server when activate this option. If you need interaction with the PTY, consider starting gotty with tmux or GNU Screen and run your main command on it.

## Sharing with Multiple Clients

Gotty starts a new process when a new client connects to the server. This means users cannot share a single terminal with others by default. However, you can use terminal multiplexers for sharing a single process with multiple clients.

For example, you can start a new tmux session named `gotty` with `top` command by the command below.

```sh
$ gotty tmux new -A -s gotty top
```

This command doesn't allow clients to send keystrokes, however, you can attach the session from your local terminal and run operatitons like switching the mode of the `top` command. To connect to the tmux session from your terminal, you can use following command.

```sh
$ tmux new -A -s gotty
```

By using terminal multiplexers, you can have the control of your terminal and allow clients to just see your screen.

## Playing with Docker

When you want to create a jailed environment for each client, you can use Docker containers like following:

```sh
$ gotty -w docker run -it --rm busybox
```

# License

The MIT License
