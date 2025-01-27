import React from 'react';

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

    return (
        <div className="table-container">
            <table className="round-time-table">
                <thead className="round-time-thead">
                <tr>
                    <th className="round-time-th"></th>
                    <th className="round-time-th" colSpan={4} style={{minWidth: '20em'}}>Section Times</th>
                    <th className="round-time-th"> Round Times</th>
                </tr>
                </thead>
                <tbody className="scrollable-tbody">
                {roundTimesArray.map((roundTime, roundIndex) => (
                    <React.Fragment key={roundIndex}>
                        <tr>
                            <td className="round-time-td">{roundIndex + 1}</td>
                            {sectionTimeRows[roundIndex].map((sectionTime, sectionTimeIndex) => (
                                <td className="round-time-td-section" key={sectionTimeIndex}>{formatTime(sectionTime)}</td>
                            ))}
                            {/* Add empty cells if sectionTimeRow has less than 4 items */}
                            {Array(4 - sectionTimeRows[roundIndex].length).fill(<td></td>)}
                            <td className="round-time-td">{formatTime(roundTime)}</td>
                        </tr>
                    </React.Fragment>
                ))}
                </tbody>
            </table>
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