# gopwsafe

[![GoDoc](https://godoc.org/github.com/tkuhlman/gopwsafe?status.svg)](https://godoc.org/github.com/tkuhlman/gopwsafe)
[![Build Status](https://travis-ci.org/tkuhlman/gopwsafe.svg)](https://travis-ci.org/tkuhlman/gopwsafe)
[![Coverage Status](https://coveralls.io/repos/tkuhlman/gopwsafe/badge.svg?branch=master&service=github)](https://coveralls.io/github/tkuhlman/gopwsafe?branch=master)


** In Progress **

A password safe written in go using  and implementing the [password safe](http://pwsafe.org/) version 3 database.
Simply download and run, no install needed.

The pwsafe package contains interfaces for reading/writing to Password Safe v3 databases. This package is utilized by both the gui and cli interfaces with the
preference going to the gtk based gui library.

The gui is implemented with the library [google/gxui](https://github.com/google/gxui)
The very basic cli is implemented using the [prompt](https://github.com/Bowery/prompt) library.

The project has been largely tested and developed on Linux (Ubuntu).

# GTK GUI
Features:
- The ability to have multiple windows open with different databases in each is a key feature that many other Password Safe implementations don't have.
- Simple database search.
- Tree representation based on db and group.
- Copy/Paste shortcuts with timeout on pasting of sensitive fields.
- Shortcut key to open user preferred browser.

# References
- V3 Password Safe Specification - http://sourceforge.net/p/passwordsafe/code/HEAD/tree/trunk/pwsafe/pwsafe/docs/formatV3.txt

# Todo
- Everything listed above in features needs finishing
- pwsafe
    - Finish implemenation of all record fields
    - Finish implementation of all header fields.
- Add the ability to create a new empty password db.
- Add opening of more than one file at a time.
- Keep list of recent files opened, do this in a way that it works for both interfaces.
- The ability copy/move entries from one open db to another.
- Edit of multiple entries at once for select fields, ie modify the group

## Eventually
- The ability to diff two different databases.
  - Full diff
  - diff based on particular fields, name, username, url, password
- Automatic storage of old passwords.
- Look at gomobile would it be possible to write my code in such a way it can be used on Android and ios. See the utils talk slides for more details on gomobile.
