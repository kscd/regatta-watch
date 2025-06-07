import './App.css'
import { Infoboard } from './components/InfoBoard.tsx'
import { RoundTimeBoard } from './components/RoundTimeBoard.tsx'
import { Map } from './components/Map.tsx'
//import { Ping } from './Ping.tsx'
import { CountdownTimer } from './countdownTimer.tsx';
import { useDataBase } from './hooks/useDataBase.tsx';
import {Button} from "@mui/material";
import React from "react";
import {ClockDialog} from "./components/clockDialog.tsx";
import {useClockTime} from "./hooks/useClockTime.tsx";

function App() {
    const [isClockDialogOpen, setIsClockDialogOpen] = React.useState(false);

    const regattaStartDate = new Date(1722682800000).getTime(); // Sat Aug 03 2024 13:00:00 GMT+0200 (Central European Summer Time)

    const {position, pearlChain, roundTime} = useDataBase();

    const clockTime = useClockTime();

    const handleOpenClockDialog = () => {
        setIsClockDialogOpen(true);
    };

    const handleCloseClockDialog = () => {
        setIsClockDialogOpen(false);
    };

    return (
        <>
            <div className={"page-container"}>
                <div className={"map-container"}>
                    <Map positionN={position.latitude} positionW={position.longitude} heading={position.heading} pearlChain={pearlChain} />
                </div>
                <div className={"boat-container"}>
                    <h2 className="boat-name">PSC Bluebird (Conger)</h2>
                    <Infoboard positionN={position.latitude} positionW={position.longitude} heading={position.heading}
                        velocity={position.velocity} distance={position.distance} round={position.round}
                        section={position.section} crew0={position.crew0} crew1={position.crew1}
                        nextCrew0={position.next_crew0} nextCrew1={position.next_crew1}/>
                    <CountdownTimer targetDate={regattaStartDate}/>
                    <div className={"show-counter"}>
                        {clockTime}
                        <Button variant="contained" onClick={handleOpenClockDialog}>Configure clock</Button>
                    </div>
                    <RoundTimeBoard roundTimes={roundTime.round_times} sectionTimes={roundTime.section_times}></RoundTimeBoard>
                </div>
            </div>
            <ClockDialog open={isClockDialogOpen} handleClose={handleCloseClockDialog} />
        </>
    )
}

export default App
