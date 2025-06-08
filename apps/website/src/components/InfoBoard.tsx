import {Gauge, gaugeClasses} from '@mui/x-charts/Gauge';

type InfoboardProps = { latitude: number; longitude: number; heading: number, velocity: number, distance: number, crew0: string, crew1: string, nextCrew0: string, nextCrew1: string};

export const Infoboard: React.FC<InfoboardProps> = ({ latitude, longitude, heading, velocity, distance, crew0,crew1,nextCrew0,nextCrew1}) => {
  const maxVelocity = 10;
  const velocityGaugeAngle = 90;
  const heading_ = calcHeadingFromNumber(heading);

  const headingWidth = 20;

  const displayCrew = crew0 !== "?"
  const displayNextCrew = nextCrew0 !== "?"

  return(
      <div className="infoboard">
        <div className="infoboard-details">
        {calcLatitudeFromNumber(latitude)}, {calcLongitudeFromNumber(longitude)}
        <br/>
        {distanceFormat.format(distance)} NM, {distanceFormat.format(distance * 1.852)} km
        <br/>
        {displayCrew && <>{crew0}, {crew1}</>}
        </div>
        <div className={"infoboard-gauges"}>
        <Gauge
            width={150} height={100}
            value={velocity/100 > maxVelocity ? maxVelocity : velocity/100}
            valueMax={maxVelocity}
            startAngle={-velocityGaugeAngle}
            endAngle={velocityGaugeAngle}
            cornerRadius="50%"
            sx={{
              [`& .${gaugeClasses.valueText}`]: {
                fontSize: 16,
                transform: 'translate(0px, -20px)',
              },
            }}
            text={`${velocityFormat.format(velocity/100)}kn\n${velocityFormat.format(velocity*1.852/100)}km/h`}
        />
        <Gauge
            width={100}
            height={100}
            value={headingWidth}
            valueMax={360}
            startAngle={heading_-headingWidth/2}
            endAngle={360+heading_-headingWidth/2}
            cornerRadius="50%"
            innerRadius="70%"
            sx={{
                [`& .${gaugeClasses.valueText}`]: {
                    fontSize: 16,
                },
            }}
            text={`${heading_}°`}
        />
        </div>
        <div className={"infoboard-next"}>
          {displayNextCrew && <>next: {nextCrew0}, {nextCrew1}</>}
        </div>
      </div>
  );
}

const calcLatitudeFromNumber = (a: number) => {
  a = a > 90 ? 90 : a
  a = a < -90? -90: a

  const char = a >= 0? "N" : "S"
  const degree = Math.abs(Math.trunc(a))
  const minute = (Math.abs(a) - degree)*60
  return `${degree}° ${minuteFormat.format(minute)}' ${char}`
};

const calcLongitudeFromNumber = (a: number) => {
  a = (a + 180) % 360 - 180

  const char = a >= 0 || a == 180 ? "E" : "W"
  const degree = Math.abs(Math.trunc(a))
  const minute = (Math.abs(a) - degree)*60
  return `${degree}° ${minuteFormat.format(minute)}' ${char}`
};

const calcHeadingFromNumber = (a: number) => {
  a = ((a % 360) + 360) % 360 // positive-only modulo
  return Math.trunc(a)
};

const minuteFormat = new Intl.NumberFormat("en-US", {
  minimumFractionDigits: 3,
  maximumFractionDigits: 3,
  minimumIntegerDigits: 2,
});

const velocityFormat = new Intl.NumberFormat("en-US", {
  minimumFractionDigits: 1,
  maximumFractionDigits: 1,
});

const distanceFormat = new Intl.NumberFormat("en-US", {
  minimumFractionDigits: 2,
  maximumFractionDigits: 2,
});