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

// 0 degrees are north and the heading rotates clockwise
func calculateHeading(oldLatitude, oldLongitude, newLatitude, newLongitude float64) float64 {
	oldLatitudeRadians := oldLatitude * math.Pi / 180
	oldLongitudeRadians := oldLongitude * math.Pi / 180
	newLatitudeRadians := newLatitude * math.Pi / 180
	newLongitudeRadians := newLongitude * math.Pi / 180

	X := math.Cos(newLatitudeRadians) * math.Sin(newLongitudeRadians-oldLongitudeRadians)
	Y := math.Cos(oldLatitudeRadians)*math.Sin(newLatitudeRadians) - math.Sin(oldLatitudeRadians)*math.Cos(newLatitudeRadians)*math.Cos(newLongitudeRadians-oldLongitudeRadians)
	return math.Atan2(X, Y) * 180 / math.Pi
}

// uses euclidean geometry
func calculateDistanceInNM(oldLatitude, oldLongitude, newLatitude, newLongitude float64) float64 {
	cosLatitude := (math.Cos(newLatitude) + math.Cos(oldLatitude)) / 2
	deltaN := (newLatitude - oldLatitude) * nauticalMilesPerDegree
	deltaW := (newLongitude - oldLongitude) * nauticalMilesPerDegree * cosLatitude
	return math.Sqrt(deltaN*deltaN + deltaW*deltaW) // nautical miles
}

func calculateIfBuoysPassed(buoys []buoy, oldPosition, newPosition *Position) ([]bool, error) {
	var isPassed []bool
	for i := range buoys {
		lat1, lon1 := calculateNewPosition(buoys[i].Latitude, buoys[i].Longitude, buoys[i].PassAngle+180, buoys[i].ToleranceInMeters)
		lat2, lon2 := calculateNewPosition(buoys[i].Latitude, buoys[i].Longitude, buoys[i].PassAngle, buoyFarOffPointDistance)

		buoyLS := newLineSegment(lat1, lon1, lat2, lon2)
		boatLS := newLineSegment(oldPosition.Latitude, oldPosition.Longitude, newPosition.Latitude, newPosition.Longitude)

		isBuoyPassed, err := isIntersecting(buoyLS, boatLS)
		if err != nil {
			return nil, fmt.Errorf("error checking if buoys passed: %v", err)
		}
		if !isBuoyPassed {
			isPassed = append(isPassed, false)
			continue
		}
		isBuoyPassed = isPassDirectionCorrect(
			lat1, lon1,
			lat2, lon2,
			oldPosition.Latitude, oldPosition.Longitude,
			newPosition.Latitude, newPosition.Longitude,
			buoys[i].IsPassDirectionClockwise,
		)
		isPassed = append(isPassed, isBuoyPassed)
	}
	return isPassed, nil
}

func isPassDirectionCorrect(buoyLat1, buoyLon1, buoyLat2, buoyLon2, boatLat1, boatLon1, boatLat2, boatLon2 float64, isPassDirectionClockwise bool) bool {
	// For the three vectors buoy to buoy far off (vec a), buoy to boat old
	// (vec b), buoy to boat new (vec c) calculate the polar angle and do
	// some comparisons to figure out which way the boat rotated around the
	// buoy.

	vecAY := boatLat1 - buoyLat1
	vecAX := (boatLon1 - buoyLon1) * math.Cos((boatLat1+buoyLat1)/2/180*math.Pi) // compensate for non-square coordinate system
	vecBY := buoyLat2 - buoyLat1
	vecBX := (buoyLon2 - buoyLon1) * math.Cos((buoyLat2+buoyLat1)/2/180*math.Pi)
	vecCY := boatLat2 - buoyLat1
	vecCX := (boatLon2 - buoyLon1) * math.Cos((boatLat2+buoyLat1)/2/180*math.Pi)

	alpha := math.Atan2(vecAY, vecAX) * 180 / math.Pi
	beta := math.Atan2(vecBY, vecBX) * 180 / math.Pi
	gamma := math.Atan2(vecCY, vecCX) * 180 / math.Pi
	alpha = math.Mod(alpha-beta, 360)
	gamma = math.Mod(gamma-beta, 360)

	if alpha >= 0 && gamma <= 0 {
		// clockwise rotation
		if isPassDirectionClockwise {
			return true
		}
		return false
	}

	if alpha <= 0 && gamma >= 0 {
		// anti-clockwise rotation
		if isPassDirectionClockwise {
			return false
		}
		return true
	}

	// something went wrong, there was never an intersection, return false
	return false
}
