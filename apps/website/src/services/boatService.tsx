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

const getPearlChain = async (boat: string): Promise<PearlChain> => {

    const config = BoatService.getPearlChainConfiguration();

    const response = await fetch('http://localhost:8091/fetchpearlchain',
        {
            method: 'POST',
            body: JSON.stringify({boat: boat, length: config.length, interval: config.interval})
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

const setPearlChainConfiguration = ( config: PearlChainConfiguration): void => {
    localStorage.setItem('PearlChainLength', config.length.toString());
    localStorage.setItem('PearlChainInterval', config.interval.toString());
}

const getPearlChainConfiguration = (): PearlChainConfiguration => {
    localStorage.getItem('PearlChainLength');
    localStorage.getItem('PearlChainInterval');
    const length = parseInt(localStorage.getItem('PearlChainLength') || '3600', 10);
    const interval = parseInt(localStorage.getItem('PearlChainInterval') || '60', 10);
    return {
        length: length,
        interval: interval
    };
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

export type PearlChain = {
    positions: Position[];
}

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

export type PearlChainConfiguration = {
    length: number; // in seconds
    interval: number; // in seconds
}

export const BoatService = {
    getPosition,
    getPearlChain,
    getRoundTime,
    ping,
    getPearlChainConfiguration,
    setPearlChainConfiguration
}
