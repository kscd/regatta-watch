type InfoboardProps = { positionN: number; positionW: number; heading: number, velocity: number, distance: number, round: number, section: number, crew0: string, crew1: string, nextCrew0: string, nextCrew1: string};

export const Infoboard: React.FC<InfoboardProps> = ({ positionN, positionW, heading, velocity, distance, round, section,crew0,crew1,nextCrew0,nextCrew1}) => (
    <div className="show-counter">
    <table style={{width: "100%"}} color="#FFFFFF">
      <tbody>
      <tr>
        <td>position</td>
        <td>{calcLatitudeFromNumber(positionN)}, {calcLongitudeFromNumber(positionW)}</td>
      </tr>
      <tr>
        <td>heading</td>
        <td>{calcHeadingFromNumber(heading)}</td>
      </tr>
      <tr>
        <td>velocity</td>
        <td>{velocityFormat.format(velocity)} kn, {velocityFormat.format(velocity * 1.852)} km/h</td>
      </tr>
      <tr>
        <td>distance</td>
        <td>{distanceFormat.format(distance)} NM, {distanceFormat.format(distance * 1.852)} km</td>
      </tr>
      <tr>
        <td>round</td>
        <td>{round}, section {section}</td>
      </tr>
      <tr>
        <td>crew</td>
        <td>{crew0}, {crew1}</td>
      </tr>
      <tr>
        <td>next crew</td>
        <td>{nextCrew0}, {nextCrew1}</td>
      </tr>
      </tbody>
    </table>
    </div>
);

let calcLatitudeFromNumber = (a: number) => {
      a = a > 90 ? 90 : a
      a = a < -90? -90: a

      let char = a >= 0? "N" : "S"
      let degree = Math.abs(Math.trunc(a))
      let minute = (Math.abs(a) - degree)*60
      return `${degree}° ${minuteFormat.format(minute)}' ${char}`
    };

      let calcLongitudeFromNumber = (a: number) => {
      a = (a + 180) % 360 - 180

      let char = a >= 0 || a == 180 ? "E" : "W"
      let degree = Math.abs(Math.trunc(a))
      let minute = (Math.abs(a) - degree)*60
      return `${degree}° ${minuteFormat.format(minute)}' ${char}`
    };

      let calcHeadingFromNumber = (a: number) => {
      a = ((a % 360) + 360) % 360 // positive-only modulo

      let degree = Math.trunc(a)
      return `${degree}°`
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