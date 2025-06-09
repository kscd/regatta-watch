import {Box, Drawer, List, ListItem, ListItemButton, ListItemIcon, ListItemText} from "@mui/material";
import AccessTimeIcon from "@mui/icons-material/AccessTime";
import React from "react";

type MenuDrawerProps = {
    open: boolean;
    handleClose: () => void;
    onOpenDialog: () => void;
};

export const MenuDrawer: React.FC<MenuDrawerProps> = ({open, handleClose, onOpenDialog}) =>
    <Drawer open={open} onClose={handleClose} anchor={'right'}>
        <Box sx={{ width: 250 }} role="presentation">
            <List>
                <ListItem>
                    <ListItemButton onClick={onOpenDialog}>
                        <ListItemIcon>
                            <AccessTimeIcon />
                        </ListItemIcon>
                        <ListItemText primary="Configure clock" />
                    </ListItemButton>
                </ListItem>
            </List>
        </Box>
    </Drawer>
