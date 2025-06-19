# Givilsta - A different whitelisting mechanism for blocklist maintainers <!-- omit in toc -->

Givilsta is a tool designed to help blocklist maintainers or hobbyist manage their lists more effectively by providing an enhanced whitelisting mechanism.
It allows maintainers to not only whitelist by listing each necessary entry but
also through the implementation of custom rules to flexibly manage blocklists.

# Table of Contents <!-- omit in toc -->

- [Background](#background)
- [Installation \& Upgrade](#installation--upgrade)
- [Format Specification](#format-specification)
  - [Introduction](#introduction)
  - [Generic Format](#generic-format)
  - [Commenting](#commenting)
  - [Separators](#separators)
  - [Flags](#flags)
    - [No Flag: The purest form of ruling](#no-flag-the-purest-form-of-ruling)
    - [`ALL`: The "ends-with" rule](#all-the-ends-with-rule)
    - [`REG`: The regular expression rule](#reg-the-regular-expression-rule)
    - [`RZDB`: The broad and powerful rule](#rzdb-the-broad-and-powerful-rule)
- [Usage \& Examples](#usage--examples)
  - [CLI](#cli)
    - [Examples](#examples)
- [LICENSE](#license)

# Background

Givilsta is the golang implementation of the [ultimate-hosts-blacklist-whitelist](https://github.com/ultimate-Hosts-Blacklist/whitelist-tool)
project.

While the original project is production ready, Givilsta is
an attempt to compare the performance of the original project with a Go implementation.
This project does not aim to be compatible with the original project but rather to
provide similar functionality with potentially improved performance.

# Installation & Upgrade

You can install or upgrade Givilsta using the following command:

```bash
go install github.com/funilrys/givilsta@latest
```

# Format Specification

## Introduction

In a world where blocklists and whitelist lists are getting bigger and bigger, the whitelisting mechanism we all use is still the same: we list each entry _(domain, IP, etc.)_ that we want to whitelist and use some kind of GNU or Shell magic spells to proceed with the whitelisting process.

What if we want more ? What if want to whitelist a whole domain and all its subdomains ? What if we want to whitelist with a regular expression ?

This is where Givilsta and [ultimate-hosts-blacklist-whitelist](https://github.com/ultimate-Hosts-Blacklist/whitelist-tool) before it come into play.

With Givilsta, you still can whitelist entries the same way we all do, but you also get some nice features to help you manage your blocklists more effectively by adding prefixing your entries with a flag that Givilsta will understand and use to apply the appropriate whitelisting mechanism.

## Generic Format

Givilsta expects one rule per line in each of the whitelisting files you provide.
Each rule can be prefixed with a flag that indicates the type of whitelisting to apply. The format is as follows:

```
<flag><separator><entry>
```

Where `<flag>` is one of the predefined flags (see below) and `<entry>` is the actual entry to whitelist.

## Commenting

Givilsta supports comments in the whitelisting files. Any line that starts with a `#` character is considered a comment and will be ignored by Givilsta.

If a line contains a comment, the comment will be ignored as well. This means that you can have comments on the same line as a rule, and Givilsta will still process the rule correctly.

For example:

```shell
# This is a comment
ALL example.com # This is a comment but the rule will still be processed.
```

## Separators

The separator is used to distinguish the flag from the entry. Givilsta supports the following separators:

- ` ` _(space)_
- `:` _(colon)_
- `@` _(at)_
- `#` _(hash)_
- `,` _(comma)_

## Flags

### No Flag: The purest form of ruling

This is the purest form of whitelisting. It is what all know and cherish.

For example, if you want to whitelist the domain `example.com`, you can simply list it without any flag:

```text
example.org
```

Therefore, if the source file contains the following entries:

```text
example.com
example.org
```

Givilsta will only whitelist `example.com` as they are.

### `ALL`: The "ends-with" rule

This flag is used to indicate that any entry that ends with the specified entry should be whitelisted.

For example, if you want to whitelist all entries ending with `gov.uk`, you can prefix the entry with the `ALL` flag:

```text
ALL gov.uk
```

### `REG`: The regular expression rule

This flag is used to indicate that the entry is a regular expression and should be treated as such.

For example, if you want to whitelist all entries that match the regular expression `^example\.(com|org)$`, you can prefix the entry with the `REG` flag:

```text
REG ^example\.(com|org)$
```

### `RZDB`: The broad and powerful rule

**_Alias:_** `RZD`

This flag is used to indicate that the entry should match any RZDB (Root Zone Database) entry. This is useful for whitelisting for example any possible occurence of a word,
company or brand that might be used in a domain name without having to think about the TLD (Top Level Domain).

For example, if you want to whitelist any entry that matches `example.[gTLD]`, you can prefix the entry with the `RZDB` flag:

```text
RZDB example
```

**Beware, this flag is extremely
broad and powerful as it will fetch the
[IANA Root Zone Database](https://www.iana.org/domains/root/db) and the
[Public Suffix List](https://publicsuffix.org/)
to build a set of rules with all possible gTLDs or extensions.**


# Usage & Examples

## CLI

To use Givilsta, you can run the command line interface (CLI) with various flags to customize its behavior. Please refer to the help command for more details on the available options:

```bash
Usage:
  givilsta [flags]
  givilsta [command]

Available Commands:
  completion  Generate the autocompletion script for the specified shell
  help        Help about any command
  version     Print the version number of your application

Flags:
  -c, --handle-complement         Whether to handle complements subjects or not.
                                        A complement subject is www.example.com when the subject is example.com - and vice-versa.
                                        This is useful for domains that have a 'www' subdomain and want them to be whitelisted when the domain
                                        (without 'wwww' prefix) is whitelist listed.
  -h, --help                      help for givilsta
  -l, --log-level string          The log level to use. Can be one of: debug, info, warn, error. (default "error")
  -o, --output string             The output file to write the cleaned up subjects to. If not specified, we will print to stdout.
  -s, --source string             The source file to cleanup.
  -w, --whitelist strings         The whitelist file to use for the cleanup. Can be specified multiple times.
  -a, --whitelist-all strings     The whitelist file to use for the cleanup. Any entries in this file-s will be prefixed with the 'ALL' flag. Can be specified multiple times.
  -r, --whitelist-regex strings   The whitelist file to use for the cleanup. Any entries in this file-s will be prefixed with the 'REG' flag. Can be specified multiple times.
  -z, --whitelist-rzdb strings    The whitelist file to use for the cleanup. Any entries in this file-s will be prefixed with the 'RZDB' flag. Can be specified multiple times.

Use "givilsta [command] --help" for more information about a command.
```

### Examples

```shell
# content of test.list
example.com
example.org
api.example.org
test.example.com
```

```shell
# content of whitelist.list
api.example.org
ALL .com
```

```shell
$ givilsta -s test.list -w whitelist.list
example.org
```



# LICENSE

```
Copyright (c) 2025 Nissar Chababy

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
```