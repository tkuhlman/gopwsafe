# gopwsafe

[![GoDoc](https://godoc.org/github.com/tkuhlman/gopwsafe?status.svg)](https://godoc.org/github.com/tkuhlman/gopwsafe)
[![Build Status](https://travis-ci.org/tkuhlman/gopwsafe.svg)](https://travis-ci.org/tkuhlman/gopwsafe)
[![Coverage Status](https://coveralls.io/repos/tkuhlman/gopwsafe/badge.svg?branch=master&service=github)](https://coveralls.io/github/tkuhlman/gopwsafe?branch=master)

A password safe written in go using  and implementing the [password safe](http://pwsafe.org/) version 3 database.
Simply download and run, no install needed.

The pwsafe package contains interfaces for reading/writing to Password Safe v3 databases. This package is utilized by both the gui and cli interfaces with the
preference going to the gtk based gui library.

The gui is implemented with the library [go-gtk](https://github.com/mattn/go-gtk/)
The very basic cli is implemented using the [prompt](https://github.com/Bowery/prompt) library.

The project has been largely tested and developed on Linux (Ubuntu).

# GTK GUI
Features:
- The ability to have multiple windows open with different databases in each is a key feature that many other Password Safe implementations don't have.
- Simple database search.
- Tree representation based on db and group.
- Keyboard shortcuts, for copy/paste, opening url in a browser, etc.

# References
- V3 Password Safe Specification - https://github.com/pwsafe/pwsafe/blob/master/docs/formatV3.txt

# Roadmap
- Add a timeout to clear the clipboard a minute or so after copying a password.
- The ability copy/move entries from one open db to another.
- Automatic storage of old passwords.
- Add a file selection tool for opening.
- The ability to diff two different databases.
  - Full diff
  - diff based on particular fields, name, username, url, password
- Edit of multiple entries at once for select fields, ie modify the group

# Building

The go-gtk project is dependent on gtk2 and so that must be installed to build.
On linux this is straight forward for example for Ubuntu:

    sudo apt-get install -y build-essential libgtk2.0-dev

Then simply go get this package:

    go get github.com/tkuhlman/gopwsafe

## Building for a Mac

On a mac getting the basic gtk dependencies is a bit more involved.
To install the dependencies you can use [brew](http://brew.sh) and as noted [here](https://github.com/mattn/go-gtk/issues/165) these
are the basic steps.

    $ brew install go
    $ brew install  brew install cairo pixman fontconfig freetype libpng gtksourceview
    $ go get github.com/mattn/go-gtk
    # Ensure they can be found by exporting the PKG_CONFIG_PATH.
    $ export PKG_CONFIG_PATH=":/usr/local/opt/cairo/lib/pkgconfig:/usr/local/opt/pixman/lib/pkgconfig:/usr/local/opt/fontconfig/lib/pkgconfig:/usr/local/opt/freetype/lib/pkgconfig:/usr/local/opt/libpng/lib/pkgconfig:/usr/X11/lib/pkgconfig:/usr/local/opt/gtksourceview/lib/pkgconfig:${PKG_CONFIG_PATH}"
    $ cd $GOPATH/src/github.com/mattn/go-gtk
    $ make all

Lastly just build the gopwsafe as normal for go.

    go get github.com/tkuhlman/gopwsafe
