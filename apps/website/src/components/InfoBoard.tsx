import {Gauge, gaugeClasses} from '@mui/x-charts/Gauge';
import {BoatInfo} from "../services/boatService.tsx";
import React from "react";

type InfoBoardProps = { boatState: BoatInfo };

export const InfoBoard: React.FC<InfoBoardProps> = ({boatState}) => {
    const maxVelocity = 10;
    const velocityGaugeAngle = 90;
    const heading_ = calcHeadingFromNumber(boatState.heading);

    const headingWidth = 20;

    const displayCrew = boatState.crew0 !== "?"
    const displayNextCrew = boatState.next_crew0 !== "?"

    return(
        <div className="info-board">
            <div className="info-board-details">
                {calcLatitudeFromNumber(boatState.latitude)}, {calcLongitudeFromNumber(boatState.longitude)}
                <br/>
                {distanceFormat.format(boatState.distance)} NM, {distanceFormat.format(boatState.distance * 1.852)} km
                <br/>
                {displayCrew && <>{boatState.crew0}, {boatState.crew1}</>}
            </div>
            <div className={"info-board-gauges"}>
                <Gauge
                    width={150} height={100}
                    value={boatState.velocity > maxVelocity*100 ? maxVelocity*100 : boatState.velocity}
                    valueMax={maxVelocity*100}
                    startAngle={-velocityGaugeAngle}
                    endAngle={velocityGaugeAngle}
                    cornerRadius="50%"
                    sx={{
                        [`& .${gaugeClasses.valueText}`]: {
                            fontSize: 16,
                            transform: 'translate(0px, -20px)',
                        },
                    }}
                    text={`${velocityFormat.format(boatState.velocity)}kn\n${velocityFormat.format(boatState.velocity*1.852)}km/h`}
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
            <div className={"info-board-next"}>
                {displayNextCrew && <>next: {boatState.next_crew0}, {boatState.next_crew1}</>}
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