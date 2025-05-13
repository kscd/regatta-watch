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
        { field: 'round', headerName: '', width: 70, cellClassName: 'bold-column-cell' },
        { field: 'sectionTime1', headerName: 'Section 1', width: 110 },
        { field: 'sectionTime2', headerName: 'Section 2', width: 110 },
        { field: 'sectionTime3', headerName: 'Section 3', width: 110 },
        { field: 'sectionTime4', headerName: 'Section 4', width: 110 },
        { field: 'roundTime', headerName: 'Round Time', width: 130, cellClassName: 'bold-column-cell' },
    ];

    const rows = roundTimesArray.map((roundTime, roundIndex) => {
        const row: any = { id: roundIndex + 1, round: roundIndex + 1, roundTime: formatTime(roundTime) };
        sectionTimeRows[roundIndex].forEach((sectionTime, sectionTimeIndex) => {
            row[`sectionTime${sectionTimeIndex + 1}`] = formatTime(sectionTime);
        });
        return row;
    });

    const paginationModel = { page: 0, pageSize: 5 };

    return (
        <div className="table-container">
             <DataGrid
                rows={rows}
                columns={columns}
                initialState={{ pagination: { paginationModel } }}
                pageSizeOptions={[5, 10, 50]}
                sx={{ border: 0, fontSize: 20 }}
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