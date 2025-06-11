import React from "react";
import L from 'leaflet';
import {Circle, MapContainer, Polyline, TileLayer} from "react-leaflet";
import {PearlChain, Position} from "../services/boatService.tsx";
import {Buoy} from "../services/buoyService.tsx";

type MapProps = {buoys: Buoy[], boatPositions: Position[], pearlChains: PearlChain[]};

export const Map: React.FC<MapProps> = ({buoys, boatPositions, pearlChains}) => {
    const initialZoom = 15;

    const bounds: L.LatLngBoundsExpression = [
        [53.5677 - 0.015, 10.006 - 0.015],
        [53.5677 + 0.015, 10.006 + 0.015]
    ];

    const positions: L.LatLngExpression[] = [];
    const polyLines: L.LatLngExpression[][] = [];
    for (const [index, position] of boatPositions.entries()) {
        positions.push([position.latitude, position.longitude]);
        polyLines.push([[position.latitude, position.longitude]]);

        if (pearlChains[index].positions && Array.isArray(pearlChains[index].positions)) {
            for (const position of pearlChains[index].positions) {
                polyLines[index].push([position.latitude, position.longitude]);
            }
        }
    }

    const pathOptionsBoatList = [{
        color: 'blue', fillColor: 'blue', fillOpacity: 1,
    },{
        color: 'grey', fillColor: 'grey', fillOpacity: 1,
    }];

    const pathOptionsPearlChainPolylineList = [{
        color: 'blue', opacity: 0.25
    },{
        color: 'grey', opacity: 0.25
    }];

    const pathOptionsPearlChainList = [{
        color: 'blue', fillColor: 'blue', fillOpacity: 1
    },{
        color: 'grey', fillColor: 'grey', fillOpacity: 1
    }];

    return (
        <MapContainer
            center={[53.5677, 10.006]}
            zoom={initialZoom}
            style={{height: '100%', width: '100%'}}
            minZoom={initialZoom}
            maxZoom={18}
            maxBounds={bounds}
            maxBoundsViscosity={1.0}
        >
            <TileLayer
                attribution='&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a> contributors'
                url="https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png"
            />
            {
                buoys.map((buoy, index) => (
                    <Circle key={index} center={[buoy.latitude, buoy.longitude]} radius={10} pathOptions={{color: 'red', fillColor: 'yellow', fillOpacity: 1}} />
                ))
            }
            {
                polyLines.map((polyline, index) => (
                    <Polyline key={index} positions={polyline} pathOptions={pathOptionsPearlChainPolylineList[index]} />
                ))
            }
            {
                positions.map((position, index) => (
                    <Circle key={index} center={position} radius={20} pathOptions={pathOptionsBoatList[index]}/>
                ))
            }
            {
                polyLines.map((polyLine, index) => (
                    polyLine.map((position, index2) => (
                        <Circle key={`${index}.${index2}`} center={position} radius={7.5} pathOptions={pathOptionsPearlChainList[index]}/>
                    ))
                ))
            }
        </MapContainer>
    )
}