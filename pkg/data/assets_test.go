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

func TestStock(t *testing.T) {

	params := gopter.DefaultTestParametersWithSeed(42)
	ps := gopter.NewProperties(params)
	arbs := arbitrary.DefaultArbitraries()

	arbs.RegisterGen(GenStock(GenTicker()))
	arbs.RegisterGen(gen.SliceOf(GenStock(GenTicker())))

	ps.Property("Stock.Dir == S when Stock.Price < 0", prop.ForAll(
		func(s Stock) bool {
			if s.Price < 0 {
				return s.Dir() == S
			}
			return s.Dir() == L
		},
		GenStock(GenTickers())))

	ps.Property("Stock.Empty == true when Stock object empty", prop.ForAll(
		func(s Stock) bool {
			if s.Ticker == "" {
				return s.Empty()
			}
			return !s.Empty()
		},
		GenStock(GenTickers())))

	ps.Property("Stocks sort by price", arbs.ForAll(
		func(ss Stocks) bool {
			ps := make([]float64, len(ss))
			for i, s := range ss {
				ps[i] = s.Price
			}
			sort.Float64s(ps)
			sort.Sort(ss)

			for i, s := range ss {
				if ps[i] != s.Price {
					return false
				}
			}
			return true
		}))

	ps.Property("Stocks.Price == sum of all prices", arbs.ForAll(
		func(ss Stocks) bool {
			sumP := 0.0
			for _, s := range ss {
				sumP += s.Price
			}
			eSumP := ss.Price()
			return sumP == eSumP
		}))

	ps.Property("Stocks.Shares == sum of all shares", arbs.ForAll(
		func(ss Stocks) bool {
			sumS := 0
			for _, s := range ss {
				sumS += s.Shares
			}
			eSumS := ss.Shares()
			return sumS == eSumS
		}))

	ps.Run(gopter.NewFormatedReporter(true, 80, os.Stdout))
}

func TestPut(t *testing.T) {

	params := gopter.DefaultTestParametersWithSeed(42)
	ps := gopter.NewProperties(params)
	arbs := arbitrary.DefaultArbitraries()

	arbs.RegisterGen(GenCall(GenTickers()))
	arbs.RegisterGen(gen.SliceOf(GenPut(GenTickers())))

	ps.Property("Put.Dir == S when Put.Price < 0", prop.ForAll(
		func(p Put) bool {
			if p.Price < 0 {
				return p.Dir() == S
			}
			return p.Dir() == L
		},
		GenPut(GenTickers())))

	ps.Property("Put.Empty == true when Put object empty", prop.ForAll(
		func(p Put) bool {
			if p.Underlying.Ticker == "" {
				return p.Empty()
			}
			return !p.Empty()
		},
		GenPut(GenTickers())))

	ps.Property("Puts sort by ticker if > 1 ticker exists", arbs.ForAll(
		func(ps Puts) bool {
			tickers := make([]string, len(ps))
			for i, p := range ps {
				tickers[i] = p.Underlying.Ticker
			}
			sort.Strings(tickers)
			sort.Sort(ps)

			for i, p := range ps {
				if p.Underlying.Ticker != tickers[i] {
					return false
				}
			}
			return true
		}))

	ps.Property("Puts sort by strike if all same ticker", prop.ForAll(
		func(ps Puts) bool {
			strikes := make([]float64, len(ps))
			for i, p := range ps {
				strikes[i] = p.Strike
			}
			sort.Float64s(strikes)
			sort.Sort(ps)

			for i, p := range ps {
				if p.Strike != strikes[i] {
					return false
				}
			}
			return true
		}, gen.SliceOf(GenPut(GenTicker()))))

	ps.Property("Puts.Price == sum of all prices", arbs.ForAll(
		func(ps Puts) bool {
			sum := 0.0
			for _, p := range ps {
				sum += p.Price
			}
			return sum == ps.Price()
		}))

	ps.Run(gopter.NewFormatedReporter(true, 80, os.Stdout))
}

func TestCall(t *testing.T) {

	params := gopter.DefaultTestParametersWithSeed(42)
	ps := gopter.NewProperties(params)
	arbs := arbitrary.DefaultArbitraries()

	arbs.RegisterGen(GenCall(GenTickers()))
	arbs.RegisterGen(gen.SliceOf(GenPut(GenTickers())))

	ps.Property("Call.Dir == S when Call.Price < 0", prop.ForAll(
		func(c Call) bool {
			if c.Price < 0 {
				return c.Dir() == S
			}
			return c.Dir() == L
		},
		GenCall(GenTickers())))

	ps.Property("Call.Empty == true when Call object empty", prop.ForAll(
		func(c Call) bool {
			if c.Underlying.Ticker == "" {
				return c.Empty()
			}
			return !c.Empty()
		},
		GenCall(GenTickers())))

	ps.Property("Calls sort by ticker if > 1 ticker exists", arbs.ForAll(
		func(cs Calls) bool {
			tickers := make([]string, len(cs))
			for i, c := range cs {
				tickers[i] = c.Underlying.Ticker
			}
			sort.Strings(tickers)
			sort.Sort(cs)

			for i, c := range cs {
				if c.Underlying.Ticker != tickers[i] {
					return false
				}
			}
			return true
		}))

	ps.Property("Calls sort by strike if all same ticker", prop.ForAll(
		func(cs Calls) bool {
			strikes := make([]float64, len(cs))
			for i, c := range cs {
				strikes[i] = c.Strike
			}
			sort.Float64s(strikes)
			sort.Sort(cs)

			for i, c := range cs {
				if c.Strike != strikes[i] {
					return false
				}
			}
			return true
		}, gen.SliceOf(GenCall(GenTicker()))))

	ps.Property("Calls.Price == sum of all prices", arbs.ForAll(
		func(cs Calls) bool {
			sum := 0.0
			for _, c := range cs {
				sum += c.Price
			}
			return sum == cs.Price()
		}))

	ps.Run(gopter.NewFormatedReporter(true, 80, os.Stdout))
}
