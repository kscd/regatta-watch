import { useEffect, useState } from 'react';

export const useCountdown = (startDate: number) => {
  const regattaMainTime = 24 * 60 * 60 * 1000
  const regattaOverTime =      60 * 60 * 1000

  let now = new Date().getTime()
  let newState = calcCounterState(now, startDate, regattaMainTime, regattaOverTime)
  const [countDown, setCountDown] = useState(newState);

  useEffect(() => {
    let now = new Date().getTime()
    let newState = calcCounterState(now, startDate, regattaMainTime, regattaOverTime)

    const interval = setInterval(() => {
      setCountDown(newState);
    }, 1000);

    return () => clearInterval(interval);
  }, [countDown]);

  return {status: countDown.state, description: getReturnState(countDown.state), timeLeft: getReturnTimes(countDown.milliseconds)};
};

const getReturnTimes = (milliseconds: number) => {
  // calculate time left
  const hours   = Math.floor( milliseconds / (1000 * 60 * 60));
  const minutes = Math.floor((milliseconds % (1000 * 60 * 60)) / (1000 * 60));
  const seconds = Math.floor((milliseconds % (1000 * 60)) / 1000);
  return {hours: hours, minutes: minutes, seconds: seconds};
};

const getReturnState = (state: number) => {
  switch(state) {
    case 0:
      return "regatta starts in"
    case 1:
      return "main time ends in"
    case 2:
      return "overtime ends in"
    case 3:
      return "regatta concluded"
    default:
      return "<unknown state>"
  }
}

const calcCounterState = (now: number, startDate: number, regattaMainTime: number, regattaOverTime: number ) => {
  if (startDate - now > 0) {
    // regatta has not yet started
     let newState = {
      state: 0,
      milliseconds: startDate - now
    }
    return newState
  } else if (startDate + regattaMainTime - now > 0) {
    // regatta is in main time slot
    let newState = {
      state: 1,
      milliseconds: startDate + regattaMainTime - now
    }
    return newState
  } else if (startDate + regattaMainTime + regattaOverTime - now > 0) {
    // regatta is in overtime slot
    let newState = {
      state: 2,
      milliseconds: startDate + regattaMainTime + regattaOverTime - now
    }
    return newState
  } else {
    // regatta is in overtime slot
    let newState = {
      state: 3,
      milliseconds: 0
    }
    return newState
  }
}

// copied from https://blog.greenroots.info/how-to-create-a-countdown-timer-using-react-hooks
