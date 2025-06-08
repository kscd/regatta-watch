import React from "react";
import {MapContainer, TileLayer} from "react-leaflet";

export const Map2: React.FC = () => {
    const initialZoom = 15;

    const bounds: L.LatLngBoundsExpression = [
        [53.5672-0.015, 10.006-0.015],
        [53.5672+0.015, 10.006+0.015]
    ];

    return(
        <MapContainer
            center={[53.5672, 10.006]}
            zoom={initialZoom}
            style={{ height: '100%', width: '100%' }}
            minZoom={initialZoom}
            maxZoom={18}
            maxBounds={bounds}
            maxBoundsViscosity={1.0}
        >
            <TileLayer
                attribution='&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a> contributors'
                url="https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png"
            />
        </MapContainer>
    )
}