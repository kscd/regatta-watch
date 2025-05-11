import { useEffect, useState } from 'react';
import {BoatService, Position, BoatInfo} from '../services/boatService';

export const useDataBase = () => {
  const startPosition = {latitude: 53.5675975, longitude: 10.004, heading: 0, velocity: 0, distance: 0, round:1, section:1, crew0:"?", crew1:"?", next_crew0:"?", next_crew1:"?"}
  const startPearlChain = {latitude: 53.5675975, longitude: 10.004, heading: 0}
  const startRoundTimes = { round_times: [] as number[], section_times: [] as number[] };

  const [position, setPosition] = useState(startPosition);
  const [pearlChain, setPearlChain] = useState([startPearlChain]);
  const [roundTime, setRoundTime] = useState(startRoundTimes);

  useEffect(() => {
    const interval = setInterval(async () => {

      BoatService.getPosition("Bluebird")
          .then((position: BoatInfo) => {
            setPosition(position);
          })
          .catch(() => console.error('Error fetching position'))

      BoatService.getPearlChain("Bluebird")
            .then((pearlChain: Position[]) => {
                setPearlChain(pearlChain);
            })
            .catch(() => console.error('Error fetching pearl chain'))

      BoatService.getRoundTime("Bluebird")
            .then((roundTime: { round_times: number[], section_times: number[] }) => {
                setRoundTime(roundTime);
            })
            .catch(() => console.error('Error fetching round time'))

  }, 1000);
    return () => clearInterval(interval);
  }, [position, pearlChain, roundTime]);

  return {position, pearlChain, roundTime}
}
