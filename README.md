# `pulse`: like a fitness tracker for your coding sessions

[![Go Reference](https://pkg.go.dev/badge/github.com/creativecreature/pulse.svg)](https://pkg.go.dev/github.com/creativecreature/pulse)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://github.com/creativecreature/pulse/blob/master/LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/creativecreature/pulse)](https://goreportcard.com/report/github.com/creativecreature/pulse)
[![Test](https://github.com/creativecreature/pulse/actions/workflows/main.yml/badge.svg?branch=main)](https://github.com/creativecreature/pulse/actions/workflows/main.yml)

This repository contains all of the code for gathering the data which I display
on my [website][1]

![Screenshot of website][2]

![Screenshot of website][3]

# How it works
After spending some time debugging different language servers in Neovim, I felt
inspired to write my own server that would simply parse metadata and aggregate
statistics about my coding sessions so that I could display it on my website.

I launch the server from this repository as a daemon every time my laptop
boots. It then receives remote procedure calls from the neovim plugin
pertaining to events such as the opening of buffers, windows gaining focus, the
initiation of new `nvim` processes, etc.

These calls contains the path to the buffer, which the server parses and writes
to an append-only log-structured key-value store. Every segment is roughly 10KB
on disk. The server requests all of the buffers from this KV store every 15
minutes, and proceeds to aggregate them to a remote MongoDB database.

I chose this approach primarily because I wanted to build a log-structured
storage engine in order to better understand the inner workings of some popular
databases. It is a work in progress, but it now includes some core features
such as hash indexes, segmentation, and compaction. As a bonus, it helps me
avoid surpassing the limits set by the free tier for the MongoDB database!

This project has evolved into a bit of a playground where I can experiment with
different ideas and technologies. The most challenging part so far has been
designing the website, as I wanted it to have a unique look and feel and to
build all of the components from scratch.


[1]: https://conner.dev
[2]: ./screenshots/website1.png
[3]: ./screenshots/website2.png
