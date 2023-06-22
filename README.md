# Code harvest
This is a side project that I created because I thought that it would be fun to
extract some metadata from my coding sessions, and display it on my [website][1]:

![Screenshot of website][2]

![Screenshot of website][3]

## Overview
The project is divided into six separate components:

- RPC server for creating coding sessions
- client for sending remote procedure calls to the server
- cli for aggregating the data by different time periods
- vim plugin for mapping autocommands to remote procedure calls
- api for serving the data
- website for displaying the data

The server, client, and cli is part of this repository. The neovim plugin can
be found [here][4].

### Server
The server operates in the background as a daemon. It handles remote procedure
calls from neovim pertaining to events such as the opening of buffers,
windows gaining focus, the initiation of new neovim instances, etc.

For each instance of neovim, I establish a new coding session. This leads to
the creation of multiple sessions per day. Every session is stored temporarily
on the file system. This is primarily to avoid surpassing any limits set by
free database tiers.

### Client
The client uses the neovim [go-client][5] to add commands to neovim which I have
mapped to autocommands in my neovim [plugin][4].

### CLI for aggregation
I use the CLI to cluster raw coding sessions by day, and daily coding sessions by
week, month, and year.

[1]: https://creativecreature.com
[2]: ./screenshots/website1.png
[3]: ./screenshots/website2.png
[4]: https://github.com/creativecreature/vim-code-harvest
[5]: https://github.com/neovim/go-client
