import './App.css'
import { Infoboard } from './components/InfoBoard.tsx'
import { RoundTimeBoard } from './components/RoundTimeBoard.tsx'
import { Map } from './components/Map.tsx'
//import { Ping } from './Ping.tsx'
import { CountdownTimer } from './countdownTimer.tsx';
import { useDataBase } from './hooks/useDataBase.tsx';

function App() {

  const regattaStartDate = new Date(1722682800000).getTime(); // Sat Aug 03 2024 13:00:00 GMT+0200 (Central European Summer Time)

  const {position, pearlChain, roundTime} = useDataBase();

  return (
    <div className={"page-container"}>
      <div className={"map-container"}>
        <Map positionN={position.latitude} positionW={position.longitude} heading={position.heading} pearlChain={pearlChain} />
      </div>
      <div className={"boat-container"}>
        <h2 className="boat-name">PSC Bluebird (Conger)</h2>
        <Infoboard positionN={position.latitude} positionW={position.longitude} heading={position.heading}
                       velocity={position.velocity} distance={position.distance} round={position.round}
                       section={position.section} crew0={position.crew0} crew1={position.crew1}
                       nextCrew0={position.next_crew0} nextCrew1={position.next_crew1}/>
        <CountdownTimer targetDate={regattaStartDate}/>
        <RoundTimeBoard roundTimes={roundTime.round_times} sectionTimes={roundTime.section_times}></RoundTimeBoard>
      </div>
    </div>
  )
}

export default App
