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

It has served as a playground for trying out new technologies, and so far it
has been a really fun project!

# How it works
This repository includes the foundation of the project:
- `rpc server`
- `rpc client`
- `log-structured KV store`
- `nvim plugin`
- `cli`


I run the server as a daemon. It receives remote procedure calls from neovim
pertaining to events such as the opening of buffers, windows gaining focus, the
initiation of new `nvim` processes, etc:


https://github.com/creativecreature/pulse/assets/12787673/c1cc1dcb-47c3-48c4-a694-056e79f186fe

The buffers I edit are written to an append-only log-structured key-value store
which performs compaction and auto-segmentation. Every segment is roughly 10KB
on disk.

Every 15 minutes, the server requests all buffers from the KV store, and writes
them to a remote mongodb database. I'm doing this primarily to avoid surpassing
any limits set by the free tier.

[1]: https://conner.dev
[2]: ./screenshots/website1.png
[3]: ./screenshots/website2.png
