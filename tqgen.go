package main

import (
	"github.com/niocs/sflag"
	"github.com/niocs/nrand"
	"math/rand"
	"strings"
	"time"
	"log"
	"fmt"
	"os"
)

type Stock struct {
	Name         string
	Liquidity    float64
	BasePx       float64
	LastArrTm    time.Time       // Last Arrival Time
	LastType     string
	LastBidPx    float64
	LastBidSz    int64
	LastAskPx    float64
	LastAskSz    int64
	LastQuoteTm  time.Time
	LastTrdPx    float64
	LastTrdSz    int64
	LastTrdTm    time.Time
	Started      bool           // has first quote been generated?
}

type Exch struct {
	DateBeg          time.Time
	DateEnd          time.Time
	DateNow          time.Time
	DateTimeNow      time.Time
	DateTimeEOD      time.Time
	SOD              time.Duration  // time duration from midnight to SOD
	EOD              time.Duration  // time duration from midnight to EOD
	Stocks           []*Stock
	TotalLiquidity   float64
}

var opt = struct {
	Usage       string  "Prints usage string"
	NumStk      int     "Number of stocks | 100"
	Seed        int64   "Randomization seed. Default is 1 | 1"
	Interval    int     "Max milliseconds between two entries.  Decreasing this increases number of lines per date. Default is 25. | 25"
	DateBeg     string  "Date in YYYYMMDD format | 20150101"
	DateEnd     string  "Date in YYYYMMDD format | 20150103"
	StartTm     string  "Trading start time | 0930"
	EndTm       string  "Trading end time | 1600"
	OutFilePat  string  "outfile name pattern with YYYYMMDD that will be replaced by date"
}{}

func GenName() string {
	var name []byte
	len := 3
	for ii := 0; ii < len; ii ++ {
		name = append(name, byte(rand.Int31n(26))+65)
	}
	return string(name)
}

func GenNames(n int) []string {
	// use a map to keep the list unique
	
	nameMap  := make(map[string]struct{})
	var names []string
	for len(nameMap) < opt.NumStk {
		name := GenName()
		if _, ok := nameMap[name]; !ok {
			nameMap[name] = struct{}{}
			names = append(names, name)
		}
	}
	return names
}

func GenStocks(names []string, seed int64) ([]*Stock, float64) {
	stocks := make([]*Stock, 0, len(names))
	nrLiq  := nrand.New(rand.Int63())
	nrLiq.SetRange(0.0, 1.0)
	nrBPx  := nrand.New(rand.Int63())
	nrBPx.SetRange(0.0, 100.0)
	var totalLiq float64
	for _, name := range names {
		liq := nrLiq.NormFloat64()
		bpx := nrBPx.NormFloat64()
		stocks = append(stocks, &Stock{
			Name      :name,
			Liquidity :liq,
			BasePx    :bpx,
			LastTrdPx :bpx,
			Started   :false})
		totalLiq += liq
	}
	return stocks, totalLiq
}

func SetupExch(stocks []*Stock, totalLiq float64) *Exch {
	datebeg, err := time.Parse("20060102", opt.DateBeg)
	if err != nil {
		panic(err)
	}
	dateend, err := time.Parse("20060102", opt.DateEnd)
	if err != nil {
		panic(err)
	}
	starttm, err := time.Parse("1504", opt.StartTm)
	if err != nil {
		panic(err)
	}
	endtm, err   := time.Parse("1504", opt.EndTm)
	if err != nil {
		panic(err)
	}

	shour,sminute,ssecond := starttm.Clock()
	ehour,eminute,esecond := endtm.Clock()
	sduration := time.Duration(shour)*time.Hour + time.Duration(sminute)*time.Minute + time.Duration(ssecond)*time.Second
	eduration := time.Duration(ehour)*time.Hour + time.Duration(eminute)*time.Minute + time.Duration(esecond)*time.Second

	datenow     := datebeg
	datetimenow := datenow.Add(sduration)
	datetimeeod := datenow.Add(eduration)

	return &Exch{
		DateBeg		: datebeg,
		DateEnd		: dateend,
		DateNow		: datenow,
		DateTimeNow	: datetimenow,
		DateTimeEOD	: datetimeeod,
		SOD		: sduration,
		EOD		: eduration,
		Stocks          : stocks,
		TotalLiquidity  : totalLiq}
}

func (exch *Exch) GetNextTickTime() (t *time.Time, newDate bool, done bool) {
	newDate = false
	done    = false
	exch.DateTimeNow = exch.DateTimeNow.Add(time.Duration(rand.Intn(opt.Interval))*time.Millisecond)
	if exch.DateTimeNow.After(exch.DateTimeEOD) {
		exch.DateNow = exch.DateNow.Add(24*time.Hour)
		exch.DateTimeNow = exch.DateNow.Add(exch.SOD)
		exch.DateTimeEOD = exch.DateNow.Add(exch.EOD)
		newDate = true
	}
	if exch.DateNow.After(exch.DateEnd) {
		done = true
	}
	return &exch.DateTimeNow, newDate, done
}

func (exch *Exch) GetNextStock(stock chan<- *Stock) {
	nr := nrand.New(rand.Int63())
	nr.SetRange(0.0, 1.0)
	for {
		p  := nr.NormFloat64() * exch.TotalLiquidity
		ii := 0
		for {
			p -= exch.Stocks[ii].Liquidity
			if p <= 0.0 { break }
			ii ++
		}
		stock <- exch.Stocks[ii]
	}
}

func (stock *Stock) GenNextTradeQuote(nr *nrand.Nrand, ticktime *time.Time) {
	if stock.Started && rand.Int31n(20) > 17 {           // new trade
		stock.LastType = "t"
		nr.SetRange(stock.LastBidPx, stock.LastAskPx)
		stock.LastTrdPx = nr.NormFloat64()
		stock.LastTrdSz = (1 + rand.Int63n(50)) * 100
		stock.LastTrdTm = *ticktime
		stock.LastArrTm = ticktime.Add(time.Duration(rand.Int31n(5)+5)*time.Millisecond)
	} else {                                              // new quote
		stock.LastType = "q"
		
		absSpread := (1.0 - stock.Liquidity + .01) * stock.LastTrdPx / 100
		nr.SetRange(stock.LastTrdPx - absSpread, stock.LastTrdPx)
		stock.LastBidPx = nr.NormFloat64()
		nr.SetRange(stock.LastTrdPx, stock.LastTrdPx + absSpread)
		stock.LastAskPx = nr.NormFloat64()
		stock.LastBidSz = (1 + rand.Int63n(50)) * 100
		stock.LastAskSz = (1 + rand.Int63n(50)) * 100
		stock.LastQuoteTm = *ticktime
		stock.LastArrTm = ticktime.Add(time.Duration(rand.Int31n(5)+5)*time.Millisecond)
		stock.Started = true
	}
}

func InitNewDayFile(t *time.Time, outfilepat string) (*os.File) {
	fn := strings.Replace(outfilepat, "YYYYMMDD", t.Format("20060102"), -1)
	of, err := os.OpenFile(fn,os.O_WRONLY | os.O_CREATE, 0755)
	if err != nil {
		log.Fatal(err)
	}
	return of
}

func main() {
	sflag.Parse(&opt)
	rand.Seed(opt.Seed)
	names := GenNames(opt.NumStk)
	
	exch := SetupExch(GenStocks(names, opt.Seed))
	var nextStockChan = make(chan *Stock,500)

	ticktime, newDate, done := exch.GetNextTickTime()
	fp := InitNewDayFile(ticktime, opt.OutFilePat)
	fmt.Fprintf(fp,"date,arrTm,ticker,type,bidPx,bidSz,askPx,askSz,quotTm,trdPx,trdSz,trdTm\n")

	go exch.GetNextStock(nextStockChan)
	nr := nrand.New(rand.Int63())
	for !done {
		stock := <-nextStockChan
		stock.GenNextTradeQuote(nr, ticktime)
		if stock.LastType == "t" {
			fmt.Fprintf(fp,"%s,%s,%s,,,,,,%f,%d,%s\n",stock.LastArrTm.Format("20060102,150405.000"),stock.Name, stock.LastType, stock.LastTrdPx, stock.LastTrdSz, stock.LastTrdTm.Format("150405.000"))
		} else {
			fmt.Fprintf(fp,"%s,%s,%s,%f,%d,%f,%d,%s,,,\n",stock.LastArrTm.Format("20060102,150405.000"),stock.Name, stock.LastType, stock.LastBidPx, stock.LastBidSz, stock.LastAskPx, stock.LastAskSz, stock.LastQuoteTm.Format("150405.000"))
		}
		ticktime, newDate, done = exch.GetNextTickTime()
		if newDate {
			for _, ii := range(exch.Stocks) { ii.Started = false }
			if strings.Index(opt.OutFilePat, "YYYYMMDD") >= 0 {
				fp.Close()
				fp = InitNewDayFile(ticktime, opt.OutFilePat)
				fmt.Fprintf(fp,"date,arrTm,ticker,type,bidPx,bidSz,askPx,askSz,quotTm,trdPx,trdSz,trdTm\n")
			}
		}
	}
}
