# Code harvest
Harvesting metadata from your coding sessions.

## Overview
I created this project to gather some statistics about my coding sessions.

I was heavily inspired by how language servers use remote procedure calls to
communicate. The latency is low, and it would also make it easy to add plugins
for multiple editors in the future.

I run the RPC server as a daemon. I've created a small [plugin][4] for neovim
that maps some autocommands to remote procedure calls.

The server saves every coding session in a mongodb database.

## Building
The `Makefile` has a **build** target for compiling both the server and client.
To get more information about the available targets run `make help`.

## Running your own version
Start by building the binaries. The server should run as a daemon. Depending on
your OS I suggest that you use either `systemd` or `launchd`. Windows has a
multitude of alternatives too, but I've never used them.

If you use neovim you can use the plugin I created. Just make sure the binaries
are in your `$PATH`.

You are also going to need a mongodb database. To start with I would suggest
going to the official mongodb website and installing the community edition on
your machine. The binaries you build uses the default uri. You can change that
in the `.envrc` file

Each instance of neovim has its own coding session. That means that the server
is going to create a lot of them. I use TMUX with many splits, and I often have
multiple instances of neovim running at the same time. I don't want the time of
my coding sessions to multiply by the number of neovim instances that I have
running. Therefore, everytime I focus a new instance of neovim, I end the
previous session and create a new one. I then use a cron to aggregate all of
the sessions into a summary. The cron is currently not part of this repository.

I would also suggest creating indexes for the `started_at` and `ended_at`
fields. That will make it quick to aggregate the sessions on a daily, weekly,
or even monthly basis.

## Examples
I use the data that this project generates to power a dashboard on my [website][1]

Here are some screenshots of what it looks like:

![Screenshot of website][2]

![Screenshot of website][3]


[1]: https://conner.dev
[2]: ./screenshots/website1.png
[3]: ./screenshots/website2.png
[4]: https://github.com/creativecreature/vim-code-harvest
