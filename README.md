# Migrate

[![CircleCI branch](https://img.shields.io/circleci/project/github/trivigy/migrate/master.svg?label=master&logo=circleci)](https://circleci.com/gh/trivigy/workflows/migrate)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE.md)
[![](https://godoc.org/github.com/trivigy/migrate?status.svg&style=flat)](http://godoc.org/github.com/trivigy/migrate)
[![GitHub tag (latest SemVer)](https://img.shields.io/github/tag/trivigy/migrate.svg?style=flat&color=e36397&label=release)](https://github.com/trivigy/migrate/releases/latest)

The library was originally forked from [sql-migrate](https://github.com/rubenv/sql-migrate) 
which is a really great project. However I needed a tool that is somewhat more 
golang idiomatic in nature. So I completely re-wrote the API for the forked 
library and optimized on a whole lot of stuff inside. I am using this project to 
back a few other major developments I am working on and will keep updating things 
as needed. If you do find bugs please feel free to submit a pull request.

## Features
* Usable as an embedded CLI tool
* Supports SQLite, PostgreSQL, MySQL, MSSQL (through [gorp](https://github.com/go-gorp/gorp))
* Migrations are defined with SQL for full flexibility
* Transaction based migrations with ability to run transactionless
* Migration rollback support through `up` and `down` commands
* Supports multiple database types in one project

## Installation
To embed the application, use the following from within your project directory:
```bash
go get -u github.com/trivigy/migrate
```

## Usage
The way the library works is purely through embedding it in another project as 
a runnable `cmd`. You can then create multiple of these for different databases 
that your project supports.

For example here is a possible project structure where this tool would be embedded:
```
$ tree ./project
./project
.
├── cmd
│   └── migrate
│       └── main.go
├── go.mod
├── go.sum
├── internal
│   └── migrations
│       ├── 0.0.1_create-users-table.go
│       ├── 0.0.2_create-emails-table.go
│       └── 0.0.3_create-zipcodes-table.go
├── README.md
└── main.go
```

In this case your main project is inside of `main.go` located at the tree root. 
You will then follow by creating the `./cmd/migrate` folder and adding a `main.go` 
file there. Here is an example of what should be added to that file.

> There is absolutely no requirement to call the embedded application `migrate`.
In fact if you are using multiple migration setups, you will have to create a 
few of these and call them differently.

> Skip down here if you just want to see how to write migrations [HERE]()

### `./cmd/migrate/main.go`
```
package main

import (
	"os"

	"github.com/trivigy/migrate"

	_ "github.com/username/project/internal/migrations"
)

func init() {
	migrate.SetConfigs(map[string]migrate.DatabaseConfig{
		"development": {
			Driver: "postgres",
			Source: "host=127.0.0.1 user=postgres dbname=database sslmode=disable",
		},
	})
}

func main() {
	if err := migrate.Execute(); err != nil {
		os.Exit(1)
	}
}

```

> Most important part here is to add the `main()` function with that exact call 
to `migrate.Execute()`. Once you do that, you can run the embedded command to 
help you do the rest.

As you can see, the configuration for the tool are done programmically through 
`migrate.SetConfigs()`. The key of the passed map acts as the environment name. 
You later reference it when calling different commands. The environment names 
can be anything you want. In this case I chose to call it `development`.

Currently `SQLite`, `PostgreSQL`, `MySQL`, `MSSQL` drivers are supported and the 
values of `driver` and `source` are passed varbatum down to `sql.Open(driver, source)`. 
Thus the format for `source` is depended on the type of the database.

Use `--help` to learn about what commands you can run:
```
$ go run ./cmd/migrate --help
Idiomatic GO database migration tool

Usage:
  main [command]

Available Commands:
  create      Create a newly versioned migration template file
  down        Undo the last applied database migration
  status      Show migration status for the current database
  up          Migrates the database to the most recent version

Flags:
  -v, --version   Print version information and quit.
      --help      Show help information.

Use "main [command] --help" for more information about a command.
```

Use the `--help` flag in combination with any of the commands to get an overview of its usage:
```
$ go run ./cmd/migrate up --help
Migrates the database to the most recent version

Usage:
  main up [flags]

Flags:
      --dry-run      Simulate a migration printing planned queries.
  -n, --num NUMBER   Indicate NUMBER of migrations to apply.
  -e, --env ENV      Run with configurations named ENV. (required)
      --help         Show help information.
```

## MySQL Caveat

If you are using MySQL, you must append `?parseTime=true` to the source DSN 
configuration. See [here](https://github.com/go-sql-driver/mysql#parsetime) for 
more information. For example:
### alternative `./cmd/migrate/main.go`
```
package main

import (
	"os"

	"github.com/trivigy/migrate"

	_ "github.com/username/project/internal/migrations"
)

func init() {
	migrate.SetConfigs(map[string]migrate.DatabaseConfig{
		"testing": {
			Driver: "mysql",
			Source: "root@/dbname?parseTime=true",
		},
	})
}

func main() {
	if err := migrate.Execute(); err != nil {
		os.Exit(1)
	}
}

```

## Writing Migrations
Migrations are embedded into the command by referencing them with 
`import _ "github.com/username/project/internal/migrations"`. You might have 
noticed this from the `./cmd/migrate/main.go` examples above. I am chosing to 
place the migration files inside `./internal/migrations` but in face you may 
chose to place them elsewhere.

To help you create migration files quicker, there is the `create` command. What 
is special about it is that it will auto-increment the migration version tags 
which need to be unique.
```
$ go run ./cmd/migrate create --help
Create a newly versioned migration template file

Usage:
  main create NAME[:TAG] [flags]

Flags:
  -d, --dir PATH   Specify directory PATH to create miration file. (default ".")
      --help       Show help information.
```

The tags follow an almost complete semver model. You can use `major`, `minor`, 
`patch`, and `build` parts of the semantic versioning scheme to tag your 
migrations. The migrations get sorted based on this semantic tagging scheme. For 
more detail on the precedence order read [semver](https://semver.org/).

An example migration file might look like this:
```
$ cat ./internal/migrations/0.0.4_create-zipcodes-table.go 
package migrations

import (
        "github.com/trivigy/migrate"
)

func init() {
        migrate.Append(migrate.Migration{
                Tag: "0.0.4",
                Up: []migrate.Operation{
                        {Query: `CREATE TABLE zipcodes (id int)`},
                },
                Down: []migrate.Operation{
                        {Query: `DROP TABLE zipcodes`},
                },
        })
}
```
The filename for the migration files **DO NOT** follow a strict naming convention 
of `{tag}_{filename}.go`. The actual filename is there just to help the developer 
communicate file purpose. However, when using the `create` command filenames are 
generated with that name.
