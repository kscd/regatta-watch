import React from 'react';
import {Button, Dialog, DialogContent, DialogTitle, Divider, MenuItem, Select, SelectChangeEvent} from "@mui/material";
import {LocalizationProvider} from "@mui/x-date-pickers/LocalizationProvider";
import {AdapterDayjs} from "@mui/x-date-pickers/AdapterDayjs";
import {DateTimePicker} from "@mui/x-date-pickers/DateTimePicker";
import dayjs from 'dayjs';
import {ClockService} from "../services/clockService.tsx";

type ClockDialogProps = {
    open: boolean;
    handleClose: () => void;
};

export const ClockDialog: React.FC<ClockDialogProps> = ({open, handleClose}) => {
    const [customTime, setCustomTime] = React.useState<dayjs.Dayjs | null>(dayjs());
    const [speed, setSpeed] = React.useState('1');

    const handleCustomTimeChange = (newValue: dayjs.Dayjs | null) => {
        setCustomTime(newValue);
    };

    const handleSpeedChange = (event: SelectChangeEvent) => {
        setSpeed(event.target.value as string);
    };

    const setClockConfiguration = async () => {
        if (!customTime) {
            return;
        }

        const timeString = customTime.format('YYYY-MM-DDTHH:mm:ssZ');
        await ClockService.setClockConfiguration(timeString, parseInt(speed));
        handleClose();
    }

    const setClockConfigurationToRegatta2024 = async () => {
        const regatta2024Start = dayjs('2024-08-03T13:00:00+02:00'); // Start of 24h Regatta 2024
        await ClockService.setClockConfiguration(regatta2024Start.format('YYYY-MM-DDTHH:mm:ssZ'), parseInt(speed));
        handleClose();
    }

    const setClockConfigurationToRegatta2025 = async () => {
        const regatta2025Start = dayjs('2025-08-02T13:00:00+02:00'); // Start of 24h Regatta 2025
        await ClockService.setClockConfiguration(regatta2025Start.format('YYYY-MM-DDTHH:mm:ssZ'), parseInt(speed));
        handleClose();
    }

    const resetClockConfiguration = async () => {
        await ClockService.resetClockConfiguration();
        handleClose();
    }

    return (
        <Dialog open={open} onClose={handleClose}>
            <DialogTitle id="alert-dialog-title">
                {"Set clock configuration"}
            </DialogTitle>
            <DialogContent className='clock-dialog'>
                <div className="clock-dialog-content">
                    Speed of clock
                    <Select value={speed} onChange={handleSpeedChange}>
                        <MenuItem value={1}>1</MenuItem>
                        <MenuItem value={2}>2</MenuItem>
                        <MenuItem value={5}>5</MenuItem>
                        <MenuItem value={10}>10</MenuItem>
                        <MenuItem value={20}>20</MenuItem>
                        <MenuItem value={60}>60 (1 minute to 1 second)</MenuItem>
                        <MenuItem value={120}>120 (2 minutes to 1 second)</MenuItem>
                        <MenuItem value={360}>360 (6 minutes to 1 second)</MenuItem>
                        <MenuItem value={1800}>1800 (30 minutes to 1 second) </MenuItem>
                        <MenuItem value={3600}>3600 (1 hour to 1 second)</MenuItem>
                    </Select>
                </div>
                <Divider />
                <div className="clock-dialog-content">
                    <LocalizationProvider dateAdapter={AdapterDayjs}>
                        <DateTimePicker label="Custom Time" value={customTime} onChange={handleCustomTimeChange}/>
                    </LocalizationProvider>
                    <Button variant="contained" disabled={customTime === null} onClick={setClockConfiguration}>
                        Set configuration
                    </Button>
                </div>
                <Divider />
                <div className="clock-dialog-content">
                    <Button variant="contained" onClick={setClockConfigurationToRegatta2024}>Start of 24h Regatta 2024</Button>
                    <Button variant="contained" onClick={setClockConfigurationToRegatta2025}>Start of 24h Regatta 2025</Button>
                </div>
                <Divider />
                <div className="clock-dialog-content" onClick={resetClockConfiguration}>
                    <Button variant="contained">Reset time configuration</Button>
                </div>
            </DialogContent>
        </Dialog>
    )
}