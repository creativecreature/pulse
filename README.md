# Code harvest
Harvesting metadata from your coding sessions.

## Background
I've enjoyed building custom charts ever since I started to code. I also think
it's fun to get some data-driven insights about the work I do.

Therefore, I've created this project to generate some statistics about my
coding sessions.

The sessions are aggregated on a daily basis, and some of the information can
be viewed on my [website][1]

![Screenshot of website][2]

![Screenshot of website][3]

## Overview
I was heavily inspired by how language servers use remote procedure calls to
communicate. The latency is low, and it makes it easy to add plugins for other
editors in the future.

I run the RPC server as a daemon. I've created a small [plugin][4] for neovim
that maps different autocommands to remote procedure calls on the server.

The server writes each session to a mongodb database. You can provide a URI in
the `.envrc` file.

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
your own machine. The binaries you build are going to write to that database by
default (given that you don't modify any of the settings).

If you don't want to use mongodb you could easily change that by modifying the
`saveSession` function. You could, for example, write the sessions to disk, send
them to some API of yours, or use an entirely different database.

## Making use of the data
The server is going to create a lot of sessions. I use TMUX with many splits,
and I often have multiple instances of vim running at the same time. I don't
want the sessions time to multiply by the number of running instances.

Therefore, everytime I focus a new instance of vim, I end the previous session
and create a new one. I then use a cron to aggregate all of the sessions into
a summary.

I would suggest creating indexes for the `started_at` and `ended_at` fields.
That will make it quick to aggregate the sessions on a daily, weekly, or even
monthly basis.

I do have plans to make some changes to this in the future. When I do I will
update the documentaion accordingly.

[1]: https://creativecreature.com
[2]: ./screenshots/website1.png
[3]: ./screenshots/website2.png
[4]: https://github.com/creativecreature/vim-code-harvest
