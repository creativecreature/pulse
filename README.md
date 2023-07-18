# Code Harvest
I use this project to derive metadata from my coding sessions. Some of it can
be viewed on my [website][1]

![Screenshot of website][2]

![Screenshot of website][3]

The project includes a server that operates in the background as a daemon. It
receives remote procedure calls from neovim pertaining to events such as the
opening of buffers, windows gaining focus, the initiation of new neovim
instances, etc:

As you can see in the video above, each instance of neovim establishes a new
coding session. This leads to the creation of several sessions per day. Every
session is stored temporarily on the file system. This is primarily to avoid
surpassing any limits set by free database tiers. There is a CLI which can be
used to subsequently cluster these sessions by day, week, month, and year. The
results are then written to a more permanent storage:

The neovim plugin that I've created for sending the remote procedure calls to
the server can be found [here][4].

[1]: https://creativecreature.com
[2]: ./screenshots/website1.png
[3]: ./screenshots/website2.png
[4]: https://github.com/creativecreature/vim-code-harvest
