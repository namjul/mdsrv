# mdsrv

mdsrv is a utility for serving Markdown files over HTTP. You point it at a
directory, and the server will map the request path to a Markdown file on the
filesystem.

## Quickstart

If you have Go installed then you can simply clone this repository, and build
it from source via `go build`.

    $ git clone https://github.com/andrewpillar/mdsrv
    $ cd mdsrv
    $ go build -tags "netgo"

Once built you can use it to start serving files,

    $ ./mdsrv

by default it will bind to `:8080`, and serve files from the current directory.
