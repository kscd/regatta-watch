import { useEffect, useState } from 'react';
import {ClockService, ClockTime} from '../services/clockService';
import dayjs from 'dayjs';

export const useClockTime = () => {

    const [clockTime, setClockTime] = useState('');

    useEffect(() => {
      const interval = setInterval(async () => {
          ClockService.getClockTime()
            .then((clockTime: ClockTime) => {
                setClockTime(dayjs(clockTime.time).format('MMMM D, YYYY HH:mm:ss'));
            })
        }, 1000);

      return () => clearInterval(interval);
    }, [clockTime]);

    return clockTime;
}
