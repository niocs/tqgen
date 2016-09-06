# tqgen
A configurable repeatable simulated trade/quote generator for testing databases.

## build
`$ go build tqgen.go`
## usage
(check this [blog post](https://sahas.ra.naman.ms/2016/06/08/tqgen-a-program-for-generating-fake-tradequote-data/) for detailed usage information)
```
$ tqgen --OutFilePat <output_file_pattern> \
        [--NumStk <number_of_randomly_generated_stocks>] \
        [--Seed <randomization_seed>] \
        [--DateBeg <begin_date>] \
        [--DateEnd <end_date>] \
        [--StartTm <sod_time>] \
        [--EndTm <eod_time>]
```
field | description
------|-------------
`<output_file_pattern>  ` | should have YYYYMMDD in file name,  and all occurances of it will be replaced by date,  to produce files date-wise.  Output format is csv.
`<randomization_seed>   ` | output is repeatable as long as we use the same seed.
`<begin_date>,<end_date>` | need to be in YYYYMMDD format. Defaults to 20150101-20150103.
`<sod_time>,<eod_time>  ` | need to be in HHMM format. Defaults to 0930-1600.
##example
```
$ tqgen --OutFilePat /data1/tq/tq.YYYYMMDD.csv \
        --NumStk 5000 \
        --Seed 234532 \
        --DateBeg 20150101 \
        --DateEnd 20160531
```
