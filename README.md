# `flaker`
Flaker is a small CLI that lets you run a given command repeatedly in parallel. This is primarily designed around
running tests repeatedly to detect flaky tests.

## Usage
```text
NAME:
   flaker - Repeatedly run a command (usually a test) to check for flakiness

USAGE:
   flaker [options] -- [command]

VERSION:
   0.0.2

COMMANDS:
   help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   -p value, --parallel value     Number of concurrent runs of the command (default: 1)
   -d value, --duration value     How long to run the commands for. This is the minimum run time as this will wait until currently running commands finish before returning (default: 0s)
   -r value, --refresh value      How frequently to refresh the output. This takes a number followed by a unit such as '1m', '30s', '1h' (default: 1s)
   -o value, --output_file value  File to write results to in JSON format
   --root_command value           Specifies the underlying root command to use to enable piping and redirection (default: "bash", "-c")
   --help, -h                     show help (default: false)
   --version, -v                  print the version (default: false)
```

![flaker](flaker.gif)
