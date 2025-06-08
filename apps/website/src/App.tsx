import { Infoboard } from './components/InfoBoard.tsx'
import { RoundTimeBoard } from './components/RoundTimeBoard.tsx'
//import { Ping } from './Ping.tsx'
import { CountdownTimer } from './countdownTimer.tsx';
import { useDataBase } from './hooks/useDataBase.tsx';
import {Button} from "@mui/material";
import React from "react";
import {ClockDialog} from "./components/clockDialog.tsx";
import {useClockTime} from "./hooks/useClockTime.tsx";
import 'leaflet/dist/leaflet.css';
import './App.css'
import {Map} from "./components/Map.tsx";

function App() {
    const [isClockDialogOpen, setIsClockDialogOpen] = React.useState(false);

    //const regattaStartDate = new Date(1722682800000).getTime(); // Sat Aug 03 2024 13:00:00 GMT+0200 (Central European Summer Time)
    const regattaStartDate = new Date(1754132400000).getTime(); // Sat Aug 02 2025 13:00:00 GMT+0200 (Central European Summer Time)

    const {position, pearlChain, roundTime} = useDataBase();

    const clockTime = useClockTime();

    const handleOpenClockDialog = () => {
        setIsClockDialogOpen(true);
    };

    const handleCloseClockDialog = () => {
        setIsClockDialogOpen(false);
    };

    const boatPosition = {
        latitude: position.latitude,
        longitude: position.longitude,
        heading: position.heading
    }

    return (
        <>
            <div className={"page-container"}>
                <div className={"header-container"}>
                    <div className={"clock-container"}>
                        {clockTime}
                        <Button variant="contained" onClick={handleOpenClockDialog}>Configure clock</Button>
                    </div>
                    <h1>24h Regatta 2025</h1>
                    <div className={"countdown-container"}>
                        <CountdownTimer targetDate={regattaStartDate}/>
                    </div>
                </div>
                <div className={"regatta-container"}>
                    <div className={"boat-container"}>
                        <h2 className="boat-name">PSC Vivace (Kielzugvogel)</h2>
                        <Infoboard
                            latitude={position.latitude}
                            longitude={position.longitude}
                            heading={position.heading}
                            velocity={position.velocity}
                            distance={position.distance}
                            crew0={position.crew0}
                            crew1={position.crew1}
                            nextCrew0={position.next_crew0}
                            nextCrew1={position.next_crew1}
                        />
                        <RoundTimeBoard roundTimes={roundTime.round_times} sectionTimes={roundTime.section_times}></RoundTimeBoard>
                    </div>
                    <div className="map-container">
                        <Map boatPosition={boatPosition} pearlChain={pearlChain} />
                    </div>
                    <div className={"boat-container"}>
                        <h2 className="boat-name">PSC Bluebird (Conger)</h2>
                        <Infoboard
                            latitude={position.latitude}
                            longitude={position.longitude}
                            heading={position.heading}
                            velocity={position.velocity}
                            distance={position.distance}
                            crew0={position.crew0}
                            crew1={position.crew1}
                            nextCrew0={position.next_crew0}
                            nextCrew1={position.next_crew1}
                        />
                        <RoundTimeBoard roundTimes={roundTime.round_times} sectionTimes={roundTime.section_times}></RoundTimeBoard>
                    </div>
                </div>
            </div>
            <ClockDialog open={isClockDialogOpen} handleClose={handleCloseClockDialog} />
        </>
    )
}

export default App
