package data

import (
	"errors"
	"sort"
)

type Type int

const (
	Spread        Type = iota
	Strangle      Type = iota
	Straddle      Type = iota
	CoveredCall   Type = iota
	CoveredPut    Type = iota
	IronCondor    Type = iota
	IronButterfly Type = iota
	CallButterfly Type = iota
	PutButterfly  Type = iota
	JadeLizard    Type = iota
	NakedStock    Type = iota
	NakedCall     Type = iota
	NakedPut      Type = iota
	Custom        Type = iota
	Empty         Type = iota
)

func (t Type) String() string {
	return []string{
		"Spread",
		"Strangle",
		"Straddle",
		"CoveredCall",
		"CoveredPut",
		"IronCondor",
		"IronButterfly",
		"CallButterfly",
		"PutButterfly",
		"JadeLizard",
		"NakedStock",
		"NakedCall",
		"NakedPut",
		"Custom",
		"Empty"}[t]
}

// Aggregates kinds and directions of strategies and returns the conditions under which they are valid.
func conditions() map[Type]func(*Strategy) (Direction, bool) {
	c := make(map[Type]func(*Strategy) (Direction, bool))

	c[Spread] = func(s *Strategy) (Direction, bool) {
		if !s.hasNStocks(0) {
			return None, false
		}

		put := s.hasNPuts(1, 1) && s.hasNCalls(0, 0)
		call := s.hasNPuts(0, 0) && s.hasNCalls(1, 1)

		if put {
			if s.Lp[0].Strike < s.Sp[0].Strike {
				return S, true
			} else if s.Lp[0].Strike > s.Sp[0].Strike {
				return L, true
			}
		} else if call {
			if s.Lc[0].Strike < s.Sc[0].Strike {
				return L, true
			} else if s.Lc[0].Strike > s.Sc[0].Strike {
				return S, true
			}
		}
		return None, false
	}

	c[Empty] = func(s *Strategy) (Direction, bool) {
		return None, s.empty()
	}

	c[Strangle] = func(s *Strategy) (Direction, bool) {
		if !s.hasNStocks(0) {
			return None, false
		}
		short := s.hasNPuts(0, 1) && s.hasNCalls(1, 0)
		long := s.hasNPuts(1, 0) && s.hasNCalls(0, 1)

		if short && s.Sp[0].Strike < s.Sc[0].Strike {
			return S, true
		} else if long && s.Lp[0].Strike < s.Lc[0].Strike {
			return L, true
		}
		return None, false
	}

	c[Straddle] = func(s *Strategy) (Direction, bool) {
		if !s.hasNStocks(0) {
			return None, false
		}
		short := s.hasNPuts(0, 1) && s.hasNCalls(1, 0)
		long := s.hasNPuts(1, 0) && s.hasNCalls(0, 1)

		if short && s.Sp[0].Strike == s.Sc[0].Strike {
			return S, true
		} else if long && s.Lp[0].Strike == s.Lc[0].Strike {
			return L, true
		}
		return None, false
	}

	c[CoveredCall] = func(s *Strategy) (Direction, bool) {
		if !s.hasNStocks(1) {
			return None, false
		}

		if s.Stocks[0].Shares != 100 {
			return None, false
		}

		if !s.hasNPuts(0, 0) {
			return None, false
		}

		if s.Stocks[0].Dir() == L && s.hasNCalls(1, 0) {
			return S, true
		} else if s.Stocks[0].Dir() == S && s.hasNCalls(0, 1) {
			return L, true
		}

		return None, false
	}

	c[CoveredPut] = func(s *Strategy) (Direction, bool) {
		if !s.hasNStocks(1) {
			return None, false
		}

		if s.Stocks[0].Shares != 100 {
			return None, false
		}

		if !s.hasNCalls(0, 0) {
			return None, false
		}

		if s.Stocks[0].Dir() == L && s.hasNPuts(1, 0) {
			return L, true
		} else if s.Stocks[0].Dir() == S && s.hasNPuts(0, 1) {
			return S, true
		}
		return None, false
	}

	c[IronCondor] = func(s *Strategy) (Direction, bool) {
		if !s.hasNStocks(0) {
			return None, false
		}
		if !(s.hasNPuts(1, 1) && s.hasNCalls(1, 1)) {
			return None, false
		}
		short := (s.Lp[0].Strike < s.Sp[0].Strike) && (s.Sp[0].Strike < s.Sc[0].Strike) && (s.Sc[0].Strike < s.Lc[0].Strike)
		long := (s.Sp[0].Strike < s.Lp[0].Strike) && (s.Lp[0].Strike < s.Lc[0].Strike) && (s.Lc[0].Strike < s.Sc[0].Strike)

		if short {
			return S, true
		} else if long {
			return L, true
		}
		return None, false
	}

	c[IronButterfly] = func(s *Strategy) (Direction, bool) {
		if !s.hasNStocks(0) {
			return None, false
		}
		if !(s.hasNPuts(1, 1) && s.hasNCalls(1, 1)) {
			return None, false
		}
		short := (s.Lp[0].Strike < s.Sp[0].Strike) && (s.Sp[0].Strike == s.Sc[0].Strike) && (s.Sc[0].Strike < s.Lc[0].Strike)
		long := (s.Sp[0].Strike < s.Lp[0].Strike) && (s.Lp[0].Strike == s.Lc[0].Strike) && (s.Lc[0].Strike < s.Sc[0].Strike)

		if short {
			return S, true
		} else if long {
			return L, true
		}
		return None, false
	}

	c[CallButterfly] = func(s *Strategy) (Direction, bool) {
		if !s.hasNStocks(0) {
			return None, false
		}
		if !s.hasNPuts(0, 0) {
			return None, false
		}
		if !s.hasNCalls(2, 2) {
			return None, false
		}

		long := s.Sc[0].Strike == s.Sc[1].Strike &&
			s.Lc[0].Strike < s.Sc[0].Strike &&
			s.Sc[1].Strike < s.Lc[1].Strike
		short := s.Lc[0].Strike == s.Lc[1].Strike &&
			s.Sc[0].Strike < s.Lc[0].Strike &&
			s.Lc[1].Strike < s.Sc[1].Strike

		if long {
			return L, true
		} else if short {
			return S, true
		}

		return None, false
	}

	c[PutButterfly] = func(s *Strategy) (Direction, bool) {
		if !s.hasNStocks(0) {
			return None, false
		}
		if !s.hasNCalls(0, 0) {
			return None, false
		}
		if !s.hasNPuts(2, 2) {
			return None, false
		}

		long := s.Sp[0].Strike == s.Sp[1].Strike &&
			s.Lp[0].Strike < s.Sp[0].Strike &&
			s.Sp[1].Strike < s.Lp[1].Strike
		short := s.Lp[0].Strike == s.Lp[1].Strike &&
			s.Sp[0].Strike < s.Lp[0].Strike &&
			s.Lp[1].Strike < s.Sp[1].Strike

		if long {
			return L, true
		} else if short {
			return S, true
		}
		return None, false
	}

	c[JadeLizard] = func(s *Strategy) (Direction, bool) {
		if !s.hasNStocks(0) {
			return None, false
		}
		if !s.hasNCalls(1, 1) {
			return None, false
		}

		long := s.hasNPuts(1, 0) && s.Lc[0].Strike < s.Sc[0].Strike && s.Lp[0].Strike <= s.Lc[0].Strike
		short := s.hasNPuts(0, 1) && s.Sc[0].Strike < s.Lc[0].Strike && s.Sp[0].Strike <= s.Sc[0].Strike

		if long {
			return L, true
		} else if short {
			return S, true
		}
		return None, false
	}

	c[NakedStock] = func(s *Strategy) (Direction, bool) {
		if s.hasAnyOptions() {
			return None, false
		}
		if s.hasNStocks(1) {
			return s.Stocks[0].Dir(), true
		}
		return None, false
	}

	c[NakedCall] = func(s *Strategy) (Direction, bool) {
		if !s.hasNStocks(0) {
			return None, false
		}
		put := s.hasNPuts(0, 0)
		long := len(s.Lc) == 1
		short := len(s.Sc) == 1

		if put && long {
			return L, true
		} else if put && short {
			return S, true
		}
		return None, false
	}

	c[NakedPut] = func(s *Strategy) (Direction, bool) {
		if !s.hasNStocks(0) {
			return None, false
		}
		if !(s.hasNCalls(0, 0)) {
			return None, false
		}
		long := len(s.Lp) == 1 && len(s.Sp) == 0
		short := len(s.Sp) == 1 && len(s.Lp) == 0

		if long {
			return L, true
		} else if short {
			return S, true
		}
		return None, false
	}

	c[Custom] = func(s *Strategy) (Direction, bool) {
		for k, v := range c {
			if k == Custom {
				continue
			}
			if _, ok := v(s); ok {
				return None, false
			}
		}

		// Compute direction
		price := s.Price()

		if price > 0 {
			return L, true
		}
		return S, false
	}

	return c
}

// Container for all strategy kinds.
// Note it is not 'strategy safe' to construct this via the standard constructor.
// Use the NewStrategy method to allow for dynamic inference of strategy Type and direction.
type Strategy struct {
	Ticker string
	Stocks Stocks
	Lp     Puts
	Sp     Puts
	Sc     Calls
	Lc     Calls
	Type   Type
	Dir    Direction
}

// Determines the Type of a strategy. Defaults to Custom if there are no other matches.
func (s *Strategy) CheckKind() (Type, Direction) {
	ks := conditions()
	for k, v := range ks {
		d, ok := v(s)
		if ok {
			return k, d
		}
	}

	// In case Type doesn't get selected above, return Custom.
	d, _ := ks[Custom](s)
	return Custom, d
}

func (s *Strategy) Price() float64 {
	return s.PriceOptions() + s.Stocks.Price()
}

func (s *Strategy) PriceOptions() (price float64) {
	return s.Lp.Price() + s.Sp.Price() + s.Sc.Price() + s.Lc.Price()
}

func (s *Strategy) CountOptions() (count int) {
	return len(s.Lp) + len(s.Sp) + len(s.Sc) + len(s.Lc)
}

func (s *Strategy) hasNCalls(sc, lc int) bool {
	return len(s.Sc) == sc && len(s.Lc) == lc
}

func (s *Strategy) hasNPuts(lp, sp int) bool {
	return len(s.Lp) == lp && len(s.Sp) == sp
}

func (s *Strategy) hasNStocks(ss int) bool {
	return len(s.Stocks) == ss
}

func (s *Strategy) hasAnyOptions() bool {
	return len(s.Lc) == 1 || len(s.Sc) == 1 || len(s.Sp) == 1 || len(s.Lp) == 1
}

func (s *Strategy) hasAny(ss, lp, sp, sc, lc bool) bool {
	return len(s.Stocks) > 0 || len(s.Lp) > 0 || len(s.Sp) > 0 || len(s.Sc) > 0 || len(s.Lc) > 0
}

func (s *Strategy) empty() bool {
	if len(s.Stocks) == 0 && len(s.Lp) == 0 && len(s.Sp) == 0 && len(s.Sc) == 0 && len(s.Lc) == 0 {
		return true
	} else {
		return false
	}
}

// Create a new strategy from arrays of Stocks, Puts, and Calls
func NewStrategy(ss Stocks, ps Puts, cs Calls) (Strategy, error) {

	s := Strategy{}

	ticker, ok := parseTicker(ss, ps, cs)
	if !ok {
		return s, nil
	}
	s.Ticker = ticker

	e := parseCalls(&s, cs)
	if e != nil {
		return Strategy{}, e
	}

	e = parsePuts(&s, ps)
	if e != nil {
		return Strategy{}, e
	}

	e = parseStocks(&s, ss)
	if e != nil {
		return Strategy{}, e
	}

	kind, dir := s.CheckKind()
	s.Type = kind
	s.Dir = dir

	return s, nil
}

// Returns the first ticker found in a series of assets. If not found will return empty.
func parseTicker(ss Stocks, ps Puts, cs Calls) (string, bool) {

	lss, lps, lcs := len(ss), len(ps), len(cs)
	if lss == 0 && lps == 0 && lcs == 0 {
		return "", false
	} else if lps == 0 && lcs == 0 {
		return ss[0].Ticker, true
	} else if lps == 0 {
		return cs[0].Underlying.Ticker, true
	} else {
		return ps[0].Underlying.Ticker, true
	}

}

// Inserts stocks into strategy object. If any stocks do not match ticker, an error will be thrown.
func parseStocks(s *Strategy, ss Stocks) error {
	sort.Sort(ss)

	for _, st := range ss {
		if st.Ticker != s.Ticker {
			return errors.New("stocks and options must have same ticker")
		}
		s.Stocks = append(s.Stocks, st)
	}
	return nil
}

// Inserts calls into strategy object. If any calls do not match ticker, an error will be thrown.
func parseCalls(s *Strategy, cs Calls) error {
	sort.Sort(cs)
	for _, o := range cs {
		// Throw error if tickers do not match
		if o.Underlying.Ticker != s.Ticker {
			return errors.New("all options must have same underlying ticker")
		}

		// Place options into correct location in set.
		if o.Dir() == L {
			s.Lc = append(s.Lc, o)
		} else {
			s.Sc = append(s.Sc, o)
		}
	}
	return nil
}

// Inserts puts into strategy object. If any puts do not match ticker, an error will be thrown.
func parsePuts(s *Strategy, ps Puts) error {
	sort.Sort(ps)
	for _, o := range ps {
		// Throw error if tickers do not match
		if o.Underlying.Ticker != s.Ticker {
			return errors.New("all options must have same underlying ticker")
		}

		// Place options into correct location in set.
		if o.Dir() == L {
			s.Lp = append(s.Lp, o)
		} else {
			s.Sp = append(s.Sp, o)
		}
	}
	return nil
}
