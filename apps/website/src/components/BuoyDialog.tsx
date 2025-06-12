import React, {useEffect, useState} from "react";
import {Button, Dialog, DialogContent, DialogTitle, IconButton, TextField} from "@mui/material";
import {Map} from "./Map.tsx";
import {Buoy} from "../services/buoyService.tsx";
import AddIcon from "@mui/icons-material/Add";
import RemoveIcon from '@mui/icons-material/Remove';

type BuoyDialogProps = {
    open: boolean;
    handleClose: () => void;
    buoys: Buoy[];
};

export const BuoyDialog: React.FC<BuoyDialogProps> = ({open, handleClose, buoys}) => {
    const [buoyData, setBuoyData] = useState<Buoy[]>([]);

    useEffect(() => {
        setBuoyData(buoys.map(buoy => ({ ...buoy })));
    }, [buoys]);

    const handleChange = (event: React.ChangeEvent<HTMLInputElement>, buoyId: string) => {
        const { name, value } = event.target;

        const parsedValue = name === "latitude" || name === "longitude"
            ? parseFloat(value) || 0
            : value;

        setBuoyData(prevBuoyData =>
            prevBuoyData.map(buoy =>
                buoy.id === buoyId
                    ? {
                        ...buoy,
                        [name]: parsedValue,
                    }
                    : buoy
            )
        );
    };

    const handleCoordinateChange = (buoyId: string, fieldName: 'latitude' | 'longitude', increment: number) => {
        setBuoyData(prevBuoyData =>
            prevBuoyData.map(buoy => {
                if (buoy.id === buoyId) {
                    const currentValue = buoy[fieldName];
                    const newValue = (typeof currentValue === 'number' ? currentValue : 0) + increment;
                    return {
                        ...buoy,
                        [fieldName]: Math.round(newValue * 1000000) / 1000000,
                    };
                }
                return buoy;
            })
        );
    };

    const handleConfirm = () => {
        handleClose();
    }

    return (
        <Dialog open={open} onClose={handleClose} maxWidth={false}>
            <DialogTitle id="alert-dialog-title">
                {"Set buoy configuration"}
            </DialogTitle>
            <DialogContent className='buoy-dialog'>
                <div className="buoy-dialog-map-container">
                    <Map buoys={buoyData} boatPositions={[]} pearlChains={[]} buoysOld={buoys} />
                </div>
                <div className="buoy-dialog-elements-container">
                    <div className={"buoy-dialog-controls-container"}>
                    {buoyData.map(buoy => (
                        <div key={buoy.id} className={"buoy-container"}>
                            <h3>{buoy.id}</h3>
                            <div className={"buoy-controls-container"}>
                                <div className={"buoy-coordinate-container"}>
                                    <IconButton onClick={() => handleCoordinateChange(buoy.id, 'latitude', -0.0001)}>
                                        <RemoveIcon />
                                    </IconButton>
                                    <TextField
                                        label="Latitude"
                                        name="latitude"
                                        value={buoy.latitude}
                                        onChange={(e: React.ChangeEvent<HTMLInputElement>) => handleChange(e, buoy.id)}
                                    />
                                    <IconButton onClick={() => handleCoordinateChange(buoy.id, 'latitude', +0.0001)}>
                                        <AddIcon />
                                    </IconButton>

                                </div>
                                <div className={"buoy-coordinate-container"}>
                                    <IconButton onClick={() => handleCoordinateChange(buoy.id, 'longitude', -0.0001)}>
                                        <RemoveIcon />
                                    </IconButton>
                                    <TextField
                                        label="Longitude"
                                        name="longitude"
                                        value={buoy.longitude}
                                        onChange={(e: React.ChangeEvent<HTMLInputElement>) => handleChange(e, buoy.id)}
                                    />
                                    <IconButton onClick={() => handleCoordinateChange(buoy.id, 'longitude', +0.0001)}>
                                        <AddIcon />
                                    </IconButton>
                                </div>
                            </div>
                        </div>
                    ))}
                </div>
                <Button variant={"contained"} onClick={handleConfirm} fullWidth={true}>Confirm</Button>
                </div>
            </DialogContent>
        </Dialog>
    )
}
