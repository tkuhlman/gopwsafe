# gopwsafe

[![GoDoc](https://godoc.org/github.com/tkuhlman/gopwsafe?status.svg)](https://godoc.org/github.com/tkuhlman/gopwsafe)
[![Build Status](https://travis-ci.org/tkuhlman/gopwsafe.svg)](https://travis-ci.org/tkuhlman/gopwsafe)
[![Coverage Status](https://coveralls.io/repos/tkuhlman/gopwsafe/badge.svg?branch=master&service=github)](https://coveralls.io/github/tkuhlman/gopwsafe?branch=master)

A password safe written in go using  and implementing the [password safe](http://pwsafe.org/) version 3 database.
Simply download and run, no install needed.

The pwsafe package contains interfaces for reading/writing to Password Safe v3 databases. This package is utilized by both the gui and cli interfaces with the
preference going to the gtk based gui library.

The gui is implemented with the library [gotk3](https://github.com/gotk3/gotk3)
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

Go dependencies are managed with [deb](https://github.com/golang/dep), ie run `deb ensure`.

The gotk3 project is dependent on gtk3 and so that must be installed to build.
Details on this installation are in the gotk3 [project wiki](https://github.com/gotk3/gotk3/wiki#installation).

After the dependencies are installed a normal go build is all that is needed.
The build make target in the project will create a fully static build so the machines where the project is run need not have the build dependencies installed.
