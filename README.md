# gopwsafe
A password safe written in go using [go-gtk](https://github.com/mattn/go-gtk) and implementing the [password safe](http://pwsafe.org/) version 3 database.

The project consists for 3 components:
- pwsafe contains interfaces for reading/writing to Password Safe v3 databases, all of which is part of the pwsafe package
- gtk contains the gtk based gui leveraging the db library, gtk package
- cli contains a command line based interface leveraging the library, cli package

The project has been largely tested and developed on OS X and Linux (Ubuntu).

# GTK GUI
Features:
- The ability to have multiple windows open with different databases in each is a key feature I want.
- Simple database search.
- Tree representation based on group.
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
  - The ability to diff two different databases.
    - Full diff
    - diff based on particular fields, name, username, url, password
  - Automatic storage of old passwords.
