# Usage

```
usage: confman [<flags>] <command> [<args> ...]

A tool for easily managing configurations for services

Flags:
  --help     Show context-sensitive help (also try --help-long and --help-man).
  --version  Show application version.
  --debug    Show debugging output
  --aws-kms-key-alias="parameter_store_key"
             KMS key alias used for config en/decryption

Commands:
  help [<command>...]
    Show help.

  read [<flags>] <service> <keys>...
    Reads a configuration

  add [<flags>] <service> <key>
    Adds a configuration

  list [<flags>] <service>
    Lists configuration

  delete [<flags>] <service> [<keys>...]
    Deletes configuration
```

## Read

```
usage: confman read [<flags>] <service> <keys>...

Reads a configuration

Flags:
      --help        Show context-sensitive help (also try --help-long and --help-man).
      --version     Show application version.
      --debug       Show debugging output
      --aws-kms-key-alias="parameter_store_key"
                    KMS key alias used for config en/decryption
  -q, --quiet       Print only the value (only works for a single key
  -f, --format=txt  Format of output
```

## Add

```
usage: confman add [<flags>] <service> <key>

Adds a configuration

Flags:
      --help         Show context-sensitive help (also try --help-long and --help-man).
      --version      Show application version.
      --debug        Show debugging output
      --aws-kms-key-alias="parameter_store_key"
                     KMS key alias used for config en/decryption
  -f, --format=txt   Format of output
  -v, --value=VALUE  Value to add (don't use this flag for secret values)
```

## List

```
usage: confman list [<flags>] <service>

Lists configuration

Flags:
      --help        Show context-sensitive help (also try --help-long and --help-man).
      --version     Show application version.
      --debug       Show debugging output
      --aws-kms-key-alias="parameter_store_key"
                    KMS key alias used for config en/decryption
      --reveal      Reveal values
  -f, --format=txt  Format of output
```

## Delete

```
usage: confman delete [<flags>] <service> [<keys>...]

Deletes configuration

Flags:
      --help             Show context-sensitive help (also try --help-long and --help-man).
      --version          Show application version.
      --debug            Show debugging output
      --aws-kms-key-alias="parameter_store_key"
                         KMS key alias used for config en/decryption
      --delete-all-keys  Ignore 'keys' argument and delete all keys for service
  -f, --format=txt       Format of output
```

# Configuration

## Environment variables

- `CONFMAN_REVEAL_VALUES` `(true,false)`: whether to show values when calling `list` or not
- `CONFMAN_DEFAULT_FORMAT` `(txt,json)`: format to output in
- `CONFMAN_KMS_KEY_ALIAS` `(string)`: alias of KMS key, e.g. `parameter_store_key`.
