[![Go Reference](https://pkg.go.dev/badge/github.com/zafnz/configape.svg)](https://pkg.go.dev/github.com/zafnz/configape)
[![Go Report Card](https://goreportcard.com/badge/github.com/zafnz/configape)](https://goreportcard.com/report/github.com/zafnz/configape)

# Config Ape

A go library to make configuration and commandline arguments trivial and convenient and easy to use. You specify your application configuration using a standard golang struct, and struct tags (like json) to specify options for those fields.

Example:
```go
var config = struct {
    Name string `help:"Your name" required:"true"`
    Verbosity int `name:"verbose" help:"Verbosity level" default:"0" cfgtype:"counter"`
    Database struct {
        Host string `help:"Database host"`
        Port int `help:"Database port"`
    }
    // This allows the user to override which config file to read from. By default it is config.json
    ConfigFile string `name:"config" help:"Config file to read from" cfgtype:"configfile"`
    // Any additional arguments on the command line are put into this slice.
    Files []string `cfgtype:"remaining" help:"Files to process"`
}{}
func main() {
    // Checks for config.json, an environment variable named CFG_NAME, and a commandline argument .named --name
    // If you run this as my-program --name=Bob
    configape.Apply(&config, configape.Options{ConfigFile: "config.json"})
    // Output:
    // Hello Bob
    fmt.Printf("Hello %s\n", config.Name)
}
```
If you add `--help`, you will automatically get the cli help:
```
example (v0.0.0)

  --simple-string <SimpleString>
    This is the help for simple string
  --simple-int <SimpleInt> (default: 42)
  --required-string <RequiredString>
  --something <something>
  --short, -s
    This is the help for the short flag
  --config <ConfigFile>
    Load config from this file

database

  --database-host <Host>
    Database host to connect to
  --database-port <Port>
    Database port to use with the database host
```

## Options
The `Apply` function takes a pointer to a struct, and an optional options struct. The options struct can contain the following fields (full documentation is in the [godoc](https://godoc.org/github.com/zafnz/configape)):
| Field | Description |
| --- | --- |
| `ConfigFile` | The name of the config file to read from, if not specified it defaults to `config.json` |
| `ConfigFileType` | The type of the config file, if not specified it is automatically detected from the file extension |
| `EnvironmentPrefix` | The prefix to use for environment variables, if not specified it defaults to `CFG_` |
| `Help` | A function to call to display help text, if not specified it defaults to `configape.Help` |
| `HelpHeader` | Help text that is prefixed to the help output. |
| `HelpFooter` | Help text that is appended to the help output. |
| `Name` | The name of the application, if not specified it defaults to the name of the executable |
| `Version` | The version of the application, if not specified it defaults to `0.0.0` |
| `Writer` | The writer to use for output, if not specified it defaults to `os.Stderr` |
| `DisableEnviroment` | If set to true, then environment variables are not used to set config variables |
| `DisableConfigFile` | If set to true, then config files are not used to set config variables |
| `DisableCommandLine` | If set to true, then command line arguments are not used to set config variables |
| `DisableHelp` | If set to true, then the help text is not displayed to the user |
| `DisableVersion` | If set to true, then the version text is not displayed to the user |
| `DisableHelpOnMissingRequired` | If set to true, then the help text is not displayed to the user if a required config variable is missing |
| `AllowUnknownConfigFileKeys` | If set to true, then unknown keys in the config file are ignored, otherwise an error is returned |


## Tags
By default all exported struct fields are used and available to be set by the user, the name is automatically calculated by hyphenating the camel case and using lowercase letters (eg `HomeDirectory` becomes `home-directory`), you can change this with the `name` tag (see below). 

| Tag | Description |
| --- | --- |
| `name` | The name of the config variable, if not specified the name is calculated from the struct field name, hyphenating the camel case and using lowercase letters (eg `HomeDirectory` becomes `home-directory`) |
| `help` | The help text to display to the user |
| `required` | If the variable is required, if not set an error is returned |
| `default` | The default value for the variable |
| `cfgtype` | The type of the variable, see below for more information |
| `cli` | Override the cli argument name, by default it is the `name` value (see defaults for it), set to "-" to disable setting this field via the cli |
| `env` | The name of the environment variable to use, if not specified the name is calculated by uppercasing the name tag and prepending `CFG_`. Set to `-` to disable this config field being set in the environment |

## Special fields
There are a few special fields specified with `cfgtype` that can be used in your config struct:
| cfgtype | Description |
| --- | --- |
| `configfile` | The user can specify a config file to read from, see below for more information |
| `counter` | The user can specify the variable multiple times on the commandline, and the value is incremented each time. Otherwise in config files and environment it can just be set to a number. (You can also specify it's exact value on the commandline by using the `--verbose=9` format) |

## Booleans/flags
If a struct field is of a boolean type, then it is a flag, and specifying `--field-name` will set it to true. You can also specify `--field-name=false` to set it to false. If you want to specify a default value, you can use the `default` tag, eg `default:"true"`.

## Config file
Config Ape by default looks for a file called `config.json` in the current working directory, but you can provide a different file name with the options argument to `Apply`. The config file can be yaml, json, or toml. The config file is loaded first, and then the environment, followed by the command line arguments. The command line arguments override the environment, and the environment overrides the config file.

If you provide a variable in your config struct that has the `type:"configfile"` tag, then the user can supply a config file to read from
rather than the default. For example:
```go
var config = struct {
    ConfigFile string `cfg:"config",help:"Config file to read from",type:"configfile"`
}{}
```

## Environment
By default all config variables are settable by enviroment variables prefixed with "CFG_" (to avoid name collisions with other environment
variables). For example, the config variable `Name` can be set by the environment variable `CFG_NAME`. You can change the prefix by setting the `EnvPrefix` field in the options argument to `Apply`. You can disable a prefix by setting the `EnvPrefix`` to `!` (exclamation mark), which is not recommended.

## Command line arguments
By default all config variables are settable by command line arguments prefixed with "--" (double dash). For example, the config variable `YourHouse` can be set by the command line argument `--your-house`. The library is smart enough to handle many variations. Eg `--your-house`, `--your_house` or `--YourHouse`.
You can disable a config from being set via the cli by setting the `cli` tag to `-`.

## Sections
Config Ape can handle structs within structs, and will automatically create sections for them. For example:
```go
var config = struct {
    Name string `help:"Your name",required:"true"`
    Verbosity int `name:"verbose",help:"Verbosity level",default:"0",cfgtype:"counter"`
    Database struct {
        Host string `help:"Database host"`
        Port int `help:"Database port"`
    }
}{}
```
In this example, you can set the database host by setting the environment variable `CFG_DATABASE_HOST` or the command line argument `--database-host`. You can also set the database port by setting the environment variable `CFG_DATABASE_PORT` or the command line argument `--database-port`.

## Help
ConfigApe automatically generates help text for you, and displays it if the user specifies `--help` on the command line. If you wish to handle this yourself, then add a field called `Help` of type boolean, then check for that being true after calling `Apply`. You can also use the `Help` function to output the default help text. For example:
```go
var config = struct {
    Name string `help:"Your name" required:"true"`
    Help bool `cfgtype:"help" help:"Show help"`
}{}
cfgape.Apply(&config, nil)
if config.Help {
    fmt.Print(configape.Help(&config))
    os.Exit(0)
}
```