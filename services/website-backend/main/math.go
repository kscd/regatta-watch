package main

import (
	"errors"
	"fmt"
	"math"
)

type lineSegment struct {
	slope      float64
	offset     float64 // is x value if isVertical is true
	isVertical bool
	start      float64 // x start if not vertical, y start if vertical
	end        float64 // x end if not vertical, y end if vertical
}

func (l *lineSegment) y(x float64) (float64, error) {
	if l.isVertical {
		return 0, errors.New("line segment is vertical")
	}
	return l.slope*x + l.offset, nil
}

func newLineSegment(x0, y0, x1, y1 float64) lineSegment {
	if x0 == x1 {
		return lineSegment{
			isVertical: true,
			offset:     x0,
			start:      y0,
			end:        y1,
		}
	}

	slope := (y1 - y0) / (x1 - x0)
	offset := y1 - slope*x1
	return lineSegment{
		slope:  slope,
		offset: offset,
		start:  x0,
		end:    x1,
	}
}

func isIntersecting(lineSegment1, lineSegment2 lineSegment) (bool, error) {
	// edge case: both line segments vertical

	// line segment lines are equal
	if lineSegment1.isVertical && lineSegment2.isVertical && lineSegment1.offset == lineSegment2.offset {
		ls1Low := math.Min(lineSegment1.start, lineSegment1.end)
		ls1High := math.Max(lineSegment1.start, lineSegment1.end)
		ls2Low := math.Min(lineSegment2.start, lineSegment2.end)
		ls2High := math.Max(lineSegment2.start, lineSegment2.end)
		if ls1High < ls2Low || ls2High < ls1Low {
			return false, nil
		}
		return true, nil
	}

	// line segment lines are parallel
	if lineSegment1.isVertical && lineSegment2.isVertical && lineSegment1.offset != lineSegment2.offset {
		return false, nil
	}

	// edge case: line segment1 vertical
	if lineSegment1.isVertical && !lineSegment2.isVertical {
		ls1Low := math.Min(lineSegment1.start, lineSegment1.end)
		ls1High := math.Max(lineSegment1.start, lineSegment1.end)
		ls2Low := math.Min(lineSegment2.start, lineSegment2.end)
		ls2High := math.Max(lineSegment2.start, lineSegment2.end)

		if ls2Low > lineSegment1.offset || ls2High < lineSegment1.offset {
			return false, nil
		}

		y, err := lineSegment2.y(lineSegment1.offset)
		if err != nil {
			return false, fmt.Errorf("error in calculating intersection: %w", err)
		}

		if y < ls1Low || y > ls1High {
			return false, nil
		}

		return true, nil
	}

	// edge case: line segment2 vertical
	if !lineSegment1.isVertical && lineSegment2.isVertical {
		ls1Low := math.Min(lineSegment1.start, lineSegment1.end)
		ls1High := math.Max(lineSegment1.start, lineSegment1.end)
		ls2Low := math.Min(lineSegment2.start, lineSegment2.end)
		ls2High := math.Max(lineSegment2.start, lineSegment2.end)

		if ls1Low > lineSegment2.offset || ls1High < lineSegment2.offset {
			return false, nil
		}

		y, err := lineSegment1.y(lineSegment2.offset)
		if err != nil {
			return false, fmt.Errorf("error in calculating intersection: %w", err)
		}

		if y < ls2Low || y > ls2High {
			return false, nil
		}

		return true, nil
	}

	// main cases

	// line segment lines are equal
	if lineSegment1.slope == lineSegment2.slope && lineSegment1.offset == lineSegment2.offset {
		ls1Low := math.Min(lineSegment1.start, lineSegment1.end)
		ls1High := math.Max(lineSegment1.start, lineSegment1.end)
		ls2Low := math.Min(lineSegment2.start, lineSegment2.end)
		ls2High := math.Max(lineSegment2.start, lineSegment2.end)
		if ls1High < ls2Low || ls2High < ls1Low {
			return false, nil
		}
		return true, nil
	}

	// line segment lines are parallel
	if lineSegment1.slope == lineSegment2.slope && lineSegment1.offset != lineSegment2.offset {
		return false, nil
	}

	// main case
	xInt := -(lineSegment2.offset - lineSegment1.offset) / (lineSegment2.slope - lineSegment1.slope)
	ls1Low := math.Min(lineSegment1.start, lineSegment1.end)
	ls1High := math.Max(lineSegment1.start, lineSegment1.end)
	ls2Low := math.Min(lineSegment2.start, lineSegment2.end)
	ls2High := math.Max(lineSegment2.start, lineSegment2.end)
	if xInt < ls1Low || xInt < ls2Low || xInt > ls1High || xInt > ls2High {
		return false, nil
	}

	return true, nil
}

func calculateNewPosition(lat1, lon1, heading, distance float64) (float64, float64) {
	// Earth's radius in meters
	const R = 6371e3

	// Convert latitude, longitude, and heading to radians
	lat1 = lat1 * math.Pi / 180
	lon1 = lon1 * math.Pi / 180
	heading = heading * math.Pi / 180

	// Compute the new latitude
	lat2 := math.Asin(math.Sin(lat1)*math.Cos(distance/R) +
		math.Cos(lat1)*math.Sin(distance/R)*math.Cos(heading))

	// Compute the new longitude
	lon2 := lon1 + math.Atan2(math.Sin(heading)*math.Sin(distance/R)*math.Cos(lat1),
		math.Cos(distance/R)-math.Sin(lat1)*math.Sin(lat2))

	// Convert the results back to degrees
	lat2 = lat2 * 180 / math.Pi
	lon2 = lon2 * 180 / math.Pi

	return lat2, lon2
}
