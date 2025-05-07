type LatencyType = 'input' | 'audio' | 'frame';

interface LatencyLog {
  sendTimestamps: number[];
  receiveTimestamps: number[];
}

const latencyLogs: Record<LatencyType, LatencyLog> = {
  input: { sendTimestamps: [], receiveTimestamps: [] },
  audio: { sendTimestamps: [], receiveTimestamps: [] },
  frame: { sendTimestamps: [], receiveTimestamps: [] },
};

export function logSendTime(type: LatencyType) {
  latencyLogs[type].sendTimestamps.push(performance.now());
}

export function logReceiveTime(type: LatencyType) {
  latencyLogs[type].receiveTimestamps.push(performance.now());
}

export function downloadLatencyCSV() {
  let csv = "type,direction,time\n";
  for (const type of Object.keys(latencyLogs) as LatencyType[]) {
    const log = latencyLogs[type];
    for (let i = 0; i < log.sendTimestamps.length; i++) {
      const send = log.sendTimestamps[i];
      csv += `${type},send,${send.toFixed(3)}\n`;
    }
    for (let i = 0; i < log.receiveTimestamps.length; i++) {
      const recv = log.receiveTimestamps[i];
      csv += `${type},recv,${recv.toFixed(3)}\n`;
    }

    // Clear it out
    latencyLogs[type].sendTimestamps = []
    latencyLogs[type].receiveTimestamps = []
  }

  const blob = new Blob([csv], { type: 'text/csv' });
  const a = document.createElement('a');
  a.href = URL.createObjectURL(blob);
  a.download = 'latency_data.csv';
  a.click();


// 
}
