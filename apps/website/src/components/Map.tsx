import React from "react";
import L from 'leaflet';
import {Circle, MapContainer, Polyline, TileLayer} from "react-leaflet";
import {PearlChain, Position} from "../services/boatService.tsx";

type MapProps = { boatPositions: Position[], pearlChains: PearlChain[]};

export const Map: React.FC<MapProps> = ({boatPositions, pearlChains}) => {
    const initialZoom = 15;

    const bounds: L.LatLngBoundsExpression = [
        [53.5677 - 0.015, 10.006 - 0.015],
        [53.5677 + 0.015, 10.006 + 0.015]
    ];

    const buoy1: L.LatLngExpression = [53.565538, 10.009123]
    const buoy2: L.LatLngExpression = [53.562266, 10.00422]
    const buoy3: L.LatLngExpression = [53.575497, 10.005418]

    const positions: L.LatLngExpression[] = [];
    const polylines: L.LatLngExpression[][] = [];
    for (const [index, position] of boatPositions.entries()) {
        positions.push([position.latitude, position.longitude]);
        polylines.push([[position.latitude, position.longitude]]);
        for (const position of pearlChains[index].positions) {
            polylines[index].push([position.latitude, position.longitude]);
        }
    }

    const pathOptionsBoatList = [{
        color: 'blue',
        fillColor: 'blue',
        fillOpacity: 1,
    },{
        color: 'grey',
        fillColor: 'grey',
        fillOpacity: 1,
    }];

    const pathOptionsPearlChainPolylineList = [{
        color: 'blue',
        opacity: 0.25
    },{
        color: 'grey',
        opacity: 0.25
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
            <Circle center={buoy1} radius={10} pathOptions={{color: 'red', fillColor: 'orange', fillOpacity: 1}}/>
            <Circle center={buoy2} radius={10} pathOptions={{color: 'red', fillColor: 'orange', fillOpacity: 1}}/>
            <Circle center={buoy3} radius={10} pathOptions={{color: 'red', fillColor: 'orange', fillOpacity: 1}}/>
            {
                polylines.map((polyline, index) => (
                    <Polyline key={index} positions={polyline} pathOptions={pathOptionsPearlChainPolylineList[index]} />
                ))
            }
            {
                positions.map((position, index) => (
                    <Circle key={index} center={position} radius={20} pathOptions={pathOptionsBoatList[index]}/>
                ))
            }
            {
                pearlChains[0].positions.map((position, index) => (
                    <Circle key={index} center={[position.latitude, position.longitude]} radius={7.5} pathOptions={{color: 'blue', fillColor: 'blue', fillOpacity: 0.5}}/>
                ))
            }
        </MapContainer>
    )
}