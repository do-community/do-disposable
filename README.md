# do-disposable
Allows you to create a disposable droplet with a simple CLI interface. This is useful for trying to run actions which may be network intensive to do on your standard internet connection or too CPU intensive for your system since you can start a droplet to the size you need and then have it easily disposed of when you're done.

You can find a binary for your operating system in the [releases page](https://github.com/JakeMakesStuff/do-disposable/releases).

The following sub-commands are included:
- `auth`: Authenticates the user and creates the configuration if this doesn't exist. This is the sub-command you will be prompted to run on first launch of this tool.
- `setregion`: Allows you to modify the region. Note that you need to go through the setup with do-disposable auth first (that will also configure this for the first time).
- `setsize`: Allows you to modify the droplet size. Note that you need to go through the setup with do-disposable auth first (that will also configure this for the first time).
- `up`: Allows you to start up a new disposable droplet.

Note that by default, `up` will use the default values from your configuration and the latest Debian release for your distro. However, you can override this with flags (see `do-disposable up -h` for details).
