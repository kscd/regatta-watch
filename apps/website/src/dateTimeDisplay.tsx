import React from 'react';

type DateTimeDisplayProps = { value: number; type: string; isDanger: boolean};

export const DateTimeDisplay: React.FC<DateTimeDisplayProps> = ({ value, type, isDanger }) => {
  return (
    <div className={isDanger ? 'countdown danger' : 'countdown'}>
      <p>{value}</p>
      <span>{type}</span>
    </div>
  );
};

// copied from https://blog.greenroots.info/how-to-create-a-countdown-timer-using-react-hooks
