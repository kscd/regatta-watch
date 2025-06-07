const setClockConfiguration = async (clock_time: string, clock_speed: number): Promise<void> => {
    const response = await fetch('http://localhost:8091/setclockconfiguration',
        {
            method: 'POST',
            body: JSON.stringify({
                clock_time: clock_time,
                clock_speed: clock_speed,
            })
        }
    );
    if (!response.ok) {
        throw new Error('Network response was not ok');
    }
}

const resetClockConfiguration = async (): Promise<void> => {
    const response = await fetch('http://localhost:8091/resetclockconfiguration',
        {
            method: 'POST',
            body: JSON.stringify({})
        }
    );
    if (!response.ok) {
        throw new Error('Network response was not ok');
    }
}

const getClockTime = async (): Promise<ClockTime> => {
    const response = await fetch('http://localhost:8091/getclocktime');
    if (!response.ok) {
        throw new Error('Network response was not ok');
    }
    return await response.json();
}

export type ClockTime = {
    time: string;
};

export const ClockService = {
    setClockConfiguration,
    resetClockConfiguration,
    getClockTime
}