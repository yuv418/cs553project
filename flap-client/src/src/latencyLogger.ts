type LatencyType = 'frame' | 'audio';

interface LatencyLog {
  sendTimestamps: number[];
  receiveTimestamps: number[];
}

const latencyLogs: Record<LatencyType, LatencyLog> = {
  frame: { sendTimestamps: [], receiveTimestamps: [] },
  audio: { sendTimestamps: [], receiveTimestamps: [] },
};

export function logSendTime(type: LatencyType) {
  latencyLogs[type].sendTimestamps.push(performance.now());
}

export function logReceiveTime(type: LatencyType) {
  latencyLogs[type].receiveTimestamps.push(performance.now());
}

export function downloadLatencyCSV() {
  let csv = "type,send_ms,receive_ms,diff_ms\n";
  for (const type of Object.keys(latencyLogs) as LatencyType[]) {
    const log = latencyLogs[type];
    const len = Math.min(log.sendTimestamps.length, log.receiveTimestamps.length);
    for (let i = 0; i < len; i++) {
      const send = log.sendTimestamps[i];
      const receive = log.receiveTimestamps[i];
      const diff = receive - send;
      csv += `${type},${send.toFixed(3)},${receive.toFixed(3)},${diff.toFixed(3)}\n`;
    }
  }

  const blob = new Blob([csv], { type: 'text/csv' });
  const a = document.createElement('a');
  a.href = URL.createObjectURL(blob);
  a.download = 'latency_data.csv';
  a.click();
}
