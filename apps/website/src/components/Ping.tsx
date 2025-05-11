type PingProps = {color: string, uptimeCount: number};

export const Ping: React.FC<PingProps> = ({ color, uptimeCount }) => {
  return (
    <>
      <div style={{float:"left", backgroundColor:color, width:"5cm", height:"1cm", color:"black"}}>uptime: {uptimeCount}s</div>
    </>
  )
}