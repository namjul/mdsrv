# mdsrv

mdsrv is a utility for serving Markdown files over HTTP. You point it at a
directory, and the server will map the request path to a Markdown file on the
filesystem.

* [Quickstart](#quickstart)
* [Serving Markdown](#serving-markdown)
* [Template File](#template-file)
* [Enabling TLS](#enabling-tls)

## Quickstart

If you have Go installed then you can simply clone this repository, and build
it from source via `go build`.

    $ git clone https://github.com/andrewpillar/mdsrv
    $ cd mdsrv
    $ ./make.sh

Once built you can use it to start serving files,

    $ ./mdsrv.out
    2006/01/02 15:04:05 INFO   serving markdown documents in . on :8080

by default it will bind to `:8080`, and serve files from the current directory.
If no template file is specified, then the plain text of the Markdown files will
be served, instead of being parsed.

## Serving Markdown

mdsrv allows for serving Markdown either as their plain text files, or as
rendered HTML. By default, if no template file is given, then all of the
Markdown files will be served as plain text.

By default mdsrv will treat all `readme.md` files as the index for a directory
that is being served. So if the URI points to a directory, or is `/`, then
mdsrv will look for a `readme.md` to try and serve. Each URI is mapped directly
to a corresponding Markdown file with the appended `.md` suffix. For example,
if the URI `/docs/user` is requested, then mdrsrv will look for a file at the
location of `/docs/user.md`.

The plain text of a Markdown file can be viewed by setting the `Accept` header
in the HTTP request to `text/plain`.

## Template File

The template file is used for structuring how each Markdown document will be
rendered when served as HTML. This is specified via the `-tmpl` flag given to
`mdsrv`,

    $ mdsrv -tmpl document.tmpl

the template file will have access to two variables, `Title`, and `Document`.
The `Title` variable will be the title of the document that is derived from the
name of the Markdown file itself. The `Document ` variable will be the rendered
HTML of the Markdown document. You can then use these values to create a
template page for each Markdown file to be rendered into,

    <!DOCTYPE HTML>
    <html lang="en">
        <head>
            <meta charset="utf-8">
            <title>{{ .Title }}</title>
        </head>
        <body>{{ .Document }}</body>
    </html>

## Enabling TLS

TLS can be enabled via the `-cert` and `-key` flags. This take paths to the TLS
certificate and key to use respectively.

    $ mdsrv -cert /var/lib/ssl/server.crt -key /var/lib/ssl/server.key
