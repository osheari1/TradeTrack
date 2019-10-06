package data

import (
	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/arbitrary"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
	"os"
	"sort"
	"testing"
)

func TestStrategy(t *testing.T) {

	params := gopter.DefaultTestParametersWithSeed(42)
	ps := gopter.NewProperties(params)
	arbs := arbitrary.DefaultArbitraries()

	arbs.RegisterGen(GenStock(GenTickers()))
	arbs.RegisterGen(GenPut(GenTickers()))
	arbs.RegisterGen(GenCall(GenTickers()))
	arbs.RegisterGen(gen.SliceOf(GenStock(GenTickers())))
	arbs.RegisterGen(gen.SliceOf(GenCall(GenTickers())))
	arbs.RegisterGen(gen.SliceOf(GenPut(GenTickers())))

	ps.Property("parseTicker return ('', false) when arguments empty", prop.ForAll(
		func(ss Stocks, ps Puts, cs Calls) bool {
			v, ok := parseTicker(ss, ps, cs)
			return v == "" && ok == false
		},
		gen.SliceOfN(0, GenStock(GenTickers())),
		gen.SliceOfN(0, GenPut(GenTickers())),
		gen.SliceOfN(0, GenCall(GenTickers()))))

	ps.Property("parseTicker return first ticket in non-empty arguments", arbs.ForAll(
		func(ss Stocks, ps Puts, cs Calls) bool {
			v, ok := parseTicker(ss, ps, cs)
			if !ok {
				return v == ""
			} else if len(ps) == 0 && len(cs) == 0 {
				return v == ss[0].Ticker
			} else if len(ps) == 0 {
				return v == cs[0].Underlying.Ticker
			} else {
				return v == ps[0].Underlying.Ticker
			}
		}))

	ps.Property("parseStocks throws error on non-matching ticker", prop.ForAll(
		func(ss Stocks) bool {
			e := parseStocks(&Strategy{}, ss)
			if e != nil {
				return true
			}
			return false
		},
		gen.SliceOfN(5, GenStock(GenTicker()))))

	ps.Property("parseStocks adds sorted stocks to strategy", prop.ForAll(
		func(ss Stocks) bool {
			s := Strategy{Ticker: ss[0].Ticker}
			e := parseStocks(&s, ss)

			if e != nil {
				return false
			}

			sort.Sort(ss)

			for i, st := range ss {
				if st.Price != s.Stocks[i].Price {
					return false
				}
			}
			return true
		},
		gen.SliceOfN(5, GenStock(GenTicker()))))

	ps.Property("parseCalls throws error on non-matching ticker", prop.ForAll(
		func(cs Calls) bool {
			e := parseCalls(&Strategy{}, cs)
			if e != nil {
				return true
			}
			return false
		},
		gen.SliceOfN(5, GenCall(GenTicker()))))

	ps.Property("parseCalls adds sorted calls to strategy split by direction", prop.ForAll(
		func(cs Calls) bool {
			s := Strategy{Ticker: cs[0].Underlying.Ticker}
			e := parseCalls(&s, cs)
			if e != nil {
				return false
			}

			sort.Sort(cs)
			li, si := 0, 0
			for _, c := range cs {
				if c.Dir() == L {
					if s.Lc[li].Strike != c.Strike {
						return false
					}
					li++
				} else {
					if s.Sc[si].Strike != c.Strike {
						return false
					}
					si++
				}
			}
			return true
		},
		gen.SliceOfN(5, GenCall(GenTicker()))))

	ps.Property("parsePuts throws error on non-matching ticker", prop.ForAll(
		func(ps Puts) bool {
			e := parsePuts(&Strategy{}, ps)
			if e != nil {
				return true
			}
			return false
		},
		gen.SliceOfN(5, GenPut(GenTicker()))))

	ps.Property("parsePuts adds sorted puts to strategy split by direction", prop.ForAll(
		func(ps Puts) bool {
			s := Strategy{Ticker: ps[0].Underlying.Ticker}
			e := parsePuts(&s, ps)
			if e != nil {
				return false
			}

			sort.Sort(ps)

			li, si := 0, 0
			for _, p := range ps {
				if p.Dir() == L {
					if s.Lp[li].Strike != p.Strike {
						return false
					}
					li++
				} else {
					if s.Sp[si].Strike != p.Strike {
						return false
					}
					si++
				}
			}
			return true
		},
		gen.SliceOfN(5, GenPut(GenTicker()))))

	ps.Property("Strategy.Price == price of underlying assets", prop.ForAll(
		func(s Strategy) bool {
			return s.Price() == s.Stocks.Price()+s.PriceOptions()
		},
		GenStrategy(GenTicker())))

	ps.Run(gopter.NewFormatedReporter(true, 80, os.Stdout))
}

func TestStrategyCreation(t *testing.T) {

	params := gopter.DefaultTestParametersWithSeed(42)
	ps := gopter.NewProperties(params)

	check := func(s Strategy) bool {
		t, d := s.CheckKind()
		if t != s.Type || d != s.Dir {
			return false
		}
		return true
	}

	ps.Property("Long put spread", prop.ForAll(
		func(s Strategy) bool {
			return check(s)
		},
		GenLongPutSpreadStrategy(GenTicker())))

	ps.Property("Short put spread", prop.ForAll(
		func(s Strategy) bool {
			return check(s)
		},
		GenShortPutSpreadStrategy(GenTicker())))

	ps.Property("Empty", prop.ForAll(
		func(s Strategy) bool {
			return check(s)
		},
		GenEmptyStrategy()))

	ps.Property("Long strangle", prop.ForAll(
		func(s Strategy) bool {
			return check(s)
		},
		GenLongStrangleStrategy(GenTicker())))

	ps.Property("Short strangle", prop.ForAll(
		func(s Strategy) bool {
			return check(s)
		},
		GenShortStrangleStrategy(GenTicker())))

	ps.Property("Long straddle", prop.ForAll(
		func(s Strategy) bool {
			return check(s)
		},
		GenLongStraddleStrategy(GenTicker())))

	ps.Property("Short straddle", prop.ForAll(
		func(s Strategy) bool {
			return check(s)
		},
		GenShortStraddleStrategy(GenTicker())))

	ps.Property("Long covered call", prop.ForAll(
		func(s Strategy) bool {
			return check(s)
		},
		GenLongCoveredCallStrategy(GenTicker())))

	ps.Property("Short covered call", prop.ForAll(
		func(s Strategy) bool {
			return check(s)
		},
		GenShortCoveredCallStrategy(GenTicker())))

	ps.Property("Long covered put", prop.ForAll(
		func(s Strategy) bool {
			return check(s)
		},
		GenLongCoveredPutStrategy(GenTicker())))

	ps.Property("Short covered put", prop.ForAll(
		func(s Strategy) bool {
			return check(s)
		},
		GenShortCoveredPutStrategy(GenTicker())))

	ps.Property("Long iron condor", prop.ForAll(
		func(s Strategy) bool {
			return check(s)
		},
		GenLongIronCondorStrategy(GenTicker())))

	ps.Property("Short iron condor", prop.ForAll(
		func(s Strategy) bool {
			return check(s)
		},
		GenShortIronCondorStrategy(GenTicker())))

	ps.Property("Long iron butterfly", prop.ForAll(
		func(s Strategy) bool {
			return check(s)
		},
		GenLongIronButterflyStrategy(GenTicker())))

	ps.Property("Short iron butterfly", prop.ForAll(
		func(s Strategy) bool {
			return check(s)
		},
		GenShortIronButterflyStrategy(GenTicker())))

	ps.Property("Long call butterfly", prop.ForAll(
		func(s Strategy) bool {
			return check(s)
		},
		GenLongCallButterflyStrategy(GenTicker())))

	ps.Property("Short call butterfly", prop.ForAll(
		func(s Strategy) bool {
			return check(s)
		},
		GenShortCallButterflyStrategy(GenTicker())))

	ps.Property("Long put butterfly", prop.ForAll(
		func(s Strategy) bool {
			return check(s)
		},
		GenLongPutButterflyStrategy(GenTicker())))

	ps.Property("Short put butterfly", prop.ForAll(
		func(s Strategy) bool {
			return check(s)
		},
		GenShortPutButterflyStrategy(GenTicker())))

	ps.Property("Long jade lizard", prop.ForAll(
		func(s Strategy) bool {
			return check(s)
		},
		GenLongJadeLizardStrategy(GenTicker())))

	ps.Property("Short jade lizard", prop.ForAll(
		func(s Strategy) bool {
			return check(s)
		},
		GenShortJadeLizardStrategy(GenTicker())))

	ps.Property("Long naked stock", prop.ForAll(
		func(s Strategy) bool {
			return check(s)
		},
		GenLongNakedStockStrategy(GenTicker())))

	ps.Property("Short naked stock", prop.ForAll(
		func(s Strategy) bool {
			return check(s)
		},
		GenShortNakedStockStrategy(GenTicker())))

	ps.Property("Long naked call", prop.ForAll(
		func(s Strategy) bool {
			return check(s)
		},
		GenLongNakedCallStrategy(GenTicker())))

	ps.Property("Short naked call", prop.ForAll(
		func(s Strategy) bool {
			return check(s)

		},
		GenShortNakedCallStrategy(GenTicker())))

	ps.Property("Long naked put", prop.ForAll(
		func(s Strategy) bool {
			return check(s)

		},
		GenLongNakedPutStrategy(GenTicker())))

	ps.Property("Short naked put", prop.ForAll(
		func(s Strategy) bool {
			return check(s)

		},
		GenShortNakedPutStrategy(GenTicker())))

	ps.Property("Long custom", prop.ForAll(
		func(s Strategy) bool {
			return check(s)
		},
		GenLongCustomStrategy(GenTicker())))

	ps.Property("Short custom", prop.ForAll(
		func(s Strategy) bool {
			return check(s)

		},
		GenShortCustomStrategy(GenTicker())))

	ps.Run(gopter.NewFormatedReporter(true, 80, os.Stdout))
}
