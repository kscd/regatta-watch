import React from 'react';
import { useCountdown } from './hooks/useCountdown.tsx';
import {DateTimeDisplay} from './dateTimeDisplay.tsx'

type CountdownTimerProps = { targetDate: number};

export const CountdownTimer: React.FC<CountdownTimerProps> = ({ targetDate }) => {

  const countDownState = useCountdown(targetDate);

  if (countDownState.timeLeft.hours + countDownState.timeLeft.minutes + countDownState.timeLeft.seconds <= 0) {
    return <ExpiredNotice />;
  } else {
    return (
      <ShowCounter description={countDownState.description} hours={countDownState.timeLeft.hours} minutes={countDownState.timeLeft.minutes} seconds={countDownState.timeLeft.seconds}/>
    );
  }
};

const ExpiredNotice = () => {
    return (
      <div className="expired-notice">
        <span>Regatta concluded</span>
      </div>
    );
  };

type ShowCounterProps = {description: string, hours: number, minutes: number, seconds:number };

export const ShowCounter: React.FC<ShowCounterProps> = ({description, hours, minutes, seconds }) => {
    return (
      <div className="show-counter">
      <h3>{description}</h3>
          <DateTimeDisplay value={hours} type={'Hours'} isDanger={hours == 0} />
          <p>:</p>
          <DateTimeDisplay value={minutes} type={'Mins'} isDanger={hours == 0} />
          <p>:</p>
          <DateTimeDisplay value={seconds} type={'Seconds'} isDanger={hours == 0} />
      </div>
    );
};

// copied from https://blog.greenroots.info/how-to-create-a-countdown-timer-using-react-hooks
