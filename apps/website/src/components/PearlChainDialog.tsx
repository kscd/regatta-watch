import React from "react";
import {Button, Dialog, DialogContent, DialogTitle, Slider} from "@mui/material";
import {BoatService} from "../services/boatService.tsx";

type PearlChainDialogProps = {
    open: boolean;
    handleClose: () => void;
};

export const PearlChainDialog: React.FC<PearlChainDialogProps> = ({open, handleClose}) => {
    const config = BoatService.getPearlChainConfiguration()
    const [pearlChainLength, setPearlChainLength] = React.useState<number>(config.length);
    const [pearlChainInterval, setPearlChainInterval] = React.useState<number>(config.interval);

    const handlePearlChainIntervalChange = (_: Event, newValue: number | number[]) => {
        setPearlChainInterval(newValue as number);
    };

    const handlePearlChainLengthChange = (_: Event, newValue: number | number[]) => {
        setPearlChainLength(newValue as number);
    };

    const handleConfirm = () => {
        BoatService.setPearlChainConfiguration({length: pearlChainLength, interval: pearlChainInterval});
        handleClose();
    }

    return (
        <Dialog open={open} onClose={handleClose}>
            <DialogTitle id="alert-dialog-title">
                {"Set pearl chain configuration"}
            </DialogTitle>
            <DialogContent className='pearl-chain-dialog'>
                <div className="pearl-chain-slider-container">
                    Pearl chain length
                    <div className="pearl-chain-slider-with-text">
                        <div className="pearl-chain-slider">
                            <Slider
                                aria-label="Pearl chain length"
                                value={pearlChainLength}
                                onChange={handlePearlChainLengthChange}
                                max={86400} // 24 hours
                                step={1200} // 20 minutes
                            />
                        </div>
                        {minutesToTimeString(pearlChainLength)}
                    </div>
                </div>
                <div className="pearl-chain-slider-container">
                    Pearl chain interval
                    <div className="pearl-chain-slider-with-text">
                        <div className="pearl-chain-slider">
                            <Slider
                                aria-label="Pearl chain interval"
                                value={pearlChainInterval}
                                onChange={handlePearlChainIntervalChange}
                                max={3600} // 1 hour
                                step={60} // 1 minute
                            />
                        </div>
                        {minutesToTimeString(pearlChainInterval)}
                    </div>
                </div>
                <Button variant={"contained"} onClick={handleConfirm}>
                    Confirm
                </Button>
            </DialogContent>
        </Dialog>
    )
}

const minutesToTimeString = (seconds: number) => {
    const minutesRaw = Math.trunc(seconds / 60);
    const hours = Math.trunc(minutesRaw / 60);
    const minutes = minutesRaw % 60;
    return `${hours}h ${minuteFormat.format(minutes)}min`
};

const minuteFormat = new Intl.NumberFormat("en-US", {
    minimumIntegerDigits: 2,
});
