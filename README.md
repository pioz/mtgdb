# MTGDB

MTGDB is a tool written in [GO](https://golang.org) to create and populare a
database with all [Magic The Gathering](https://magic.wizards.com/) cards
available from [Scryfall](https://scryfall.com/). MTGDB also download the
image of each card.

All cards will be saved on a single table `cards` with the following fields:

```
en_name
es_name
fr_name
de_name
it_name
pt_name
ja_name
ko_name
ru_name
zhs_name
zht_name
set_code
collector_number
is_token
icon_name
scryfall_id
```

## Install

```
git clone github.com/pioz/mtgdb
cd mtgdb
go build -o mtgdb main.go
./mtgdb
```

## Questions or problems?

If you have any issues please add an [issue on
GitHub](https://github.com/pioz/mtgdb/issues) or fork the project and send a
pull request.

## Copyright

Copyright (c) 2020 [Enrico Pilotto (@pioz)](https://github.com/pioz). See
[LICENSE](https://github.com/pioz/mtgdb/blob/master/LICENSE) for details.
