import { useEffect, useState } from 'react';
import {BoatService, Status} from '../services/boatService';

export const useUptime = () => {

    const [counter, setCounter] = useState(0);

    useEffect(() => {
      const interval = setInterval(async () => {
          BoatService.ping()
            .then((response: Status) => {
                if (response == "200") {
                    setCounter(counter + 1)
                }
            })
        }, 1000);
  
      return () => clearInterval(interval);
    }, [counter]);
  
    return counter
}