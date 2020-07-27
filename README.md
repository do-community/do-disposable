# do-disposable
Allows you to create a disposable Droplet with a simple CLI interface. This is useful for trying to run actions which may be network intensive to do on your standard internet connection or too CPU intensive for your system since you can start a Droplet to the size you need and then have it easily disposed of when you're done.

The following sub-commands are included:
- `auth`: Authenticates the user and creates the configuration if this doesn't exist. This is the sub-command you will be prompted to run on first launch of this tool.
- `setregion`: Allows you to modify the region. Note that you need to go through the setup with do-disposable auth first (that will also configure this for the first time).
- `setsize`: Allows you to modify the Droplet size. Note that you need to go through the setup with do-disposable auth first (that will also configure this for the first time).
- `up`: Allows you to start up a new disposable Droplet.

Additionally, when deploying the Droplet, the following commands are deployed to the Droplet:
- `copyfrom <host file/folder path> [droplet save location]`: Allows you to copy a file/folder from the host to the Droplet.
- `copyback <droplet file/folder path> [host save location]`: Allows you to copy a file/folder back from the Droplet.

## First Usage
You can find a binary for your operating system in the [releases page](https://github.com/JakeMakesStuff/do-disposable/releases).

On first boot, you should be logged in to DigitalOcean and then run `do-disposable auth`. On the first boot, this will then walk through through the [authentication](#authentication), [set region](#set-region) and [set size](#set-size) configurations.

From here, you can simply start the droplet with `do-disposable up` (see [starting the droplet](#starting-the-droplet)) for help with this.

## Authentication
To change your SSH key/token after the first boot, you can run `do-disposable auth`. When you run this, you will be prompted for your DigitalOcean token:

![token_prompt](https://i.imgur.com/lEnuaSL.png)

You can get this from the API section of the DigitalOcean dashboard:
![api](https://i.imgur.com/naOZtuJ.png) ![token](https://i.imgur.com/VRWmHsH.png)

This will also automatically generate/save the SSH key which do-disposable will use internally.

## Set Region
To set the region outside of the first boot, you can use `do-disposable setregion`. This will prompt you to hit the key of the region which you wish to use:

![setregion](https://i.imgur.com/5t28FT0.png)

## Set Size
To set the size outside of the first boot, you can use `do-disposable setsize`. This will prompt you to hit the key of the size which you wish to use:

![setsize](https://i.imgur.com/Ao8mXFu.png)

## Starting The Droplet
To start the Droplet, you can use `do-disposable up`. Note that by default, `up` will use the default values from your configuration and the latest Debian release for your distro. The following flags can be used:
- `distro`: Allows you to override the distro slug with another one from the DigitalOcean API. This defaults to the latest Debian release (e.g.: `-distro ubuntu-19-10-x64`).
- `region`: Allows you to override the region slug with another one. Will default to the one set above (e.g.: `-region nyc3`).
- `size`: Allows you to override the size slug with another one. Will default to the one set above (e.g.: `-size s-1vcpu-2gb`).

To get slugs for different Droplet attributes, you can use [this tool](https://slugs.do-api.dev/).

From here, you can run Linux commands (including `copyfrom` and `copyback`) and then you can exit the Droplet. Exiting will destroy the Droplet:

![session](https://i.imgur.com/UXxEv3w.png)
