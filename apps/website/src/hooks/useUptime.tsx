import { useEffect, useState } from 'react';

export const useUptime = () => {

    const [counter, setCounter] = useState(0);

    useEffect(() => {
      const interval = setInterval(async () => {
        let response = await fetch('http://localhost:8091/ping');
        let result = response.status;

        if (result == 200) {
            setCounter(counter + 1)            
        }

        }, 1000);
  
      return () => clearInterval(interval);
    }, [counter]);
  
    return counter
}