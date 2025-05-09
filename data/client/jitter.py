#!/usr/bin/env python3
import csv
import bisect
import os
import matplotlib.pyplot as plt

# --- 1. CSV path ---
csv_path = os.path.expanduser('~/data/client/latency_data.csv')

# --- 2. Read timestamps ---
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

# --- 3. Sort lists ---
input_times.sort()
audio_times.sort()
frame_times.sort()

# --- 4. Compute per-input latencies (next recv after send) ---
audio_latencies = []
frame_latencies = []

for t_in in input_times:
    # next audio
    idx_a = bisect.bisect_right(audio_times, t_in)
    audio_latencies.append(audio_times[idx_a] - t_in
                            if idx_a < len(audio_times) else None)
    # next frame
    idx_f = bisect.bisect_right(frame_times, t_in)
    frame_latencies.append(frame_times[idx_f] - t_in
                            if idx_f < len(frame_times) else None)

# --- 5. Filter out any None (in case send outlasts recvs) ---
audio_lat = [lat for lat in audio_latencies if lat is not None]
frame_lat = [lat for lat in frame_latencies if lat is not None]

# --- 6. Compute jitter: |L[i] - L[i-1]| ---
audio_jitter = [abs(audio_lat[i] - audio_lat[i-1])
                for i in range(1, len(audio_lat))]
frame_jitter = [abs(frame_lat[i] - frame_lat[i-1])
                for i in range(1, len(frame_lat))]

# --- 7. Average jitter ---
avg_audio_jitter = sum(audio_jitter) / len(audio_jitter)
avg_frame_jitter = sum(frame_jitter) / len(frame_jitter)

print(f"Average audio jitter: {avg_audio_jitter:.3f} ms")
print(f"Average frame jitter: {avg_frame_jitter:.3f} ms")

# --- 8. Plot per-instance jitter ---
plt.figure()
plt.plot(range(1, len(audio_jitter)+1), audio_jitter, 'o-', label='Audio Jitter')
plt.plot(range(1, len(frame_jitter)+1), frame_jitter, 'o-', label='Frame Jitter')
plt.xlabel('Input Instance Index')
plt.ylabel('Jitter (ms)')
plt.title('Per-Instance Jitter')
plt.legend()
plt.tight_layout()
plt.savefig('jitter_instances.png')
plt.close()

# --- 9. Plot average jitter breakdown ---
plt.figure()
plt.bar(['Audio', 'Frame'], [avg_audio_jitter, avg_frame_jitter])
plt.ylabel('Average Jitter (ms)')
plt.title('Average Audio vs. Frame Jitter')
plt.tight_layout()
plt.savefig('average_jitter.png')
plt.close()

print("Saved  jitter_instances.png and average_jitter.png")

