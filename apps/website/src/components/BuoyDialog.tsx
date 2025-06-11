import React from "react";
import {Button, Dialog, DialogContent, DialogTitle} from "@mui/material";

type BuoyDialogProps = {
    open: boolean;
    handleClose: () => void;
};

export const BuoyDialog: React.FC<BuoyDialogProps> = ({open, handleClose}) => {

    const handleConfirm = () => {
        handleClose();
    }

    return (
        <Dialog open={open} onClose={handleClose}>
            <DialogTitle id="alert-dialog-title">
                {"Set buoy configuration"}
            </DialogTitle>
            <DialogContent className='pearl-chain-dialog'>
                <Button variant={"contained"} onClick={handleConfirm}>Confirm</Button>
            </DialogContent>
        </Dialog>
    )
}

