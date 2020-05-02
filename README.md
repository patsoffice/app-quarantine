# app-quarantine

app-quarantine is a CLI tool for identifying and potentially fixing quarantined applications on macOS.

## Installing

### Directly From GitHub

Installing app-quarantine is easy. First, use `go get` to install the latest version of the application
from GitHub. This command will install the `app-quarantine` executable in the Go path:

    go get -u github.com/patsoffice/app-quarantine

### Homebrew

    brew tap patsoffice/tools
    brew install app-quarantine

## Running

Checking:

    $ app-quarantine
    /Applications/Slack.app is quarantined
    /Applications/calibre.app is quarantined

Repairing (add the `-f` or `--fix` flag):

    $ app-quarantine -f
    /Applications/Slack.app is quarantined... Removed quarantine.
    /Applications/calibre.app is quarantined... Removed quarantine.
