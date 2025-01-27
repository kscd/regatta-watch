import { useEffect, useState } from 'react';

export const useDataBase = () => {

  let startPearlChain = {latitude: 53.5675975, longitude: 10.004, heading: 0}
  let startPosition = {positionN: 53.5675975, positionW: 10.004, heading: 0, velocityInKnots: 0, distanceInNM: 0, round:1, section:1, crew0:"?", crew1:"?",nextCrew0:"?", nextCrew1:"?"}
  let startRoundTimes = { roundTimes: [] as number[], sectionTimes: [] as number[] };
  const [position, setPosition] = useState(startPosition);
  const [pearlChain, setPearlChain] = useState([startPearlChain]);
  const [roundTime, setRoundTime] = useState(startRoundTimes);

  useEffect(() => {
    const interval = setInterval(async () => {

    let response = await fetch('http://localhost:8091/fetchposition',
      {
        method: 'POST',
        body: JSON.stringify({boat: "Bluebird", no_later_than: '2006-01-02T15:04:05.000Z'})
      }
    )

    let result = await response.json()

    setPosition({
      positionN: result['latitude'],
      positionW: result['longitude'],
      heading:   result['heading'],
      velocityInKnots: result['velocity'],
      distanceInNM: result['distance'],
      round: result['round'],
      section: result['section'],
      crew0: result['crew0'],
      crew1: result['crew1'],
      nextCrew0: result['next_crew0'],
      nextCrew1: result['next_crew1'],
    });

      response = await fetch('http://localhost:8091/fetchpearlchain',
      {
        method: 'POST',
        body: JSON.stringify({boat: "Bluebird", no_later_than: '2006-01-02T15:04:05.000Z'})
      }
    )

    result = await response.json()

    setPearlChain(
      result['positions']
    )

    response = await fetch('http://localhost:8091/fetchroundtime',
      {
        method: 'POST',
            body: JSON.stringify({boat: "Bluebird"})
      }
    )

    result = await response.json()
      setRoundTime({
        roundTimes: result['round_times'],
        sectionTimes: result['section_times'],
      });

  }, 1000);
    return () => clearInterval(interval);
  }, [position, pearlChain, roundTime]);

  return {position, pearlChain, roundTime}
}
