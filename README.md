# tqgen
A configurable repeatable simulated trade/quote generator for testing databases.

## download & build
`$ go get github.com/niocs/tqgen`
## usage
(check this [blog post](https://sahas.ra.naman.ms/2016/06/08/tqgen-a-program-for-generating-fake-tradequote-data/) for detailed usage information)
```
$ $GOPATH/bin/tqgen --OutFilePat <output_file_pattern> \
                   [--NumStk   <number_of_randomly_generated_stocks>] \
                   [--Seed     <randomization_seed>] \
                   [--Interval <ms_between_two_entries>]
                   [--DateBeg  <begin_date>] \
                   [--DateEnd  <end_date>] \
                   [--StartTm  <sod_time>] \
                   [--EndTm    <eod_time>]
```
field | description
------|-------------
`<output_file_pattern>  ` | Output filename.  If it has the pattern YYYYMMDD, then all occurances of it will be replaced by date,  to produce files date-wise.  All data goes into a single file otherwise.  Output format is csv.
`<randomization_seed>   ` | output is repeatable as long as we use the same seed.
`<ms_between_two_entries>`| Max interval between two consecutive entries, in milliseconds.  Default is 25.  Smaller number means more rows, so larger files.
`<begin_date>,<end_date>` | need to be in YYYYMMDD format. Defaults to 20150101-20150103.
`<sod_time>,<eod_time>  ` | need to be in HHMM format. Defaults to 0930-1600.
##examples
```
$ tqgen --OutFilePat /data1/tq/tq.csv \
        --Seed 234532 \
        --DateBeg 20150101 \
        --DateEnd 20151231
```
```
$ tqgen --OutFilePat /data1/tq/tq.YYYYMMDD.csv \
        --Seed 234532 \
        --DateBeg 20150101 \
        --DateEnd 20151231
```
