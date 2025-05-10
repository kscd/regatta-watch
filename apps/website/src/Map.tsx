import React, { useRef, useEffect } from 'react';
import alsterImg from './assets/alster.png'

type Position = {latitude: number, longitude: number, heading: number};

type MapProps = {positionN: number; positionW: number; heading: number; pearlChain: Position[]};

export const Map: React.FC<MapProps> = ({ positionN, positionW, heading, pearlChain}) => {
  const canvasRef = useRef<HTMLCanvasElement>(null);

  const buoy1 = {latitude: 53.565538, longitude: 10.014123-0.005}
  const buoy2 = {latitude: 53.558766+0.0035, longitude: 9.998720+0.0055}
  const buoy3 = {latitude: 53.576497-0.001, longitude: 10.004418+0.001}

  useEffect(() => {
    const canvas = canvasRef.current;

    if (!canvas) {
      console.error("Canvas element not found");
      return;
    }

    const context = canvas.getContext('2d');

    if (!context) {
      console.error("2D context not supported");
      return;
    }

    const drawBackgroundImage = () => {
        const backgroundImage = new Image();
        backgroundImage.src = alsterImg;
        backgroundImage.onload = () => {
          canvas.width = backgroundImage.width;
          canvas.height = backgroundImage.height;
          context.drawImage(backgroundImage, 0, 0);
          drawBuoys(buoy1);
          drawBuoys(buoy2);
          drawBuoys(buoy3);
          drawPearlChain(pearlChain);
          drawPosition(positionN, positionW, heading);
        };       
      };

    drawBackgroundImage(); // Call the function to draw the background image

    const drawBuoys = (buoy: {latitude: number, longitude: number}) => {
        context.beginPath();
        context.arc(calcPixelFromLongitude(buoy.longitude), calcPixelFromLatitude(buoy.latitude), 10, 0, 2 * Math.PI);
        context.fillStyle = 'yellow';
        context.fill();
    }


    const drawPosition = (positionN: number, positionW: number, heading: number) => {
      const headingInRadians = heading / 180 * Math.PI
      
      context.beginPath();
      context.arc(calcPixelFromLongitude(positionW), calcPixelFromLatitude(positionN), 10, 0.5 * Math.PI + headingInRadians, 0.5 * Math.PI + headingInRadians + 2 * Math.PI / 2);
      context.fillStyle = 'red';
      context.fill();

      context.beginPath();
      context.arc(calcPixelFromLongitude(positionW), calcPixelFromLatitude(positionN), 10, 1.5 * Math.PI + headingInRadians, 1.5 * Math.PI + headingInRadians + 2 * Math.PI / 2);
      context.fillStyle = 'green';
      context.fill();
    }

    const drawPearlChain = (pearlChain: Position[]) => {
      for (let i = 0; i < pearlChain.length; i++) {
        const headingInRadians = pearlChain[i].heading / 180 * Math.PI

        context.beginPath();
        context.arc(calcPixelFromLongitude(pearlChain[i].longitude), calcPixelFromLatitude(pearlChain[i].latitude), 5, 0.5 * Math.PI + headingInRadians, 0.5 * Math.PI + headingInRadians + 2 * Math.PI / 2);
        context.fillStyle = 'red';
        context.fill();

        context.beginPath();
        context.arc(calcPixelFromLongitude(pearlChain[i].longitude), calcPixelFromLatitude(pearlChain[i].latitude), 5, 1.5 * Math.PI + headingInRadians, 1.5 * Math.PI + headingInRadians + 2 * Math.PI / 2);
        context.fillStyle = 'green';
        context.fill();
      }
    }

    const draw = (event: MouseEvent) => {
      const rect = canvas.getBoundingClientRect();
      const x = event.clientX - rect.left;
      const y = event.clientY - rect.top;

      context.beginPath();
      context.arc(x, y, 10, 0, 2 * Math.PI);
      context.fillStyle = 'green';
      context.fill();
      context.fillStyle = 'red';
      context.fillText(x.toString(),x,y+10)
      context.fillText(y.toString(),x,y+20)
    };

    canvas.addEventListener('mousedown', draw);

    return () => {
      canvas.removeEventListener('mousedown', draw);
    };
  }, [positionN, positionW]);

  return <canvas ref={canvasRef} />;
};

const calcPixelFromLatitude = (latitude :number) => {
    const pixelA = 1024.8;
    const latitudeA = 53.557778;

    const pixelB = 41;
    const latitudeB = 53.577417;

    const latitudeDelta = latitudeB - latitudeA;
    const pixelDelta = pixelB - pixelA;

    return (latitude - latitudeA)/latitudeDelta*pixelDelta + pixelA
}

const calcPixelFromLongitude = (longitude :number) => {
    const pixelA = 73;
    const longitudeA = 9.997917;

    const pixelB = 435.2;
    const longitudeB = 10.010083;

    const longitudeDelta = longitudeB - longitudeA;
    const pixelDelta = pixelB - pixelA;

    return (longitude - longitudeA)/longitudeDelta*pixelDelta + pixelA
}

/*
Kennedy bridge: exact middle:

53°33'28.0"N 9°59'52.5"E
53.557778, 9.997917

73,1024.8

Langenzug bridge
53°34'38.7"N 10°00'36.3"E
53.577417, 10.010083

435.2, 41

Google maps picture: 650x1055, 1.623

1130x1828, 1.618
1280x1828, 1.428
1496x1828, 1.222

New Pictures:
*/

