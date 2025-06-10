import { InfoBoard } from './components/InfoBoard.tsx'
import { RoundTimeBoard } from './components/RoundTimeBoard.tsx'
//import { Ping } from './Ping.tsx'
import { CountdownTimer } from './countdownTimer.tsx';
import { useDataBase } from './hooks/useDataBase.tsx';
import {
    IconButton,
} from "@mui/material";
import React from "react";
import {ClockDialog} from "./components/clockDialog.tsx";
import {useClockTime} from "./hooks/useClockTime.tsx";
import 'leaflet/dist/leaflet.css';
import './App.css'
import {Map} from "./components/Map.tsx";
import MenuIcon from '@mui/icons-material/Menu';
import {MenuDrawer} from "./components/MenuDrawer.tsx";
import {PearlChainDialog} from "./components/PearlChainDialog.tsx";

function App() {
    const [isMenuDrawerOpen, setIsMenuDrawerOpen] = React.useState(false);
    const [isClockDialogOpen, setIsClockDialogOpen] = React.useState(false);
    const [isPearlChainDialogOpen, setIsPearlChainDialogOpen] = React.useState(false);

    //const regattaStartDate = new Date(1722682800000).getTime(); // Sat Aug 03 2024 13:00:00 GMT+0200 (Central European Summer Time)
    const regattaStartDate = new Date(1754132400000).getTime(); // Sat Aug 02 2025 13:00:00 GMT+0200 (Central European Summer Time)

    const {position1, pearlChain1, roundTime1, position2, pearlChain2, roundTime2} = useDataBase("Bluebird", "Vivace");
    const clockTime = useClockTime();

    const handleOpenMenuDrawer = () => {
        setIsMenuDrawerOpen(true);
    };

    const handleCloseMenuDrawer = () => {
        setIsMenuDrawerOpen(false);
    };

    const handleOpenClockDialog = () => {
        setIsClockDialogOpen(true);
        handleCloseMenuDrawer();
    };

    const handleCloseClockDialog = () => {
        setIsClockDialogOpen(false);
    };

    const handleOpenPearlChainDialog = () => {
        setIsPearlChainDialogOpen(true);
        handleCloseMenuDrawer();
    };

    const handleClosePearlChainDialog = () => {
        setIsPearlChainDialogOpen(false);
    };

    const boatPosition1 = {
        latitude: position1.BoatInfo.latitude,
        longitude: position1.BoatInfo.longitude,
        heading: position1.BoatInfo.heading,
    }

    const boatPosition2 = {
        latitude: position2.BoatInfo.latitude,
        longitude: position2.BoatInfo.longitude,
        heading: position2.BoatInfo.heading,
    }

    return (
        <>
            <div className={"page-container"}>
                <div className={"header-container"}>
                    <div className={"clock-container"}>
                        {clockTime}
                    </div>
                    <h1>24h Regatta 2025</h1>
                    <div className={"countdown-container"}>
                        <CountdownTimer targetDate={regattaStartDate}/>
                    </div>
                    <div className={"menu-button-container"}>
                        <IconButton aria-label="delete" onClick={handleOpenMenuDrawer}>
                            <MenuIcon />
                        </IconButton>
                    </div>
                </div>
                <div className={"regatta-container"}>
                    <div className={"boat-container"}>
                        <h2 className="boat-name">PSC Vivace (Kielzugvogel)</h2>
                        <InfoBoard boatState={position2.BoatInfo}/>
                        <RoundTimeBoard roundTimes={roundTime2.round_times} sectionTimes={roundTime2.section_times}></RoundTimeBoard>
                    </div>
                    <div className="map-container">
                        <Map boatPositions={[boatPosition1, boatPosition2]} pearlChains={[pearlChain1, pearlChain2]} />
                    </div>
                    <div className={"boat-container"}>
                        <h2 className="boat-name">PSC Bluebird (Conger)</h2>
                        <InfoBoard boatState={position1.BoatInfo}/>
                        <RoundTimeBoard roundTimes={roundTime1.round_times} sectionTimes={roundTime1.section_times}></RoundTimeBoard>
                    </div>
                </div>
            </div>
            <MenuDrawer
                open={isMenuDrawerOpen}
                handleClose={handleCloseMenuDrawer}
                onOpenClockDialog={handleOpenClockDialog}
                onOpenPearlChainDialog={handleOpenPearlChainDialog} />
            <ClockDialog open={isClockDialogOpen} handleClose={handleCloseClockDialog} />
            <PearlChainDialog open={isPearlChainDialogOpen} handleClose={handleClosePearlChainDialog}></PearlChainDialog>
        </>
    )
}

export default App
