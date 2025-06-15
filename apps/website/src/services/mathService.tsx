/**
 * Calculates a new latitude and longitude given a starting point,
 * heading, and distance, using a spherical Earth model.
 *
 * @param lat1 Initial latitude in degrees.
 * @param lon1 Initial longitude in degrees.
 * @param heading Bearing (angle) in degrees, clockwise from North (0 = North, 90 = East).
 * @param distance Distance in meters.
 * @returns A tuple [newLatitude, newLongitude] in degrees.
 */
function calculateNewPosition(
    lat1: number,
    lon1: number,
    heading: number,
    distance: number
): [number, number] {
    // Earth's mean radius in meters
    const R = 6371e3; // 6,371,000 meters

    // Convert latitude, longitude, and heading to radians
    const lat1Rad = lat1 * Math.PI / 180;
    const lon1Rad = lon1 * Math.PI / 180;
    const headingRad = heading * Math.PI / 180;

    // Compute the new latitude
    const lat2Rad = Math.asin(
        Math.sin(lat1Rad) * Math.cos(distance / R) +
        Math.cos(lat1Rad) * Math.sin(distance / R) * Math.cos(headingRad)
    );

    // Compute the new longitude
    const lon2Rad =
        lon1Rad +
        Math.atan2(
            Math.sin(headingRad) * Math.sin(distance / R) * Math.cos(lat1Rad),
            Math.cos(distance / R) - Math.sin(lat1Rad) * Math.sin(lat2Rad)
        );

    // Convert the results back to degrees
    const lat2Deg = lat2Rad * 180 / Math.PI;
    const lon2Deg = lon2Rad * 180 / Math.PI;

    return [lat2Deg, lon2Deg];
}


export const MathService = {
    calculateNewPosition
}