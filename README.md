# `pulse`: like a fitness tracker for your coding sessions

[![Go Reference](https://pkg.go.dev/badge/github.com/creativecreature/pulse.svg)](https://pkg.go.dev/github.com/creativecreature/pulse)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://github.com/creativecreature/pulse/blob/master/LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/creativecreature/pulse)](https://goreportcard.com/report/github.com/creativecreature/pulse)
[![Test](https://github.com/creativecreature/pulse/actions/workflows/main.yml/badge.svg?branch=main)](https://github.com/creativecreature/pulse/actions/workflows/main.yml)

This repository contains all of the code I'm using to gather data for my
[website.][1]

![Screenshot of website][2]

![Screenshot of website][3]

# How it works

After spending some time debugging different language servers in Neovim, I got
the idea to write my own RPC server that would simply parse metadata and
aggregate statistics about my coding sessions.

I run the server from this repository as a daemon, and it receives remote
procedure calls from the neovim plugin pertaining to events such as the opening
of buffers, windows gaining focus, the initiation of new `nvim` processes, etc.

These calls contains the path to the buffer, which the server parses and writes
to a log-structured append-only key-value store. The store is a work in
progress, but it now includes some core features such as hash indexes,
segmentation, and compaction.

The server runs a background job which requests all of the buffers from the KV
store, and proceeds to aggregate them to a remote database. I did it this way
primarily because I wanted to avoid surpassing the limits set by the free tier
for the remote database. If you aren't concerned about costs you could use a
much lower aggregation interval than me.

The only things that aren't included in this repository is the API which
retrieves the data and the website that displays it. The website has been the
most challenging part so far. I wanted it to have a unique look and feel and to
build all of the components from scratch. I'm in the process of making it open
source, but there are still a few things that I'd like to clean up first!

# Running this project

## 1. Download the binaries
Download and unpack the server **and** client binaries from the [releases](https://github.com/creativecreature/sturdyc/releases).
Next, you'll want to make sure that they are reachable from your `$PATH`.

## 2. Create a configuration file
Create a configuration file. It should be located at `$HOME/.pulse/config.yaml`

```yml
server:
  name: "pulse-server"
  hostname: "localhost"
  port: "1122"
  aggregationInterval: "15m"
  segmentationInterval: "5m"
  segmentSizeKB: "10"
database:
  address: "redis-<PORT>.xxxxxxxx.redis-cloud.com:<PORT>"
  password: "xxxxxxxx"
```

## 3. Launch the server as a daemon
On linux, you can setup a systemd service to run the server, and on macOS you
can create a launch daemon.

I'm using a Mac, and my launch daemon configuration looks like this:

```xml
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
  <dict>

    <key>Label</key>
    <string>dev.conner.pulse.plist</string>

    <key>RunAtLoad</key>
    <true/>

    <key>StandardErrorPath</key>
		<string>/Users/conner/.pulse/logs/stderr.log</string>

    <key>StandardOutPath</key>
		<string>/Users/conner/.pulse/logs/stdout.log</string>

    <key>EnvironmentVariables</key>
    <dict>
      <key>PATH</key>
      <string><![CDATA[/usr/local/bin:/usr/local/sbin:/usr/bin:/bin:/usr/sbin:/sbin]]></string>
    </dict>

    <key>WorkingDirectory</key>
    <string>/Users/conner</string>

    <key>ProgramArguments</key>
    <array>
			<string>/Users/conner/bin/pulse-server</string>
    </array>

		<key>KeepAlive</key>
    <true/>

  </dict>
</plist>
```

## 4. Install the neovim plugin
Here is an example using lazy.nvim:

```lua
return {
	-- Does not require any configuration.
	{ "creativecreature/pulse" },
}
```

[1]: https://conner.dev
[2]: ./screenshots/website1.png
[3]: ./screenshots/website2.png
