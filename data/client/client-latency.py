#!/usr/bin/env python3
import csv
import bisect
import os
import matplotlib.pyplot as plt

# 1. Path to your CSV
csv_path = os.path.expanduser('~/data/client/latency_data.csv')

# 2. Read and split timestamps
input_times = []
audio_times = []
frame_times = []

with open(csv_path, newline='') as f:
    reader = csv.DictReader(f)
    for row in reader:
        t = float(row['time'])
        tp, direction = row['type'], row['direction']
        if tp == 'input' and direction == 'send':
            input_times.append(t)
        elif tp == 'audio' and direction == 'recv':
            audio_times.append(t)
        elif tp == 'frame' and direction == 'recv':
            frame_times.append(t)

# 3. Sort (just in case)
input_times.sort()
audio_times.sort()
frame_times.sort()

# 4. Helper to pair each recv with the nearest preceding send
def nearest_preceding(send_list, recv_time):
    idx = bisect.bisect_right(send_list, recv_time) - 1
    return send_list[idx] if idx >= 0 else None

# 5. Compute latencies (in same ms units)
audio_latencies = [t - nearest_preceding(input_times, t) for t in audio_times]
frame_latencies = [t - nearest_preceding(input_times, t) for t in frame_times]

# 6. Plot latency vs. instance
plt.figure()
plt.plot(range(len(audio_latencies)), audio_latencies, label='Audio')
plt.plot(range(len(frame_latencies)), frame_latencies, label='Frame')
plt.xlabel('Instance Index')
plt.ylabel('Latency (ms)')
plt.title('Per-Instance Latencies')
plt.legend()
plt.tight_layout()
plt.savefig('latency_instances.png')
plt.close()

# 7. Plot average latencies
avg_audio = sum(audio_latencies) / len(audio_latencies)
avg_frame = sum(frame_latencies) / len(frame_latencies)

plt.figure()
plt.bar(['Audio', 'Frame'], [avg_audio, avg_frame])
plt.ylabel('Average Latency (ms)')
plt.title('Average Audio vs. Frame Latency')
plt.tight_layout()
plt.savefig('average_latency.png')
plt.close()

print("Saved plots as:")
print(" • latency_instances.png")
print(" • average_latency.png")

