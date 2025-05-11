const getPosition = async (boat: string): Promise<BoatInfo> => {
    const response = await fetch('http://localhost:8091/fetchposition',
        {
            method: 'POST',
            body: JSON.stringify({boat: boat})
        }
    );
    if (!response.ok) {
        throw new Error('Network response was not ok');
    }
    return response.json();
}

const getPearlChain = async (boat: string): Promise<Position[]> => {
    const response = await fetch('http://localhost:8091/fetchpearlchain',
        {
            method: 'POST',
            body: JSON.stringify({boat: boat})
        }
    );
    if (!response.ok) {
        throw new Error('Network response was not ok');
    }
    return response.json();
}

const getRoundTime = async (boat: string): Promise<RoundTime> => {
    const response = await fetch('http://localhost:8091/fetchroundtime',
        {
            method: 'POST',
            body: JSON.stringify({boat: boat})
        }
    );
    if (!response.ok) {
        throw new Error('Network response was not ok');
    }
    return response.json();
}

const ping = async (): Promise<Status> => {
    const response = await fetch('http://localhost:8091/ping');
    if (!response.ok) {
        throw new Error('Network response was not ok');
    }
    return response.json();
}

export type BoatInfo = {
    latitude: number;
    longitude: number;
    heading: number;
    velocity: number; // in knots
    distance: number; // in nautical miles
    round: number;
    section: number;
    crew0: string;
    crew1: string;
    next_crew0: string;
    next_crew1: string;
};

export type Position = {
    latitude: number;
    longitude: number;
    heading: number;
};

export type RoundTime = {
    round_times: number[];
    section_times: number[];
};

export type Status = string;

export const BoatService = {
    getPosition,
    getPearlChain,
    getRoundTime,
    ping
}
