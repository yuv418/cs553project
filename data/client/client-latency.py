#!/usr/bin/env python3
import csv
import bisect
import os
import matplotlib.pyplot as plt

# 1. Load CSV
csv_path = os.path.expanduser('~/data/client/latency_data.csv')
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

# 2. Sort
input_times.sort()
audio_times.sort()
frame_times.sort()

# 3. For each input, find the next audio & frame via bisect
audio_latencies = []
frame_latencies = []

for t_in in input_times:
    # find first audio > t_in
    idx_a = bisect.bisect_right(audio_times, t_in)
    if idx_a < len(audio_times):
        audio_latencies.append(audio_times[idx_a] - t_in)
    else:
        audio_latencies.append(None)

    # find first frame > t_in
    idx_f = bisect.bisect_right(frame_times, t_in)
    if idx_f < len(frame_times):
        frame_latencies.append(frame_times[idx_f] - t_in)
    else:
        frame_latencies.append(None)

# 4. Clean out any None (if you prefer)
instances = list(range(len(input_times)))
audio_vals = [lat for lat in audio_latencies if lat is not None]
frame_vals = [lat for lat in frame_latencies if lat is not None]
inst_audio = [i for i, lat in enumerate(audio_latencies) if lat is not None]
inst_frame = [i for i, lat in enumerate(frame_latencies) if lat is not None]

# 5. Plot per-input-instance latencies
plt.figure()
plt.plot(inst_audio, audio_vals, 'o-', label='Audio')
plt.plot(inst_frame, frame_vals, 'o-', label='Frame')
plt.xlabel('Input Instance Index')
plt.ylabel('Latency (ms)')
plt.title('Per-Input Latencies')
plt.legend()
plt.tight_layout()
plt.savefig('latency_instances.png')
plt.close()

# 6. Plot average
avg_audio = sum(audio_vals) / len(audio_vals)
avg_frame = sum(frame_vals) / len(frame_vals)

plt.figure()
plt.bar(['Audio', 'Frame'], [avg_audio, avg_frame])
plt.ylabel('Average Latency (ms)')
plt.title('Average Audio vs. Frame Latency')
plt.tight_layout()
plt.savefig('average_latency.png')
plt.close()

print("Saved latency_instances.png and average_latency.png")

