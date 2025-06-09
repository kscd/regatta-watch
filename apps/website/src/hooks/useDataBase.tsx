import { useEffect, useState } from 'react';
import {BoatService, BoatInfo, PearlChain} from '../services/boatService';

export const useDataBase = (boat1: string, boat2: string) => {
  const startPosition = {BoatInfo: {latitude: 53.5675975, longitude: 10.004, heading: 0, velocity: 0, distance: 0, round:1, section:1, crew0:"?", crew1:"?", next_crew0:"?", next_crew1:"?"}}
  const startPearlChain = {positions: [{latitude: 53.5675975, longitude: 10.004, heading: 0}]}
  const startRoundTimes = {round_times: [] as number[], section_times: [] as number[]};

  const [position1, setPosition1] = useState(startPosition);
  const [pearlChain1, setPearlChain1] = useState(startPearlChain);
  const [roundTime1, setRoundTime1] = useState(startRoundTimes);

  const [position2, setPosition2] = useState(startPosition);
  const [pearlChain2, setPearlChain2] = useState(startPearlChain);
  const [roundTime2, setRoundTime2] = useState(startRoundTimes);

  useEffect(() => {
    const interval = setInterval(async () => {
        if (boat1 !== "") {
            BoatService.getPosition(boat1)
                .then((position: BoatInfo) => {
                    setPosition1({BoatInfo: position});
                })
                .catch(() => console.error('Error fetching position'))

            BoatService.getPearlChain(boat1)
                .then((pearlChain: PearlChain) => {
                    setPearlChain1(pearlChain);
                })
                .catch(() => console.error('Error fetching pearl chain'))

            BoatService.getRoundTime(boat1)
                .then((roundTime: { round_times: number[], section_times: number[] }) => {
                    setRoundTime1(roundTime);
                })
                .catch(() => console.error('Error fetching round time'))
        }
        if (boat2 !== "") {
            BoatService.getPosition(boat2)
                .then((position: BoatInfo) => {
                    setPosition2({BoatInfo: position});
                })
                .catch(() => console.error('Error fetching position'))

            BoatService.getPearlChain(boat2)
                .then((pearlChain: PearlChain) => {
                    setPearlChain2(pearlChain);
                })
                .catch(() => console.error('Error fetching pearl chain'))

            BoatService.getRoundTime(boat2)
                .then((roundTime: { round_times: number[], section_times: number[] }) => {
                    setRoundTime2(roundTime);
                })
                .catch(() => console.error('Error fetching round time'))
        }
  }, 1000);
    return () => clearInterval(interval);
  }, [boat1, position1, pearlChain1, roundTime1, boat2, position2, pearlChain2, roundTime2]);

  return {position1, pearlChain1, roundTime1, position2, pearlChain2, roundTime2}
}
