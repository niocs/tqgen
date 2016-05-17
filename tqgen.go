package main

import (
	"fmt"
	"math/rand"
	"github.com/niocs/sflag"
	"github.com/niocs/nrand"
	"time"
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
	Stocks           []Stock
	TotalLiquidity   float64
}

var opt = struct {
	Usage    string  "Prints usage string"
	NumStk   int     "Number of stocks | 10"
	Seed     int64   "Randomization seed. Picked at random if not specified | 1"
	DateBeg  string  "Date in YYYYMMDD format | 20150101"
	DateEnd  string  "Date in YYYYMMDD format | 20150103"
	StartTm  string  "Trading start time | 0930"
	EndTm    string  "Trading end time | 1600"
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
	for len(nameMap) < opt.NumStk {
		nameMap[GenName()] = struct{}{}
	}
	names := make([]string, 0, len(nameMap))
	for name := range nameMap {
		names = append(names, name)
	}
	return names
}

func GenStocks(names []string, seed int64) ([]Stock, float64) {
	stocks := make([]Stock, 0, len(names))
	nrLiq  := nrand.New(rand.Int63())
	nrLiq.SetRange(0.0, 1.0)
	nrBPx  := nrand.New(rand.Int63())
	nrBPx.SetRange(0.0, 100.0)
	var totalLiq float64
	for _, name := range names {
		liq := nrLiq.NormFloat64()
		bpx := nrBPx.NormFloat64()
		stocks = append(stocks, Stock{
			Name      :name,
			Liquidity :liq,
			BasePx    :bpx,
			LastTrdPx :bpx,
			Started   :false})
		totalLiq += liq
	}
	return stocks, totalLiq
}

func SetupExch(stocks []Stock, totalLiq float64) *Exch {
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

	datenow := datebeg
	datetimenow := datenow.Add(sduration)
	datetimeeod := datenow.Add(eduration)

	return &Exch{
		DateBeg		    : datebeg,
		DateEnd		    : dateend,
		DateNow		    : datenow,
		DateTimeNow	    : datetimenow,
		DateTimeEOD	    : datetimeeod,
		SOD		        : sduration,
		EOD		        : eduration,
		Stocks          : stocks,
		TotalLiquidity  : totalLiq}
}

func (exch *Exch) GetNextTickTime() (time.Time, bool) {  // bool -> true if next tick, false if no more ticks
	exch.DateTimeNow = exch.DateTimeNow.Add(time.Duration(rand.Intn(25))*time.Millisecond)
	if exch.DateTimeNow.After(exch.DateTimeEOD) {
		exch.DateNow = exch.DateNow.Add(24*time.Hour)
		exch.DateTimeNow = exch.DateNow.Add(exch.SOD)
		exch.DateTimeEOD = exch.DateNow.Add(exch.EOD)
	}
	if exch.DateNow.After(exch.DateEnd) {
		return exch.DateTimeNow, false
	}
	return exch.DateTimeNow, true
}

func (exch *Exch) GetNextStock() *Stock {
	nr := nrand.New(rand.Int63())
	nr.SetRange(0.0, 1.0)
	p  := nr.NormFloat64() * exch.TotalLiquidity
	ii := 0
	for {
		p -= exch.Stocks[ii].Liquidity
		if p <= 0.0 { break }
		ii ++
	}
	return &exch.Stocks[ii]
}

func (stock *Stock) GenNextTradeQuote(ticktime time.Time) {
	if stock.Started && rand.Int31n(20) > 17 {           // new trade
		stock.LastType = "t"
		nr := nrand.New(rand.Int63())                 // TODO: this can be slow. improve later
		nr.SetRange(stock.LastBidPx, stock.LastAskPx)
		stock.LastTrdPx = nr.NormFloat64()
		stock.LastTrdSz = (1 + rand.Int63n(50)) * 100
		stock.LastTrdTm = ticktime
		stock.LastArrTm = ticktime.Add(time.Duration(rand.Int31n(5)+5)*time.Millisecond)
	} else {                                              // new quote
		stock.LastType = "q"
		
		absSpread := (1.0 - stock.Liquidity + .01) * stock.LastTrdPx / 100
		nr := nrand.New(rand.Int63())
		nr.SetRange(stock.LastTrdPx - absSpread, stock.LastTrdPx)
		stock.LastBidPx = nr.NormFloat64()
		nr.SetRange(stock.LastTrdPx, stock.LastTrdPx + absSpread)
		stock.LastAskPx = nr.NormFloat64()
		stock.LastBidSz = (1 + rand.Int63n(50)) * 100
		stock.LastAskSz = (1 + rand.Int63n(50)) * 100
		stock.LastQuoteTm = ticktime
		stock.LastArrTm = ticktime.Add(time.Duration(rand.Int31n(5)+5)*time.Millisecond)
		stock.Started = true
	}
}

func main() {
	sflag.Parse(&opt)
	rand.Seed(opt.Seed)
	names := GenNames(opt.NumStk)
	
	exch := SetupExch(GenStocks(names, opt.Seed))

	fmt.Printf("date,arrTm,ticker,type,bidPx,bidSz,askPx,askSz,quotTm,trdPx,trdSz,trdTm\n")
	ticktime, goodtick := exch.GetNextTickTime()
	for goodtick == true {
		stock := exch.GetNextStock()
		stock.GenNextTradeQuote(ticktime)
		if stock.LastType == "t" {
			fmt.Printf("%s,%s,%s,,,,,,%f,%d,%s\n",stock.LastArrTm.Format("20060102,150405.000"),stock.Name, stock.LastType, stock.LastTrdPx, stock.LastTrdSz, stock.LastTrdTm.Format("150405.000"))
		} else {
			fmt.Printf("%s,%s,%s,%f,%d,%f,%d,%s,,,\n",stock.LastArrTm.Format("20060102,150405.000"),stock.Name, stock.LastType, stock.LastBidPx, stock.LastBidSz, stock.LastAskPx, stock.LastAskSz, stock.LastQuoteTm.Format("150405.000"))
		}
		ticktime, goodtick = exch.GetNextTickTime()
	}
}
