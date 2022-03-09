# MTGDB

![Go](https://github.com/pioz/mtgdb/workflows/Go/badge.svg)

MTGDB is a tool written in [GO](https://golang.org) to create and populate a
database with all [Magic The Gathering](https://magic.wizards.com/) cards
available from [Scryfall](https://scryfall.com/). MTGDB also download the
image of each card.

All expansion sets will be saved in the table `sets` with the following
fields:

```
id
name        # Throne of Eldraine
code        # eld
parent_code # eld
released_at # 2019-10-04
typology    # expansion
icon_name   # eld
```

All cards will be saved in the table `cards` with the following fields:

```
id
en_name          # Questing Beast
es_name          # La Bestia Buscada
fr_name          # Bête de Quête
de_name          # Das Questentier
it_name          # Bestia dei Cimenti
pt_name          # Fera das Demandas
ja_name          # 探索する獣
ko_name          # 탐색하는 야수
ru_name          # Заветное Чудище
zhs_name         # 寻水兽
zht_name         # 尋水獸
set_code         # eld
collector_number # 171
foil             # 1
non_foil         # 1
has_back_side    # 0
scryfall_id      # e41cf82d-3213-47ce-a015-6e51a8b07e4f
```

## Install

```
git clone github.com/pioz/mtgdb
cd mtgdb
# go test
go build -o mtgdb ./mtgdb-cli/main.go
./mtgdb -h
```

## Questions or problems?

If you have any issues please add an [issue on
GitHub](https://github.com/pioz/mtgdb/issues) or fork the project and send a
pull request.

## Copyright

Copyright (c) 2020 [Enrico Pilotto (@pioz)](https://github.com/pioz). See
[LICENSE](https://github.com/pioz/mtgdb/blob/master/LICENSE) for details.
