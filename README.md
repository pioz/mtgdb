# MTGDB

![Go](https://github.com/pioz/mtgdb/workflows/Go/badge.svg)

MTGDB is a tool written in [GO](https://golang.org) to create and populate a
database with all [Magic The Gathering](https://magic.wizards.com/) cards
available from [Scryfall](https://scryfall.com/). MTGDB also download the
image of each card.

## Install

```
go install github.com/pioz/mtgdb/mtgdb-cli@latest
mtgdb-cli -h
```

### From sources

```
git clone github.com/pioz/mtgdb
cd mtgdb
# go test
go build -o mtgdb ./mtgdb-cli/main.go
./mtgdb -h
```

## Usage

Before using MTGDB you have to set 2 environment variables (also `.env` file works):

- `DB_CONNECTION` -> database connection string (example `user@tcp(127.0.0.1:3306)/mtgdb?charset=utf8mb4&parseTime=True`)
- `DATA_PATH` -> path where download assets like card images (example `./data`)

The first time you run MTGDB, it will migrate also the database creating the tables.

```
mtgdb-cli -h
Usage of mtgdb-cli:
  -download-concurrency int
    	Set max download concurrency
  -en
    	Download card images only in EN language (default true)
  -f	Force re-download of card images
  -fsha1
    	Force re-download of card images, but only if the sha1sum is changed
  -ftime
    	Force re-download of card images, but only if the modified date is older
  -h	Print this help
  -only string
    	Import some sets (es: -only eld,war)
  -p	Display progress bar
  -skip-assets
    	Skip download of set and card images
  -u	Update Scryfall database
```

## Questions or problems?

If you have any issues please add an [issue on
GitHub](https://github.com/pioz/mtgdb/issues) or fork the project and send a
pull request.

## Copyright

Copyright (c) 2020 [Enrico Pilotto (@pioz)](https://github.com/pioz). See
[LICENSE](https://github.com/pioz/mtgdb/blob/master/LICENSE) for details.
