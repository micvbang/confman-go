- [About](#about)
- [Installation](#installation)
- [Usage](#usage)
  * [Developer interface](#developer-interface)
    + [Read](#read)
    + [Write](#write)
    + [List](#list)
    + [Delete](#delete)
    + [Environment variables](#environment-variables)
  * [Service interface](#service-interface)

[![Build Status](https://travis-ci.com/micvbang/confman-go.svg?branch=master)](https://travis-ci.com/github/micvbang/confman-go)

# About

Confman is a tool for managing configuration (including secrets) using AWS SSM Parameter Store. It is highly inspired by chamber (https://github.com/segmentio/chamber), with which it is also compatible (see `CONFMAN_CHAMBER_COMPATIBLE` the [environment variables](#environment-variables) section).

Confman differs (mostly in spirit) from chamber in that it's meant to be used for _configuration_, not just for secrets. 
The overall goal of confman is to avoid having multiple versions of configuration lying around on different developer/development/production machines, and instead maintain it in a centralized system with ACLs from which the configuration can be retrieved/modified by relevant consumers (developers/services) whenever needed.


In some environments there's a culture of adding service configuration to the same repository as the source code itself, e.g. using something like `.env[.env]` files. I've found that managing configuration separately from the code being configured has the benefit of making it really easy to deploy services to new environments without any changes to the service being deployed.

# Installation

1. Have a working installation of Go (https://golang.org/doc/install)
2. `go get github.com/micvbang/confman-go/cmd/confman`

# Usage

Confman uses Parameter Store's path model. A service has a "service path" with key-value pairs for configuration. For instance, an email dispatch service could have the path `/email-dispatch/` with key-value pairs `DB_USER=username` and `DB_PASSWORD=password`.

This makes it possible to make namespaces for different stages of the service, allowing us to differentiate configuration between e.g. deployment and runtime. Example:

Let's assume that we use terraform to deploy our email dispatch service to AWS. In this case it would sensible to make the instance type configurable, using a small instance for dev and a larger one for prod. Since it's probably not relevant at runtime of our service to know which type of instance it's running on, we could create separate namespaces for each stage as well as for each environment, e.g. `/email-dispatch/deployment/development` and `/email-dispatch/deployment/production`. On both service paths we could add the key `TF_VAR_ec2_instance_type`.

This could be encoded in yaml like so:

```
/email-dispatch/deployment/development:
  TF_VAR_ec2_instance_type: t3.nano

/email-dispatch/deployment/production:
  TF_VAR_ec2_instance_type: t3.medium
```

At runtime, our email dispatch service needs credentials for a database. We can encode this runtime configuration in yaml like so:

```
/email-dispatch/runtime/development:
  DB_USER: dev-username
  DB_PASSWORD: dev-password

/email-dispatch/runtime/production:
  DB_USER: prod-username
  DB_PASSWORD: prod-password
```

The above example demonstrates a way in which we can separate configuration for different services, for different stages, as well as for different environments. I found that structuring configuration in this way makes it quite maintainable and pretty easy to look up what configuration is currently set for a given service.

## Developer interface

Developers use confman's CLI to manage configuration using the following commands: `read`, `write`, `list`, `delete`, and `exec`.

```
usage: confman [<flags>] <command> [<args> ...]

A tool for easily managing configurations for services

Flags:
  --help                Show context-sensitive help (also try --help-long and
                        --help-man).
  --version             Show application version.
  --debug               Show debugging output
  --aws-kms-key-alias="parameter_store_key"  
                        KMS key alias used for config en/decryption
  --chamber-compatible  Read and write data in a way that is compatible with
                        chamber

Commands:
  help [<command>...]
    Show help.

  read [<flags>] <service> <keys>...
    Reads a configuration

  write [<flags>] <service> <key>
    Writes a configuration

  list [<flags>] <service>
    Lists configuration

  delete [<flags>] <service> [<keys>...]
    Deletes configuration

  exec [<flags>] <service> [<cmd>] [<args>...]
    Populates the environment with secrets from the given configurations
```

### Read

`read` reads keys from a service path.

Example: reading the database password from our email-disapatch service:

```
$ confman read /email-dispatch/runtime/development DB_PASSWORD
/email-dispatch/runtime/development/DB_PASSWORD=secret-password
```

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

### Write

`write` writes a key to a service path.

Example: writing the database password for our email-disapatch service:

```
$ confman write /email-dispatch/runtime/development DB_PASSWORD
Enter value for key 'DB_PASSWORD': secret-password
```


```
usage: confman write [<flags>] <service> <key>

Writes a configuration

Flags:
      --help                Show context-sensitive help (also try --help-long
                            and --help-man).
      --version             Show application version.
      --debug               Show debugging output
      --aws-kms-key-alias="parameter_store_key"  
                            KMS key alias used for config en/decryption
      --chamber-compatible  Read and write data in a way that is compatible with
                            chamber
  -f, --format=txt          Format of output
  -v, --value=VALUE         Value to write (don't use this flag for secret
                            values)
```

### List

`list` lists configuration for a service path

Example: listing the runtime configuration of our email-dispatch service:

```
$ confman list /email-dispatch/runtime/development
Config for '/email-dispatch/runtime/development'
Key                      Value                    version                  last_modified_date       last_modified_user
====                     ======                   ========                 ===================      ===================
DB_PASSWORD              ***                      1                        2020-05-21T12:29:37Z     arn:aws:iam::123456789012:user/confman
DB_USER                  ***                      1                        2020-05-21T12:29:57Z     arn:aws:iam::123456789012:user/confman
```

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

### Delete

`delete` deletes keys from the given service path

Example: Deleting the database user and password from our email-dispatch service:

```
$ confman delete email-dispatch/runtime/development DB_PASSWORD DB_USER                                                          14:34:49
Deleted /email-dispatch/runtime/development/DB_PASSWORD
Deleted /email-dispatch/runtime/development/DB_USER
```

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

### Environment variables

- `CONFMAN_REVEAL_VALUES` `(true,false)`: whether to show values by default when calling `list` or not
- `CONFMAN_DEFAULT_FORMAT` `(txt,json,yaml)`: default format to output in (where relevant)
- `CONFMAN_KMS_KEY_ALIAS` `(string)`: alias of KMS key, e.g. `parameter_store_key`
- `CONFMAN_CHAMBER_COMPATIBLE` `(true,false)`: whether to read/write data in a way that is compatible with chamber


## Service interface

I haven't gotten around to writing a production-ready version of this yet, but the point to create a Go client that fetches configuration directly from AWS Parameter Store and inserts it as environment variables at the beginning of the program. Something along the lines of:

```

import (
  "github.com/micvbang/confman-go/pkg/confman"
  "os"
)

func main() {
  // Retrieves key/value pairs from Parameter Store and sets them in environment
  confman.NewClient().PopulateEnvironment()

  fmt.Println("DB_USER:", os.Getenv("DB_USER"))
}

```
