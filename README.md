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

# Todo
- Add the ability to create a new empty password db.
- pwsafe
    - Finish implemenation of all record fields
    - Finish implementation of all header fields.
- Add a timeout to clear the clipboard a minute or so after copying a password.
- The ability copy/move entries from one open db to another.
- Prompt on exit if there are unsaved changes.

## Eventually
- Automatic storage of old passwords.
- Add a file selection tool for opening.
- The ability to diff two different databases.
  - Full diff
  - diff based on particular fields, name, username, url, password
- Edit of multiple entries at once for select fields, ie modify the group
- Look at gomobile would it be possible to write my code in such a way it can be used on Android and ios. See the utils talk slides for more details on gomobile.
