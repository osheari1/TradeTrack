package data

import (
	"fmt"
	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"math"
	"reflect"
	"strconv"
)

const (
	MaxStrike      float64 = 1000
	MinStrike      float64 = 100
	MaxOptionPrice float64 = 10
	MaxStockPrice  float64 = 1000
	MaxShares      int     = 200
)

func GenDirection() gopter.Gen {
	return gen.IntRange(0, 2).Map(func(i int) Direction {
		return Direction(i)
	})
}

func GenType() gopter.Gen {
	return gen.IntRange(0, 15).Map(func(i int) Type {
		return Type(i)
	})
}

func GenTickers() gopter.Gen {
	return gopter.CombineGens(
		gen.AlphaUpperChar(),
		gen.AlphaUpperChar(),
		gen.AlphaUpperChar()).Map(func(rs []interface{}) string {
		s := ""
		for _, r := range rs {
			i, _ := strconv.Atoi(fmt.Sprint(r))
			s += string(rune(i))
		}
		return s
	})
}

func GenTicker() gopter.Gen {
	t, _ := GenTickers().Sample()
	return gen.Const(t)
}

func GenStock(ticker gopter.Gen) gopter.Gen {
	return gen.Struct(
		reflect.TypeOf(Stock{}),
		map[string]gopter.Gen{
			"Ticker": ticker,
			"Price":  gen.Float64Range(-MaxStockPrice, MaxStockPrice),
			"Shares": gen.IntRange(1, MaxShares)})
}

func GenStock100Shares(ticker gopter.Gen) gopter.Gen {
	return gen.Struct(
		reflect.TypeOf(Stock{}),
		map[string]gopter.Gen{
			"Ticker": ticker,
			"Price":  gen.Float64Range(-MaxStockPrice, MaxStockPrice),
			"Shares": gen.Const(100)})

}

func GenShortStock(ticker gopter.Gen) gopter.Gen {
	return GenStock(ticker).SuchThat(func(st Stock) bool {
		return st.Price < 0
	})
}

func GenLongStock(ticker gopter.Gen) gopter.Gen {
	return GenStock(ticker).SuchThat(func(st Stock) bool {
		return st.Price >= 0
	})
}

func GenShortStock100Shares(ticker gopter.Gen) gopter.Gen {
	return GenStock100Shares(ticker).SuchThat(func(st Stock) bool {
		return st.Price < 0
	})
}

func GenLongStock100Shares(ticker gopter.Gen) gopter.Gen {
	return GenStock100Shares(ticker).SuchThat(func(st Stock) bool {
		return st.Price >= 0
	})
}

func GenPut(ticker gopter.Gen) gopter.Gen {
	return gen.Struct(
		reflect.TypeOf(Put{}),
		map[string]gopter.Gen{
			"Underlying": GenStock100Shares(ticker),
			"Price":      gen.Float64Range(-MaxOptionPrice, MaxOptionPrice),
			"Strike":     gen.Float64Range(MinStrike, MaxStrike)})
}

func GenShortPut(ticker gopter.Gen) gopter.Gen {
	return GenPut(ticker).Map(func(p Put) Put {
		p.Price = -math.Abs(p.Price)
		return p
	})
}

func GenShortPutWithStrike(ticker gopter.Gen, strike gopter.Gen) gopter.Gen {
	return GenShortPut(ticker).FlatMap(func(p interface{}) gopter.Gen {
		return strike.Map(func(strike float64) Put {
			p := p.(Put)
			p.Strike = strike
			return p
		})
	}, reflect.TypeOf(Put{}))
}

func GenLongPut(ticker gopter.Gen) gopter.Gen {
	return GenPut(ticker).Map(func(p Put) Put {
		p.Price = math.Abs(p.Price)
		return p
	})
}

func GenLongPutWithStrike(ticker gopter.Gen, strike gopter.Gen) gopter.Gen {
	return GenLongPut(ticker).FlatMap(func(p interface{}) gopter.Gen {
		return strike.Map(func(strike float64) Put {
			p := p.(Put)
			p.Strike = strike
			return p
		})
	}, reflect.TypeOf(Put{}))
}

func GenCall(ticker gopter.Gen) gopter.Gen {
	return gen.Struct(
		reflect.TypeOf(Call{}),
		map[string]gopter.Gen{
			"Underlying": GenStock100Shares(ticker),
			"Price":      gen.Float64Range(-MaxOptionPrice, MaxOptionPrice),
			"Strike":     gen.Float64Range(MinStrike, MaxStrike)})
}

func GenShortCall(ticker gopter.Gen) gopter.Gen {
	return GenCall(ticker).Map(func(c Call) Call {
		c.Price = -math.Abs(c.Price)
		return c
	})
}

func GenShortCallWithStrike(ticker gopter.Gen, strike gopter.Gen) gopter.Gen {
	return GenShortCall(ticker).FlatMap(func(c interface{}) gopter.Gen {
		return strike.Map(func(strike float64) Call {
			p := c.(Call)
			p.Strike = strike
			return p
		})
	}, reflect.TypeOf(Call{}))
}

func GenLongCall(ticker gopter.Gen) gopter.Gen {
	return GenCall(ticker).Map(func(c Call) Call {
		c.Price = math.Abs(c.Price)
		return c
	})
}

func GenLongCallWithStrike(ticker gopter.Gen, strike gopter.Gen) gopter.Gen {
	return GenLongCall(ticker).FlatMap(func(c interface{}) gopter.Gen {
		return strike.Map(func(strike float64) Call {
			p := c.(Call)
			p.Strike = strike
			return p
		})
	}, reflect.TypeOf(Call{}))
}

func GenLongSpreadPuts(ticker gopter.Gen) gopter.Gen {
	lp := GenLongPut(ticker)
	sp := GenShortPut(ticker)

	return lp.FlatMap(func(p1 interface{}) gopter.Gen {
		return sp.FlatMap(func(p2 interface{}) gopter.Gen {
			s1 := p1.(Put).Strike
			s2 := gen.Float64Range(math.Min(s1-2, 0), s1-1)
			return s2.Map(func(s2 float64) Puts {
				p2 := p2.(Put)
				p2.Strike = s2
				return Puts{p2, p1.(Put)}
			})
		}, reflect.TypeOf(Puts{}))
	}, reflect.TypeOf(Puts{}))
}

func GenShortSpreadPuts(ticker gopter.Gen) gopter.Gen {
	return GenLongSpreadPuts(ticker).Map(func(ps Puts) Puts {
		s, l := ps[0], ps[1]
		s.Strike, l.Strike = l.Strike, s.Strike
		return Puts{l, s}
	})
}

func GenLongPutSpreadStrategy(ticker gopter.Gen) gopter.Gen {
	puts, _ := GenLongSpreadPuts(ticker).Sample()
	lp := gen.Const(Puts{puts.(Puts)[1]})
	sp := gen.Const(Puts{puts.(Puts)[0]})
	return gen.Struct(
		reflect.TypeOf(Strategy{}),
		map[string]gopter.Gen{
			"Ticker": GenTicker(),
			"Lp":     lp,
			"Sp":     sp,
			"Type":   gen.Const(Spread),
			"Dir":    gen.Const(L)})
}

func GenShortPutSpreadStrategy(ticker gopter.Gen) gopter.Gen {
	puts, _ := GenShortSpreadPuts(ticker).Sample()
	lp := gen.Const(Puts{puts.(Puts)[0]})
	sp := gen.Const(Puts{puts.(Puts)[1]})

	return gen.Struct(
		reflect.TypeOf(Strategy{}),
		map[string]gopter.Gen{
			"Ticker": ticker,
			"Lp":     lp,
			"Sp":     sp,
			"Type":   gen.Const(Spread),
			"Dir":    gen.Const(S)})
}

func GenEmptyStrategy() gopter.Gen {
	return gen.Const(Strategy{
		Type: Empty,
		Dir:  None})
}

func GenLongStrangleStrategy(ticker gopter.Gen) gopter.Gen {
	lps, _ := GenLongPut(ticker).Sample()
	lc := GenLongCall(ticker).FlatMap(func(lc interface{}) gopter.Gen {
		sp := lps.(Put).Strike
		sc := gen.Float64Range(sp+1, MaxStrike)
		return sc.Map(func(sc float64) Calls {
			lc := lc.(Call)
			lc.Strike = sc
			return Calls{lc}
		})
	}, reflect.TypeOf(Calls{}))
	lp := gen.Const(Puts{lps.(Put)})

	return gen.Struct(
		reflect.TypeOf(Strategy{}),
		map[string]gopter.Gen{
			"Ticker": ticker,
			"Lp":     lp,
			"Lc":     lc,
			"Type":   gen.Const(Strangle),
			"Dir":    gen.Const(L)})
}

func GenShortStrangleStrategy(ticker gopter.Gen) gopter.Gen {
	return GenLongStrangleStrategy(ticker).Map(func(s Strategy) Strategy {
		sp := s.Lp[0]
		sc := s.Lc[0]
		sp.Price = -sp.Price
		sc.Price = -sc.Price
		return Strategy{
			Ticker: s.Ticker,
			Sp:     Puts{sp},
			Sc:     Calls{sc},
			Type:   Strangle,
			Dir:    S}
	})
}

func GenLongStraddleStrategy(ticker gopter.Gen) gopter.Gen {
	return GenLongStrangleStrategy(ticker).Map(func(s Strategy) Strategy {
		lp := s.Lp[0]
		lc := s.Lc[0]
		lc.Strike = lp.Strike
		return Strategy{
			Ticker: s.Ticker,
			Lp:     Puts{lp},
			Lc:     Calls{lc},
			Type:   Straddle,
			Dir:    L}
	})
}

func GenShortStraddleStrategy(ticker gopter.Gen) gopter.Gen {
	return GenShortStrangleStrategy(ticker).Map(func(s Strategy) Strategy {
		sp := s.Sp[0]
		sc := s.Sc[0]
		sc.Strike = sp.Strike
		return Strategy{
			Ticker: s.Ticker,
			Sp:     Puts{sp},
			Sc:     Calls{sc},
			Type:   Straddle,
			Dir:    S}
	})
}

func GenLongCoveredCallStrategy(ticker gopter.Gen) gopter.Gen {
	return gen.Struct(
		reflect.TypeOf(Strategy{}),
		map[string]gopter.Gen{
			"Ticker": ticker,
			"Stocks": gen.SliceOfN(1, GenShortStock100Shares(ticker)),
			"Lc":     gen.SliceOfN(1, GenLongCall(ticker)),
			"Type":   gen.Const(CoveredCall),
			"Dir":    gen.Const(L)})
}

func GenShortCoveredCallStrategy(ticker gopter.Gen) gopter.Gen {
	return gen.Struct(
		reflect.TypeOf(Strategy{}),
		map[string]gopter.Gen{
			"Ticker": ticker,
			"Stocks": gen.SliceOfN(1, GenLongStock100Shares(ticker)),
			"Sc":     gen.SliceOfN(1, GenShortCall(ticker)),
			"Type":   gen.Const(CoveredCall),
			"Dir":    gen.Const(S)})

}

func GenLongCoveredPutStrategy(ticker gopter.Gen) gopter.Gen {
	return gen.Struct(
		reflect.TypeOf(Strategy{}),
		map[string]gopter.Gen{
			"Ticker": ticker,
			"Stocks": gen.SliceOfN(1, GenLongStock100Shares(ticker)),
			"Lp":     gen.SliceOfN(1, GenLongPut(ticker)),
			"Type":   gen.Const(CoveredPut),
			"Dir":    gen.Const(L)})
}

func GenShortCoveredPutStrategy(ticker gopter.Gen) gopter.Gen {
	return gen.Struct(
		reflect.TypeOf(Strategy{}),
		map[string]gopter.Gen{
			"Ticker": ticker,
			"Stocks": gen.SliceOfN(1, GenShortStock100Shares(ticker)),
			"Sp":     gen.SliceOfN(1, GenShortPut(ticker)),
			"Type":   gen.Const(CoveredPut),
			"Dir":    gen.Const(S)})
}

func GenLongIronCondorStrategy(ticker gopter.Gen) gopter.Gen {
	return GenLongStrangleStrategy(ticker).FlatMap(func(s interface{}) gopter.Gen {
		lp := s.(Strategy).Lp
		lc := s.(Strategy).Lc
		sp := GenShortPutWithStrike(ticker, gen.Float64Range(1, lp[0].Strike-1))
		sc := GenShortCallWithStrike(ticker, gen.Float64Range(lc[0].Strike+1, MaxStrike))
		return gen.Struct(
			reflect.TypeOf(Strategy{}),
			map[string]gopter.Gen{
				"Ticker": ticker,
				"Sp":     gen.SliceOfN(1, sp),
				"Lp":     gen.Const(lp),
				"Lc":     gen.Const(lc),
				"Sc":     gen.SliceOfN(1, sc),
				"Type":   gen.Const(IronCondor),
				"Dir":    gen.Const(L)})

	}, reflect.TypeOf(Strategy{}))
}

func GenShortIronCondorStrategy(ticker gopter.Gen) gopter.Gen {
	return GenLongIronCondorStrategy(ticker).Map(func(s Strategy) Strategy {
		sp := s.Sp[0]
		lp := s.Lp[0]
		lc := s.Lc[0]
		sc := s.Sc[0]

		// Swap lp <-> sp && lc <-sc
		sp.Price, lp.Price = -lp.Price, -sp.Price
		sc.Price, sc.Price = -sc.Price, -sc.Price

		return Strategy{
			Ticker: s.Ticker,
			Lp:     Puts{sp},
			Sp:     Puts{lp},
			Sc:     Calls{lc},
			Lc:     Calls{sc},
			Type:   IronCondor,
			Dir:    S}
	})
}

func GenLongIronButterflyStrategy(ticker gopter.Gen) gopter.Gen {
	return GenLongStraddleStrategy(ticker).FlatMap(func(s interface{}) gopter.Gen {
		lp := s.(Strategy).Lp
		lc := s.(Strategy).Lc
		sp := GenShortPutWithStrike(ticker, gen.Float64Range(1, lp[0].Strike-1))
		sc := GenShortCallWithStrike(ticker, gen.Float64Range(lc[0].Strike+1, MaxStrike))
		return gen.Struct(
			reflect.TypeOf(Strategy{}),
			map[string]gopter.Gen{
				"Ticker": ticker,
				"Sp":     gen.SliceOfN(1, sp),
				"Lp":     gen.Const(lp),
				"Lc":     gen.Const(lc),
				"Sc":     gen.SliceOfN(1, sc),
				"Type":   gen.Const(IronButterfly),
				"Dir":    gen.Const(L)})

	}, reflect.TypeOf(Strategy{}))
}

func GenShortIronButterflyStrategy(ticker gopter.Gen) gopter.Gen {
	return GenLongIronButterflyStrategy(ticker).Map(func(s Strategy) Strategy {
		sp := s.Sp[0]
		lp := s.Lp[0]
		lc := s.Lc[0]
		sc := s.Sc[0]

		// Swap lp <-> sp && lc <-sc
		sp.Price, lp.Price = -lp.Price, -sp.Price
		sc.Price, sc.Price = -sc.Price, -sc.Price

		return Strategy{
			Ticker: s.Ticker,
			Lp:     Puts{sp},
			Sp:     Puts{lp},
			Sc:     Calls{lc},
			Lc:     Calls{sc},
			Type:   IronButterfly,
			Dir:    S}
	})
}

func GenLongCallButterflyStrategy(ticker gopter.Gen) gopter.Gen {
	sc := GenShortCall(ticker)
	return sc.FlatMap(func(sc interface{}) gopter.Gen {
		lcl := GenLongCallWithStrike(ticker, gen.Float64Range(0, sc.(Call).Strike-1))
		lcu := GenLongCallWithStrike(ticker, gen.Float64Range(sc.(Call).Strike+1, MaxStrike))
		lc := lcl.FlatMap(func(lcl interface{}) gopter.Gen {
			return lcu.Map(func(lcu Call) Calls {
				return Calls{lcl.(Call), lcu}
			})
		}, reflect.TypeOf(Call{}))

		return gen.Struct(
			reflect.TypeOf(Strategy{}),
			map[string]gopter.Gen{
				"Ticker": ticker,
				"Sc":     gen.Const(Calls{sc.(Call), sc.(Call)}),
				"Lc":     lc,
				"Type":   gen.Const(CallButterfly),
				"Dir":    gen.Const(L)})
	}, reflect.TypeOf(Call{}))
}

func GenShortCallButterflyStrategy(ticker gopter.Gen) gopter.Gen {
	return GenLongCallButterflyStrategy(ticker).Map(func(s Strategy) Strategy {
		lc := s.Lc
		sc := s.Sc

		lc[0].Price, lc[1].Price, sc[0].Price, sc[1].Price = -sc[0].Price, -sc[0].Price, -lc[0].Price, -lc[1].Price
		lc[0].Strike, lc[1].Strike, sc[0].Strike, sc[1].Strike = sc[0].Strike, sc[1].Strike, lc[0].Strike, lc[1].Strike

		return Strategy{
			Ticker: s.Ticker,
			Lc:     lc,
			Sc:     sc,
			Type:   CallButterfly,
			Dir:    S}
	})
}

func GenLongPutButterflyStrategy(ticker gopter.Gen) gopter.Gen {
	sp := GenShortPut(ticker)
	return sp.FlatMap(func(sp interface{}) gopter.Gen {
		lpl := GenShortPutWithStrike(ticker, gen.Float64Range(0, sp.(Put).Strike-1))
		lpu := GenShortPutWithStrike(ticker, gen.Float64Range(sp.(Put).Strike+1, MaxStrike))
		lp := lpl.FlatMap(func(lpl interface{}) gopter.Gen {
			return lpu.Map(func(lpu Put) Puts {
				return Puts{lpl.(Put), lpu}
			})
		}, reflect.TypeOf(Put{}))

		return gen.Struct(
			reflect.TypeOf(Strategy{}),
			map[string]gopter.Gen{
				"Ticker": ticker,
				"Sp":     gen.Const(Puts{sp.(Put), sp.(Put)}),
				"Lp":     lp,
				"Type":   gen.Const(PutButterfly),
				"Dir":    gen.Const(L)})
	}, reflect.TypeOf(Put{}))
}

func GenShortPutButterflyStrategy(ticker gopter.Gen) gopter.Gen {
	return GenLongPutButterflyStrategy(ticker).Map(func(s Strategy) Strategy {
		lp := s.Lp
		sp := s.Sp

		lp[0].Price, lp[1].Price, sp[0].Price, sp[1].Price = -sp[0].Price, -sp[0].Price, -lp[0].Price, -lp[1].Price
		lp[0].Strike, lp[1].Strike, sp[0].Strike, sp[1].Strike = sp[0].Strike, sp[1].Strike, lp[0].Strike, lp[1].Strike

		return Strategy{
			Ticker: s.Ticker,
			Lp:     lp,
			Sp:     sp,
			Type:   PutButterfly,
			Dir:    S}
	})
}

func GenLongJadeLizardStrategy(ticker gopter.Gen) gopter.Gen {
	return GenLongPut(ticker).FlatMap(func(lp interface{}) gopter.Gen {
		lcs := gen.Float64Range(lp.(Put).Strike+1, MaxStrike-1)
		return GenLongCallWithStrike(ticker, lcs).FlatMap(func(lc interface{}) gopter.Gen {
			scs := gen.Float64Range(lc.(Call).Strike+1, MaxStrike)
			return GenShortCallWithStrike(ticker, scs).FlatMap(func(sc interface{}) gopter.Gen {
				return gen.Struct(
					reflect.TypeOf(Strategy{}),
					map[string]gopter.Gen{
						"Ticker": ticker,
						"Lp":     gen.SliceOfN(1, gen.Const(lp.(Put))),
						"Lc":     gen.SliceOfN(1, gen.Const(lc)),
						"Sc":     gen.SliceOfN(1, gen.Const(sc.(Call))),
						"Type":   gen.Const(JadeLizard),
						"Dir":    gen.Const(L)})
			}, reflect.TypeOf(Call{}))
		}, reflect.TypeOf(Call{}))
	}, reflect.TypeOf(Put{}))
}

func GenShortJadeLizardStrategy(ticker gopter.Gen) gopter.Gen {
	return GenLongJadeLizardStrategy(ticker).Map(func(s Strategy) Strategy {
		lp, lc, sc := s.Lp[0], s.Lc[0], s.Sc[0]
		lp.Price, lc.Price, sc.Price = -lp.Price, -lc.Price, -sc.Price
		return Strategy{
			Ticker: s.Ticker,
			Sp:     Puts{lp},
			Sc:     Calls{lc},
			Lc:     Calls{sc},
			Type:   JadeLizard,
			Dir:    S}
	})
}

func GenLongNakedStockStrategy(ticker gopter.Gen) gopter.Gen {
	return gen.Struct(
		reflect.TypeOf(Strategy{}),
		map[string]gopter.Gen{
			"Ticker": ticker,
			"Stocks": gen.SliceOfN(1, GenLongStock(ticker)),
			"Type":   gen.Const(NakedStock),
			"Dir":    gen.Const(L)})
}

func GenShortNakedStockStrategy(ticker gopter.Gen) gopter.Gen {
	return gen.Struct(
		reflect.TypeOf(Strategy{}),
		map[string]gopter.Gen{
			"Ticker": ticker,
			"Stocks": gen.SliceOfN(1, GenShortStock(ticker)),
			"Type":   gen.Const(NakedStock),
			"Dir":    gen.Const(S)})
}

func GenLongNakedCallStrategy(ticker gopter.Gen) gopter.Gen {
	return gen.Struct(
		reflect.TypeOf(Strategy{}),
		map[string]gopter.Gen{
			"Ticker": ticker,
			"Lc":     gen.SliceOfN(1, GenLongCall(ticker)),
			"Type":   gen.Const(NakedCall),
			"Dir":    gen.Const(L)})
}

func GenShortNakedCallStrategy(ticker gopter.Gen) gopter.Gen {
	return gen.Struct(
		reflect.TypeOf(Strategy{}),
		map[string]gopter.Gen{
			"Ticker": ticker,
			"Sc":     gen.SliceOfN(1, GenShortCall(ticker)),
			"Type":   gen.Const(NakedCall),
			"Dir":    gen.Const(S)})
}

func GenLongNakedPutStrategy(ticker gopter.Gen) gopter.Gen {
	return gen.Struct(
		reflect.TypeOf(Strategy{}),
		map[string]gopter.Gen{
			"Ticker": ticker,
			"Lp":     gen.SliceOfN(1, GenLongPut(ticker)),
			"Type":   gen.Const(NakedPut),
			"Dir":    gen.Const(L)})
}

func GenShortNakedPutStrategy(ticker gopter.Gen) gopter.Gen {
	return gen.Struct(
		reflect.TypeOf(Strategy{}),
		map[string]gopter.Gen{
			"Ticker": ticker,
			"Sp":     gen.SliceOfN(1, GenShortPut(ticker)),
			"Type":   gen.Const(NakedPut),
			"Dir":    gen.Const(S)})
}

func GenLongCustomStrategy(ticker gopter.Gen) gopter.Gen {
	s := gen.Struct(
		reflect.TypeOf(Strategy{}),
		map[string]gopter.Gen{
			"Ticker": ticker,
			"Stocks": gen.SliceOfN(1, GenStock(ticker)),
			"Lp":     gen.SliceOfN(1, GenLongPut(ticker)),
			"Sp":     gen.SliceOfN(1, GenShortPut(ticker)),
			"Sc":     gen.SliceOfN(1, GenShortCall(ticker)),
			"Lc":     gen.SliceOfN(1, GenLongCall(ticker)),
			"Type":   gen.Const(Custom),
			"Dir":    gen.Const(L)})
	return s.SuchThat(func(s Strategy) bool {
		return s.Price() >= 0
	})
}

func GenShortCustomStrategy(ticker gopter.Gen) gopter.Gen {
	s := gen.Struct(
		reflect.TypeOf(Strategy{}),
		map[string]gopter.Gen{
			"Ticker": ticker,
			"Stocks": gen.SliceOfN(1, GenStock(ticker)),
			"Lp":     gen.SliceOfN(1, GenLongPut(ticker)),
			"Sp":     gen.SliceOfN(1, GenShortPut(ticker)),
			"Sc":     gen.SliceOfN(1, GenShortCall(ticker)),
			"Lc":     gen.SliceOfN(1, GenLongCall(ticker)),
			"Type":   gen.Const(Custom),
			"Dir":    gen.Const(S)})
	return s.SuchThat(func(s Strategy) bool {
		return s.Price() < 0
	})
}

func GenStrategy(ticker gopter.Gen) gopter.Gen {
	return gen.OneGenOf(
		GenLongPutSpreadStrategy(ticker),
		GenShortPutSpreadStrategy(ticker),
		GenEmptyStrategy(),
		GenLongStrangleStrategy(ticker),
		GenShortStrangleStrategy(ticker),
		GenLongStraddleStrategy(ticker),
		GenShortStraddleStrategy(ticker),
		GenLongCoveredCallStrategy(ticker),
		GenShortCoveredCallStrategy(ticker),
		GenLongCoveredPutStrategy(ticker),
		GenShortCoveredPutStrategy(ticker),
		GenLongIronCondorStrategy(ticker),
		GenShortIronCondorStrategy(ticker),
		GenLongCallButterflyStrategy(ticker),
		GenShortCallButterflyStrategy(ticker),
		GenLongPutButterflyStrategy(ticker),
		GenShortPutButterflyStrategy(ticker),
		GenLongJadeLizardStrategy(ticker),
		GenShortJadeLizardStrategy(ticker),
		GenLongNakedStockStrategy(ticker),
		GenShortNakedStockStrategy(ticker),
		GenLongNakedCallStrategy(ticker),
		GenShortNakedCallStrategy(ticker),
		GenLongNakedPutStrategy(ticker),
		GenShortNakedPutStrategy(ticker),
		GenLongCustomStrategy(ticker),
		GenShortCustomStrategy(ticker))
}
