import React from 'react';
import {DataGrid, GridColDef} from "@mui/x-data-grid";
type RoundTimeBoardProps = {
    roundTimes: number[];
    sectionTimes: number[];
};

export const RoundTimeBoard: React.FC<RoundTimeBoardProps> = ({ roundTimes, sectionTimes }) => {
    // Helper function to chunk sectionTimes into rows of 4
    const chunkArray = (array: number[], chunkSize: number) => {
        const chunks = [];
        for (let i = 0; i < array.length; i += chunkSize) {
            chunks.push(array.slice(i, i + chunkSize));
        }
        return chunks;
    };

    const roundTimesArray = roundTimes || [] as number[];
    const sectionTimesArray = sectionTimes || [] as number[];

    // Chunk sectionTimes into rows of 4
    const sectionTimeRows = chunkArray(sectionTimesArray, 4);

    const columns: GridColDef[] = [
        { field: 'round', headerName: 'Round', width: 70, cellClassName: 'bold-column-cell'},
        { field: 'sectionTime1', headerName: 'Sec 1', width: 90, align: 'right', headerAlign: 'right' },
        { field: 'sectionTime2', headerName: 'Sec 2', width: 90, align: 'right', headerAlign: 'right' },
        { field: 'sectionTime3', headerName: 'Sec 3', width: 90, align: 'right', headerAlign: 'right' },
        { field: 'sectionTime4', headerName: 'Sec 4', width: 90, align: 'right', headerAlign: 'right' },
        { field: 'roundTime', headerName: 'Time', width: 100, cellClassName: 'bold-column-cell', align: 'right', headerAlign: 'right' },
    ];

    const rows = roundTimesArray.map((roundTime, roundIndex) => {
        const row: any = { id: roundIndex + 1, round: roundIndex + 1, roundTime: formatTime(roundTime) };
        sectionTimeRows[roundIndex].forEach((sectionTime, sectionTimeIndex) => {
            row[`sectionTime${sectionTimeIndex + 1}`] = formatTime(sectionTime);
        });
        return row;
    });

    return (
        <div className="data-grid-container">
             <DataGrid
                rows={rows.reverse()}
                columns={columns}
                density={"compact"}
                disableColumnSelector = {true}
                disableColumnResize={true}
                disableColumnFilter={true}
                hideFooterSelectedRowCount={true}
                isRowSelectable={() => false}
                rowSelection={false}
                hideFooter={true}
                sx={{ border: 0, fontSize: 16, fontFamily: 'monospace' }}
            />
        </div>
    );
};

function formatTime(seconds: number) {
    // Calculate hours, minutes, and seconds
    const hours = Math.floor(seconds / 3600);
    const minutes = Math.floor((seconds % 3600) / 60);
    const secs = Math.floor(seconds % 60)

    // Pad hours, minutes, and seconds with leading zeros if necessary
    const paddedHours = String(hours).padStart(1, '0');
    const paddedMinutes = String(minutes).padStart(2, '0');
    const paddedSeconds = String(secs).padStart(2, '0');

    return `${paddedHours}:${paddedMinutes}:${paddedSeconds}`;
}