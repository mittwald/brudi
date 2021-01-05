# brudi

When it comes to backup-creation there are several solutions to use.  
In general everybody's doing some sort of `dump` or `tar` and backing up the results incremental with [`restic`](https://github.com/restic/restic) or similar programs.

This is why `brudi` was born. `brudi` supports several backup-methods and is configurable by a simple `yaml` file.
The advantage of `brudi` is, that you can create a backup of a source of your choice and save it with `restic` afterwards in one step.
Under the hood, `brudi` uses the given binaries like `mysqldump`, `mongodump`, `pg_dump`, `tar` or `restic`.

Using `brudi` will save you from finding yourself writing bash-scripts to create your backups.

Besides creating backups, `brudi` can also be used to restore your data from backup in an emergency.

## Table of contents

- [Usage](#usage)
  - [CLI](#cli)
  - [Configuration](#configuration)
    - [Sources](#sources)
      - [Tar](#tar)
      - [MySQLDump](#mysqldump)
      - [MongoDump](#mongodump)
      - [PgDump](#pgdump)
        - [Limitations](#limitations)
      - [Redis](#redis)
    - [Restic](#restic)
      - [Forget](#forget)
    - [Sensitive data: Environment variables](#sensitive-data--environment-variables)
    - [Restoring from backup](#restoring-from-backup)
      - [TarRestore](#tarrestore)
      - [MongoRestore](#mongorestore)
      - [MySQLRestore](#mysqlrestore)
      - [PgRestore](#pgrestore)
        - [Restore using pg_restore](#restore-using-pg_restore)
        - [Restore using psql](#restore-using-psql)
- [Featurestate](#featurestate)
  - [Source backup methods](#source-backup-methods)
  - [Restore backup methods](#restore-backup-methods)
  - [Incremental backup of the source backups](#incremental-backup-of-the-source-backups)

## Usage

### CLI

In order to use the `brudi`-binary on your local machine or a remote server of your choice, ensure you have the required tools installed.

- `mongodump` (required when running `brudi mongodump`)
- `mysqldump` (required when running `brudi mysqldump`)
- `tar` (required when running `brudi tar`)
- `redis-cli` (required when running `brudi redisdump`)
- `restic` (required when running `brudi --restic`)


```shell
$ brudi --help

Easy, incremental and encrypted backup creation for different backends (file, mongoDB, mysql, etc.)
After creating your desired tar- or dump-file, brudi backs up the result with restic - if you want to

Usage:
  brudi [command]

Available Commands:
  help           Help about any command
  mongodump      Creates a mongodump of your desired server
  mongorestore   Restores a server from a mongodump
  mysqldump      Creates a mysqldump of your desired server
  mysqlrestore   Restores a database from an sqldump
  pgdump         Creates a pg_dump of your desired postgresql-server
  pgrestore      Restores a database from a pgdump using pg_restore
  psql           Restores a database from a plain-text pgdump using psql
  redisdump      Creates an rdb dump of your desired server
  tar            Creates a tar archive of your desired 
  tarrestore     Restores files from a tar archive
  version        Print the version number of brudi

Flags:
      --cleanup         cleanup backup files afterwards
  -c, --config string   config file (default is ${HOME}/.brudi.yaml)
  -h, --help            help for brudi
      --restic          backup result with 'restic backup'
      --restic-forget   executes 'restic forget' after backing up things with restic
      --version         version for brudi

Use "brudi [command] --help" for more information about a command.
```

### Docker

In case you don't want to install additional tools, you can also use `brudi` inside docker:

`docker run --rm -v ${HOME}/.brudi.yml:/home/brudi/.brudi.yml quay.io/mittwald/brudi mongodump --restic --cleanup`

The docker-image comes with all required binaries.

### Configuration

As already mentioned, `brudi` is configured via `.yaml`. The default path for this file is `${HOME}/.brudi.yaml`, but it's adjustable via `-c` or `--config`.
The config file itself can include environment-variables via `go-template`:

```yaml
restic:
  global:
    flags:
      repo: "{{ .Env.RESTIC_REPOSITORY }}"
```

Since the configuration provided by the `.yaml`-file is mapped to the corresponding CLI-flags, you can adjust literally every parameter of your source backup.  
Therefore you can simply refer to the official documentation for explanations on the available flags:

- [`restic`](https://restic.readthedocs.io/en/latest/manual_rest.html)
- [`tar`](https://www.gnu.org/software/tar/manual/html_section/tar_22.html)
- [`mongodump`](https://docs.mongodb.com/manual/reference/program/mongodump/#options)
- [`mysqldump`](https://dev.mysql.com/doc/refman/8.0/en/mysqldump.html#mysqldump-option-summary)
- [`pg_dump`](https://www.postgresql.org/docs/12/app-pgdump.html)

Every source has a an `additionalArgs`-key which's value is an array of strings. The value of this key is appended to the command, generated by `brudi`.
Even though `brudi` should support all cli-flags to be configured via the `.yaml`-file, there may be flags which are not.  
In this case, use the `additionalArgs`-key.

It is also possible to provide more than one configuration file, for example `-c mongodump.yaml -c restic.yaml`. These configs get merged at runtime.
If available, the default config will always be laoded first and then overwritten with any values from user-specified files. 
In case the same config file has been provided more than once, only the first instance will be taken into account.

#### Sources

##### Tar

```yaml
tar:
  options:
    flags:
      create: true
      gzip: true
      file: /tmp/test.tar.gz
    additionalArgs: []
    paths:
      - /tmp/testfile
  hostName: autoGeneratedIfEmpty
```

Running: `brudi tar -c ${HOME}/.brudi.yml --cleanup`

Becomes the following command:  
`tar -c -z -f /tmp/test.tar.gz /tmp/testfile`  

All available flags to be set in the `.yaml`-configuration can be found [here](pkg/source/tar/cli.go#L7).

##### MySQLDump

```yaml
mysqldump:
  options:
    flags:
      host: 127.0.0.1
      port: 3306
      password: mysqlroot
      user: root
      opt: true
      allDatabases: true
      resultFile: /tmp/test.sqldump
    additionalArgs: []
```

Running: `brudi mysqldump -c ${HOME}/.brudi.yml --cleanup`

Becomes the following command:  
`mysqldump --all-databases --host=127.0.0.1 --opt --password=mysqlroot --port=3306 --result-file=/tmp/test.sqldump --user=root`  

All available flags to be set in the `.yaml`-configuration can be found [here](pkg/source/mysqldump/cli.go#L7).

##### MongoDump

```yaml
mongodump:
  options:
    flags:
      host: 127.0.0.1
      port: 27017
      username: root
      password: mongodbroot
      gzip: true
      archive: /tmp/dump.tar.gz
    additionalArgs: []
```

Running: `brudi mongodump -c ${HOME}/.brudi.yml --cleanup`

Becomes the following command:  
`mongodump --host=127.0.0.1 --port=27017 --username=root --password=mongodbroot --gzip --archive=/tmp/dump.tar.gz`  

All available flags to be set in the `.yaml`-configuration can be found [here](pkg/source/mongodump/cli.go#L7).

##### PgDump

```yaml
pgdump:
  options:
    flags:
      host: 127.0.0.1
      port: 5432
      password: postgresroot
      username: postgresuser
      dbName: postgres
      file: /tmp/postgres.dump
    additionalArgs: []
```

Running: `brudi pgdump -c ${HOME}/.brudi.yml --cleanup`

Becomes the following command:  
`pg_dump --file=/tmp/postgres.dump --dbname=postgres --host=127.0.0.1 --port=5432 --username=postgresuser`  

All available flags to be set in the `.yaml`-configuration can be found [here](pkg/source/pgdump/cli.go#L7).

###### Limitations

Unfortunately `PostgreSQL` is very strict when it comes to version-compatibility.  
Therefore your `pg_dump`-binary requires the exact same version your `PostgreSQL`-server is running.

The Docker-image of `brudi` always has the latest version available for the corresponding alpine-version installed.

##### Redis

```yaml
redisdump:
  options:
    flags:
      host: 127.0.0.1
      password: redisdb
      rdb: /tmp/redisdump.rdb
    additionalArgs: []
```

Running: `brudi redisdump -c ${HOME}/.brudi.yml`

Becomes the following command:
`redis-cli -h 127.0.0.1 -a redisdb --rdb /tmp/redisdump.rdb bgsave`

As `redis-cli` is not a dedicated backup tool but a client for `redis`, only a limited number of flags are available by default,
as you can see [here](pkg/source/redisdump/cli.go#L7).

#### Restic

In case you're running your backup with the `--restic`-flag, you need to provide a [valid configuration for restic](https://restic.readthedocs.io/en/latest/030_preparing_a_new_repo.html).  
You can either configure `restic` via `brudi`s `.yaml`-configuration, or via the [environment variables](https://restic.readthedocs.io/en/latest/040_backup.html#environment-variables) used by `restic`.  

If you're already using `restic` in your environment, you should have everything set up perfectly to use `brudi` with `--restic`.

##### Forget

It's also possible to run `restic forget`-cmd after executing `restic backup` with `brudi` by using `--restic-forget`.  
The `forget`-policy is defined in the configuration `.yaml` for brudi.

Example `.yaml`-configuration:

```yaml
restic:
    global:
      flags:
        # you can provide the repository also via RESTIC_REPOSITORY
        repo: "s3:s3.eu-central-1.amazonaws.com/your.s3.bucket/myResticRepo"
    backup:
      flags:
        # in case there is no hostname given, the hostname from source backup is used
        hostname: "MyHost"
      # these paths are backuped additionally to your given source backup
      paths: []
  forget:
    flags:
      keepLast: 48
      keepHourly: 24
      keepDaily: 7
      keepWeekly: 2
      keepMonthly: 6
      keepYearly: 2
    ids: []
```

#### Sensitive data: Environment variables

In case you don't want to provide data directly in the `.yaml`-file, e.g. sensitive data like passwords, you can use environment-variables.
Each key of the configuration is overwritable via environment-variables. Your variable must specify the whole path to a key, seperated by `_`.  
For example, given this `.yaml`:

```yaml
mongodump:
  options:
    flags:
      username: "" # we will override this by env
      password: "" # we will override this by env
      host: 127.0.0.1
      port: 27017
      gzip: true
      archive: /tmp/dump.tar.gz
```

Set your env's:

```shell
export MONGODUMP_OPTIONS_FLAGS_USERNAME="root"
export MONGODUMP_OPTIONS_FLAGS_PASSWORD="mongodbroot"
```

As soon as a variable for a key exists in your environment, the value of this environment-variable is used in favour of your `.yaml`-config.

#### Restoring from backup

#### TarRestore

```yaml
tarrestore:
  options:
    flags:
      extract: true
      gzip: true
      file: /tmp/test.tar.gz
      target: "/"
    additionalArgs: []
  hostName: autoGeneratedIfEmpty
```

Running: `brudi tarrestore -c ${HOME}/.brudi.yml`

Becomes the following command:
`tar -x -z -f /tmp/test.tar.gz -C /`   

##### MongoRestore

 ```yaml
 mongorestore:
   options:
     flags:
       host: 127.0.0.1
       port: 27017
       username: root
       password: mongodbroot
       gzip: true
       archive: /tmp/dump.tar.gz
     additionalArgs: []
 ```
 
 Running: `brudi mongorestore -c ${HOME}/.brudi.yml `
 
 Becomes the following command:  
 `mongorestore --host=127.0.0.1 --port=27017 --username=root --password=mongodbroot --gzip --archive=/tmp/dump.tar.gz`  
 
 All available flags to be set in the `.yaml`-configuration can be found [here](pkg/source/mongorestore/cli.go#L7).

#### MySQLRestore

```yaml
mysqlrestore:
  options:
    flags:
      host: 127.0.0.1
      port: 3306
      password: mysqlroot
      user: root
      Database: test
    additionalArgs: []
    sourceFile: /tmp/test.sqldump
```

Running: `brudi mysqlrestore -c ${HOME}/.brudi.yml`

Becomes the following command:  
`mysql --database=test --host=127.0.0.1 --password=mysqlroot --port=3306 --user=root < /tmp/test.sqldump`  

All available flags to be set in the `.yaml`-configuration can be found [here](pkg/source/mysqlrestore/cli.go#L7).

#### PgRestore

Restoration for PostgreSQL databases is split into two commands, `psql` and `pgrestore`. Which one to use depends on the format of the dump created with `pg_dump`:

`psql` can be used to restore plain-text dumps, which is the default format.

`pgrestore` can be used if the `format` option of `pg_dump` was set to `tar`, `directory` or `custom`.

##### Restore using pg_restore

```yaml
pgrestore:
  options:
    flags:
      host: 127.0.0.1
      port: 5432
      username: postgresuser
      password: postgresroot
      dbname: postgres
    additionalArgs: []
    sourcefile: /tmp/postgres.dump
```

Running: `brudi pgrestore -c ${HOME}/.brudi.yml`

Becomes the following command:  
`pg_restore  --host=127.0.0.1  --port=5432 --username=postgresuser --db-name=postgres /tmp/postgress.dump`  

This command has to be used if the `format` option was set to `tar`, `directory` or `custom` in `pg_dump`.

All available flags to be set in the `.yaml`-configuration can be found [here](pkg/source/pgrestore/cli.go#L7).

##### Restore using psql

```yaml
psql:
  options:
    flags:
      host: 127.0.0.1
      port: 5432
      user: postgresuser
      password: postgresroot
      dbname: postgres
    additionalArgs: []
    sourcefile: /tmp/postgres.dump
```

Running: `brudi pgrestore -c ${HOME}/.brudi.yml`

Becomes the following command: 
`psql  --host=127.0.0.1  --port=5432 --user=postgresuser --db-name=postgres < /tmp/postgress.dump`  

This command has to be used if the `format` option was set to `plain` in `pg_dump`, which is the default.

All available flags to be set in the `.yaml`-configuration can be found [here](pkg/source/psql/cli.go#L7).

#### Restoring using restic

Backups can be pulled from a `restic` repository and applied to your server by using the `--restic` flag in your brudi command. 
Example configuration for `mongorestore`:
```yaml
mongorestore:
  options:
    flags:
      host: 127.0.0.1
      port: 27017
      username: root
      password: mongodbroot
      gzip: true
      archive: /tmp/dump.tar.gz
    additionalArgs: []
restic:
  global:
    flags:
      repo: "s3:s3.eu-central-1.amazonaws.com/your.s3.bucket/myResticRepo"
  restore:
    flags:
      target: "/"
    id: "latest"
```

This will pull the latest snapshot of `/tmp/dump.tar.gz` from the repository, which `mongorestore` then uses to restore the server.
It is also possible to specify concrete snapshot-ids instead of `latest`.      

## Featurestate

### Source backup methods

- [x] `mysqldump`
- [x] `mongodump`
- [x] `tar`
- [x] `pg_dump`
- [x] `redisdump`

### Restore backup methods

- [x] `mysqlrestore` 
- [x] `mongorestore`
- [x] `tarrestore`
- [x] `pgrestore`
- [ ]  `redisrestore`
 
### Incremental backup of the source backups

- [x] `restic`
  - [x] `commands`
    - [x] `restic backup`
    - [x] `restic forget`
    - [x] `restic restore`
  - [x] `storage`
    - [x] `s3`
