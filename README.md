# Watchtower
Watchtower handles running processes, watching various targets and running commands when changes to those targets occur.

Here's a basic example of what Watchtower can do:
* On startup:
  - Build a static website
  - Run a backend server to serve the static files
* When watched files change:
  - UI file changes: re-build the static website and then restart the backend process
  - Backend file changes: restart the backend process

As Watchtower supports a variety of watch types and trigger handlers, the interactions can get more complex than this.

## Installation
```bash
    git clone git@github.com:iplay88keys/watchtower
```

## Running
From within the cloned repo:

```bash
    go run cmd/watchtower/main.go --config-file config.yml
```

An example config file can be found in the [example config file](example.yml)

## Config
The config file provides all the information that Watchtower needs in order to run.

The config file is defined as:
```yaml
watches:
  - # ...
# - Required
# - List of watches to set up

processes:
  - # ...
# - Optional
# - List of processes to run
```

### Watches
Watches define what needs to change in order for triggers to run.

They are defined as:
```yaml
name:
# - Required
# - The name that will be used in the output when something being watched changes

config:
# - Required
# - Watch configuration

onTrigger:
  - # ...
# - Required
# - List of triggers that will be run when what is being watched changes
```

#### Watch Configs
##### Path Watcher
The patch watcher defines a set of directories to watch for file changes.

The config is defined as:
```yaml
paths:
  - # ...
# - Required
# - List of root paths to watch for changes from
  
recursive:
# - Default: false
# - Whether to watch for file recursively from each root path
# - File changes in a directory will be watched even if recursive is false if the root is a directory
  
exclusions:
  - # ...
# - Optional
# - List of file regex patterns to ignore changes for

events:
  - # ...
# - Optional
# - List of events to watch for
# - Empty will result in all events being watched
# - Valid options are:
#   - create
#   - write
#   - remove
#   - rename
#   - chmod
```

#### Trigger Configs
##### Run
The run trigger will run a set of commands in order.

The config is defined as:
```yaml
run:
  - # ...
# - Required
# - List of commands to run
# - Supports a limited set of templates
# - Valid templates are:
#   - {{.Name}}
#     - Replaced with the filename that changed 

continueOnError:
# - Default: false
# - Whether to continue running commands if one fails
# - If true, Watchtower will exit on a failed command run
```

##### Restart
The restart trigger will restart a process.

The config is defined as:
```yaml
restart:
# - Required
# - The name of the process to restart (as defined in the processes section)

runCleanup:
# - Default: false
# - Whether to run the cleanup script 
```

### Processes
Defines a list of processes to run on startup and can be restarted using watch triggers.

The config is defined as:
```yaml
name: "frontend"
# - Required
# - The name of the process for output and trigger purposes

type: "task"
# - Required
# - Valid options:
#   - task
#     - The command is run in the foreground and waits for completion
#   - background
#     - The command is run in the background 

start:
# - Required
# - The command to run to start the process
# - Also used when restarting the process

stop:
# - Optional
# - The command to run to stop the process
# - If missing, the process will be killed if it is running
# - Also used when restarting the process

restart:
# - Optional
# - The command used to restart the process
# - Used instead of the stop and start commands when restarting a process if defined

cleanup:
# - Optional
# - The command used to clean up after stopping a process
# - Also used when restarting the process if the restart command is not provided
#   - Runs between the stop and start commands
```
