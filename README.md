# GoTTY - Share your terminal as a web application

[![wercker status](https://app.wercker.com/status/03b91f441bebeda34f80e09a9f14126f/s/master "wercker status")](https://app.wercker.com/project/bykey/03b91f441bebeda34f80e09a9f14126f)

GoTTY is a simple command line tool that turns your CLI tools into web applications.

![Screenshot](https://raw.githubusercontent.com/yudai/gotty/master/screenshot.gif)

# Installation

Download the latest binary file from the [Releases](https://github.com/yudai/gotty/releases) page.

# Usage


```sh
Usage: gotty COMMAND_NAME [COMMAND_ARGUMENTS...]
```

Run `gotty` with your prefered command as its arguments (e.g. `gotty top`).

By default, gotty starts a web server at port 8080. Open the URL on your web browser and you can see the running command as if it's running on your terminal.

## Options

```
--addr, -a           IP address to listen at
--port, -p "8080"    Port number to listen at
--permit-write, -w   Permit write from client (BE CAREFUL)
```

By default, gotty doesn't allow clients to send any keystrokes or commands except terminal window resizing. When you want to permmit clients to write input to the PTY, add the `-w` option. However, accepting input from remote clients is dangerous for most commands. Make sure that only trusted clients can connect to your gotty server when activate this option. If you need interaction with the PTY, consider starting gotty with tmux or GNU Screen and run your main command on it.

# License

The MIT License
