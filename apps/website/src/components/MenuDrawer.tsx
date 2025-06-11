import {Box, Drawer, List, ListItem, ListItemButton, ListItemIcon, ListItemText} from "@mui/material";
import AccessTimeIcon from "@mui/icons-material/AccessTime";
import HighlightIcon from '@mui/icons-material/Highlight';
import SailingIcon from '@mui/icons-material/Sailing';
import React from "react";

type MenuDrawerProps = {
    open: boolean;
    handleClose: () => void;
    onOpenClockDialog: () => void;
    onOpenPearlChainDialog: () => void;
    onOpenBuoyDialog: () => void;
};

export const MenuDrawer: React.FC<MenuDrawerProps> = ({open, handleClose, onOpenClockDialog, onOpenPearlChainDialog, onOpenBuoyDialog}) =>
    <Drawer open={open} onClose={handleClose} anchor={'right'}>
        <Box sx={{ width: 250 }} role="presentation">
            <div className={"drawer-menu-heading"}>
                Configuration
            </div>
            <List>
                <ListItem>
                    <ListItemButton onClick={onOpenClockDialog}>
                        <ListItemIcon>
                            <AccessTimeIcon />
                        </ListItemIcon>
                        <ListItemText primary="Clock" />
                    </ListItemButton>
                </ListItem>
                <ListItem>
                    <ListItemButton onClick={onOpenPearlChainDialog}>
                        <ListItemIcon>
                            <SailingIcon />
                        </ListItemIcon>
                        <ListItemText primary="Pearl chain" />
                    </ListItemButton>
                </ListItem>
                <ListItem>
                    <ListItemButton onClick={onOpenBuoyDialog}>
                        <ListItemIcon>
                            <HighlightIcon />
                        </ListItemIcon>
                        <ListItemText primary="Buoy Location" />
                    </ListItemButton>
                </ListItem>
            </List>
        </Box>
    </Drawer>
