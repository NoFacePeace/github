package indicator

import (
	"fmt"
	"time"

	"github.com/NoFacePeace/github/repositories/go/external/tencent/finance"
)

type Point struct {
	Price float64
	Date  time.Time
}

func AllPrice(code string) ([]Point, error) {
	ps, err := finance.GetAllKline(code)
	if err != nil {
		return nil, fmt.Errorf("get points %v: [%w]", code, err)
	}
	ret := []Point{}
	for _, v := range ps {
		ret = append(ret, Point{
			Price: v.Last,
			Date:  v.Date,
		})
	}
	return ret, nil
}

func Price(code string) ([]Point, error) {
	ps, err := finance.GetKline(code)
	if err != nil {
		return nil, fmt.Errorf("get points %v: [%w]", code, err)
	}
	ret := []Point{}
	for _, v := range ps {
		ret = append(ret, Point{
			Price: v.Last,
			Date:  v.Date,
		})
	}
	return ret, nil
}

func SMA(ps []Point, window int) []Point {
	ret := []Point{}
	sum := 0.0
	for k, v := range ps {
		if k < window-1 {
			sum += v.Price
			ret = append(ret, Point{
				Price: 0,
				Date:  v.Date,
			})
			continue
		}
		if k == window-1 {
			sum += v.Price
			ret = append(ret, Point{
				Price: sum,
				Date:  v.Date,
			})
			continue
		}
		sum += v.Price
		sum -= ps[k-window].Price
		ret = append(ret, Point{
			Price: sum,
			Date:  v.Date,
		})
	}
	for i := 0; i < len(ret); i++ {
		ret[i].Price /= float64(window)
	}
	return ret
}

func GoldenCross(ps, short, long []Point) []Point {
	n := len(short)
	ret := []Point{}
	for i := 0; i < n-1; i++ {
		if short[i].Price == 0 || long[i].Price == 0 {
			continue
		}
		if short[i].Price < long[i].Price && short[i+1].Price >= long[i+1].Price {
			ret = append(ret, ps[i+1])
		}
		if len(ret) > 0 && short[i].Price > long[i].Price && short[i+1].Price <= long[i+1].Price {
			ret = append(ret, ps[i+1])
		}
	}
	return ret
}

func GoldenCrossMax(ps, short, long []Point) []Point {
	n := len(short)
	ret := []Point{}
	mx := 0.0
	for i := 0; i < n-1; i++ {
		if short[i].Price == 0 || long[i].Price == 0 {
			continue
		}
		if short[i].Price < long[i].Price && short[i+1].Price >= long[i+1].Price {
			ret = append(ret, ps[i+1])
			mx = 0.0
			continue
		}
		mx = max(mx, ps[i].Price)
		if len(ret) > 0 && short[i].Price > long[i].Price && short[i+1].Price <= long[i+1].Price {
			ret = append(ret, Point{
				Price: max(mx, ps[i+1].Price),
				Date:  ps[i+1].Date,
			})
		}
	}
	return ret
}

func DeadCrossMax(ps, short, long []Point) []Point {
	n := len(short)
	ret := []Point{}
	mx := 0.0
	for i := 0; i < n-1; i++ {
		if short[i].Price == 0 || long[i].Price == 0 {
			continue
		}
		if short[i].Price > long[i].Price && short[i+1].Price <= long[i+1].Price {
			ret = append(ret, ps[i+1])
			mx = 0.0
			continue
		}
		mx = max(mx, ps[i].Price)
		if len(ret) > 0 && short[i].Price < long[i].Price && short[i+1].Price >= long[i+1].Price {
			ret = append(ret, Point{
				Price: max(mx, ps[i+1].Price),
				Date:  ps[i+1].Date,
			})
		}
	}
	return ret
}

func GoldenCrossLast(ps, short, long []Point, last int) []Point {
	n := len(short)
	ret := []Point{}
	for i := 0; i < n-1; i++ {
		if short[i].Price == 0 || long[i].Price == 0 {
			continue
		}
		if short[i].Price >= long[i].Price {
			continue
		}
		if short[i+1].Price < long[i+1].Price {
			continue
		}
		ret = append(ret, ps[i+1])
		if i+1+last >= n {
			continue
		}
		mx := ps[i+2]
		for j := i + 2; j <= i+1+last; j++ {
			if ps[j].Price > mx.Price {
				mx = ps[j]
			}
		}
		ret = append(ret, mx)
	}
	return ret
}
func DeadCrossLast(ps, short, long []Point, last int) []Point {
	n := len(short)
	ret := []Point{}
	for i := 0; i < n-1; i++ {
		if short[i].Price == 0 || long[i].Price == 0 {
			continue
		}
		if short[i].Price <= long[i].Price {
			continue
		}
		if short[i+1].Price > long[i+1].Price {
			continue
		}
		ret = append(ret, ps[i+1])
		if i+1+last >= n {
			continue
		}
		mx := ps[i+2]
		for j := i + 2; j <= i+1+last; j++ {
			if ps[j].Price > mx.Price {
				mx = ps[j]
			}
		}
		ret = append(ret, mx)
	}
	return ret
}

func Win(ps []Point) float64 {
	win := 0.0
	lose := 0.0
	for i := 0; i < len(ps)-1; i += 2 {
		if ps[i].Price < ps[i+1].Price {
			win += 1
		} else {
			lose += 1
		}
	}
	return win / (win + lose) * 100
}
func WinPercent(ps []Point, percent float64) float64 {
	win := 0.0
	lose := 0.0
	for i := 0; i < len(ps)-1; i += 2 {
		if ps[i].Price*(1+percent/100) <= ps[i+1].Price {
			win += 1
		} else {
			lose += 1
		}
	}
	return win / (win + lose) * 100
}

func Earn(ps []Point) float64 {
	sum := 0.0
	for i := 0; i < len(ps)-1; i += 2 {
		sum += (ps[i+1].Price - ps[i].Price) / ps[i].Price
	}
	return sum * 100
}

func Money(ps []Point) float64 {
	sum := 0.0
	for i := 0; i < len(ps)-1; i += 2 {
		sum += (ps[i+1].Price - ps[i].Price)
	}
	return sum
}

func SMAGoldenCrossBest(ps []Point, mn, mx int) (int, int, float64, []Point) {
	bestWin := 0.0
	bestPs := []Point{}
	bestShort := 0
	bestLong := 0
	for i := mn; i <= mx; i++ {
		for j := i + 1; j <= mx; j++ {
			short := SMA(ps, i)
			long := SMA(ps, j)
			cross := GoldenCross(ps, short, long)
			win := Win(cross)
			if win > bestWin {
				bestWin = win
				bestPs = cross
				bestShort = i
				bestLong = j
			}
		}
	}
	return bestShort, bestLong, bestWin, bestPs
}

func SMACrossMaxBest(ps []Point, mn, mx int) (int, int, float64, []Point) {
	bestWin := 0.0
	bestPs := []Point{}
	bestShort := 0
	bestLong := 0
	for i := mn; i <= mx; i++ {
		for j := i + 1; j <= mx; j++ {
			short := SMA(ps, i)
			long := SMA(ps, j)
			cross := GoldenCrossMax(ps, short, long)
			win := Win(cross)
			if win > bestWin {
				bestWin = win
				bestPs = cross
				bestShort = i
				bestLong = j
			}
		}
	}
	return bestShort, bestLong, bestWin, bestPs
}

func SMACrossMaxPercentBest(ps []Point, mn, mx int, percent float64) (int, int, float64, []Point) {
	bestWin := 0.0
	bestPs := []Point{}
	bestShort := 0
	bestLong := 0
	for i := mn; i <= mx; i++ {
		for j := i + 1; j <= mx; j++ {
			short := SMA(ps, i)
			long := SMA(ps, j)
			cross := GoldenCrossMax(ps, short, long)
			// cross := DeadCrossMax(ps, short, long)
			win := WinPercent(cross, percent)
			if win > bestWin {
				bestWin = win
				bestPs = cross
				bestShort = i
				bestLong = j
			}
		}
	}
	return bestShort, bestLong, bestWin, bestPs
}

func SMAGoldenCrossLastPercent(ps []Point, mn, mx, last int, percent float64) (int, int, float64, []Point) {
	bestWin := 0.0
	bestPs := []Point{}
	bestShort := 0
	bestLong := 0
	for i := mn; i <= mx; i++ {
		for j := i + 1; j <= mx; j++ {
			short := SMA(ps, i)
			long := SMA(ps, j)
			cross := GoldenCrossLast(ps, short, long, last)
			win := WinPercent(cross, percent)
			if win > bestWin {
				bestWin = win
				bestPs = cross
				bestShort = i
				bestLong = j
			}
		}
	}
	return bestShort, bestLong, bestWin, bestPs
}

func SMADeadCrossLastPercent(ps []Point, mn, mx, last int, percent float64) (int, int, float64, []Point) {
	bestWin := 0.0
	bestPs := []Point{}
	bestShort := 0
	bestLong := 0
	for i := mn; i <= mx; i++ {
		for j := i + 1; j <= mx; j++ {
			short := SMA(ps, i)
			long := SMA(ps, j)
			cross := DeadCrossLast(ps, short, long, last)
			win := WinPercent(cross, percent)
			if win > bestWin {
				bestWin = win
				bestPs = cross
				bestShort = i
				bestLong = j
			}
		}
	}
	return bestShort, bestLong, bestWin, bestPs
}
