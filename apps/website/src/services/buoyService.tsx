
const getBuoys = async (): Promise<FetchedBuoys> => {
  const response = await fetch('http://localhost:8091/fetchbuoys',
      {
        method: 'POST',
      }
  );
  if (!response.ok) {
    throw new Error('Network response was not ok');
  }
  return response.json();
}

const setBuoys = async (buoys: Buoy[]): Promise<void> => {
  const response = await fetch('http://localhost:8091/setbuoys',
      {
        method: 'POST',
        body: JSON.stringify({
          buoys: buoys,
        })
      }
  );
  if (!response.ok) {
    throw new Error('Network response was not ok');
  }
  return response.json();
}

export type FetchedBuoys = {
  buoys: Buoy[];
}

export type Buoy = {
  id: string;
  version: number;
  latitude: number;
  longitude: number;
  pass_angle: number;
  is_pass_direction_clockwise: boolean;
  tolerance_in_meters: number;
};

export const BuoyService = {
  getBuoys,
  setBuoys,
}
