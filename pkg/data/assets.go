package data

/*
	ASSET
*/

type Asset interface {
	Empty() bool
	Dir() Direction
}

type Assets []Asset

/*
	STOCK
*/

type Stocks []Stock

type Stock struct {
	Ticker string
	Price  float64
	Shares int
}

func (s Stock) Dir() Direction {
	if s.Price < 0 {
		return S
	}
	return L
}

func (s Stock) Empty() bool {
	return s.Ticker == ""
}

func (ss Stocks) Len() int {
	return len(ss)
}

func (ss Stocks) Less(i, j int) bool {
	return ss[i].Price < ss[j].Price

}

func (ss Stocks) Swap(i, j int) {
	ss[i], ss[j] = ss[j], ss[i]
}

func (ss *Stocks) Price() (price float64) {
	for _, st := range *ss {
		price += st.Price
	}
	return price
}

func (ss *Stocks) Shares() (shares int) {
	for _, st := range *ss {
		shares += st.Shares
	}
	return shares
}

/*
	PUTS
*/
type Put struct {
	Underlying Stock
	Price      float64
	Strike     float64
}

type Puts []Put

func (p Put) Empty() bool {
	return p.Underlying.Empty()
}

func (p Put) Dir() Direction {
	if p.Price < 0 {
		return S
	}
	return L
}

func (ps Puts) Len() int {
	return len(ps)
}

func (ps Puts) Less(i, j int) bool {
	if ps[i].Underlying.Ticker != ps[j].Underlying.Ticker {
		return ps[i].Underlying.Ticker < ps[j].Underlying.Ticker
	} else {
		return ps[i].Strike < ps[j].Strike
	}
}

func (ps Puts) Swap(i, j int) {
	ps[i], ps[j] = ps[j], ps[i]
}

func (ps Puts) Price() (price float64) {
	for _, p := range ps {
		price += p.Price
	}
	return price
}

/*
	CALL
*/
type Call struct {
	Underlying Stock
	Price      float64
	Strike     float64
}

type Calls []Call

func (c Call) Empty() bool {
	return c.Underlying.Empty()
}

func (c Call) Dir() Direction {
	if c.Price < 0 {
		return S
	}
	return L
}

func (cs Calls) Len() int {
	return len(cs)
}

func (cs Calls) Less(i, j int) bool {
	if cs[i].Underlying.Ticker != cs[j].Underlying.Ticker {
		return cs[i].Underlying.Ticker < cs[j].Underlying.Ticker
	} else {
		return cs[i].Strike < cs[j].Strike
	}
}

func (cs Calls) Swap(i, j int) {
	cs[i], cs[j] = cs[j], cs[i]
}

func (cs Calls) Price() (price float64) {
	for _, c := range cs {
		price += c.Price
	}
	return price
}

/*
	DIRECTION
*/
type Direction int

const (
	L    Direction = iota
	S    Direction = iota
	None Direction = iota
)

func (d Direction) String() string {
	return []string{"Long", "Short", "None"}[d]
}
