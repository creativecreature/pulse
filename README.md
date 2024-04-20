# `pulse`: like a fitness tracker for code

[![Go Reference](https://pkg.go.dev/badge/github.com/creativecreature/pulse.svg)](https://pkg.go.dev/github.com/creativecreature/pulse)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://github.com/creativecreature/pulse/blob/master/LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/creativecreature/pulse)](https://goreportcard.com/report/github.com/creativecreature/pulse)
[![Test](https://github.com/creativecreature/pulse/actions/workflows/main.yml/badge.svg?branch=main)](https://github.com/creativecreature/pulse/actions/workflows/main.yml)
[![codecov](https://codecov.io/gh/creativecreature/pulse/graph/badge.svg?token=CYSKW3Z7E6)](https://codecov.io/gh/creativecreature/pulse)

My vision with this project was to create a personal portfolio [website][1] that
would update automatically with data from my coding sessions:

![Screenshot of website][2]

![Screenshot of website][3]

So far it has been a really fun project!

# How it works
This repository includes the foundation of the project:
- `RPC server`
- `RPC client`
- `nvim plugin`
- `cli`


I run the server in the background as a daemon. It receives remote procedure
calls from neovim pertaining to events such as the opening of buffers, windows
gaining focus, the initiation of new neovim instances, etc:

https://github.com/creativecreature/code-harvest/assets/12787673/3732845c-694a-4951-9f4e-e0f6003a1e2f

As you can see in the video above, each instance of neovim establishes a new
coding session. This leads to the creation of several sessions per day. Every
session is stored temporarily on the file system. This is primarily to avoid
surpassing any limits set by free database tiers.

The CLI aggregates these temporary sessions by day, week, month, and year. The
results are then written to a more permanent storage:

https://github.com/creativecreature/code-harvest/assets/12787673/7a94c640-c230-4e20-aa28-b20ea016c72a

[1]: https://conner.dev
[2]: ./screenshots/website1.png
[3]: ./screenshots/website2.png
