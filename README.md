# gopwsafe

** In Progress **

A password safe written in go using  and implementing the [password safe](http://pwsafe.org/) version 3 database.
Simply download and run, no install needed.

The pwsafe package contains interfaces for reading/writing to Password Safe v3 databases. This package is utilized by both the gui and cli interfaces with the
preference going to the gtk based gui library.

The gui is implemented with the [go-gtk](https://github.com/mattn/go-gtk) library.
All cli is implemented using the ??? library

The project has been largely tested and developed on OS X and Linux (Ubuntu).

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
    - Write unit test
    - Finish implemenation of all record fields
    - Finish implementation of all header fields.
- Additional Features
  - Edit of multiple entries at once for select fields, ie modify the group
  - The ability copy/move entries from one open db to another.
  - The ability to diff two different databases.
    - Full diff
    - diff based on particular fields, name, username, url, password
  - Automatic storage of old passwords.
- Setup travis-ci, possibly readthedocs
