#!/usr/bin/env python3
import csv
import bisect
import os
import matplotlib.pyplot as plt

def load_timestamps(csv_path):
    input_times, audio_times, frame_times = [], [], []
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
    return sorted(input_times), sorted(audio_times), sorted(frame_times)

def compute_latencies(input_times, recv_times):
    """Map each input send time to the next recv time (or None)."""
    lats = []
    for t_in in input_times:
        idx = bisect.bisect_right(recv_times, t_in)
        lats.append(recv_times[idx] - t_in if idx < len(recv_times) else None)
    return lats

def plot_instances_and_average(input_times, audio_lats, frame_lats):
    # Filter out invalid entries
    inst_audio = [i for i, lat in enumerate(audio_lats) if lat is not None]
    vals_audio = [lat for lat in audio_lats if lat is not None]
    inst_frame = [i for i, lat in enumerate(frame_lats) if lat is not None]
    vals_frame = [lat for lat in frame_lats if lat is not None]

    # Per-instance
    plt.figure()
    plt.plot(inst_audio, vals_audio, 'o-', label='Audio')
    plt.plot(inst_frame, vals_frame, 'o-', label='Frame')
    plt.xlabel('Input Instance Index')
    plt.ylabel('Latency (ms)')
    plt.title('Per-Input Latencies')
    plt.legend()
    plt.tight_layout()
    plt.savefig('latency_instances.png')
    plt.close()

    # Average
    avg_a = sum(vals_audio) / len(vals_audio)
    avg_f = sum(vals_frame) / len(vals_frame)
    plt.figure()
    plt.bar(['Audio', 'Frame'], [avg_a, avg_f])
    plt.ylabel('Average Latency (ms)')
    plt.title('Average Audio vs. Frame Latency')
    plt.tight_layout()
    plt.savefig('average_latency.png')
    plt.close()

def plot_over_time(input_times, audio_lats, frame_lats):
    # Align times with valid latencies
    times_a = [t for t, lat in zip(input_times, audio_lats) if lat is not None]
    vals_a  = [lat for lat in audio_lats if lat is not None]
    times_f = [t for t, lat in zip(input_times, frame_lats) if lat is not None]
    vals_f  = [lat for lat in frame_lats if lat is not None]

    plt.figure()
    plt.plot(times_a, vals_a, 'o-', label='Audio Latency')
    plt.plot(times_f, vals_f, 'o-', label='Frame Latency')
    plt.xlabel('Send Time (ms)')
    plt.ylabel('Latency (ms)')
    plt.title('Latency over Time')
    plt.legend()
    plt.tight_layout()
    plt.savefig('latency_over_time.png')
    plt.close()

def main():
    csv_path = os.path.expanduser('~/data/client/latency_data.csv')
    input_times, audio_times, frame_times = load_timestamps(csv_path)

    audio_lats = compute_latencies(input_times, audio_times)
    frame_lats = compute_latencies(input_times, frame_times)

    plot_instances_and_average(input_times, audio_lats, frame_lats)
    plot_over_time(input_times, audio_lats, frame_lats)

    print("Saved  latency_instances.png, average_latency.png, latency_over_time.png")

if __name__ == "__main__":
    main()

