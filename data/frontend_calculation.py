#!/usr/bin/env python3
import os
import csv
import shutil
import bisect
import matplotlib.pyplot as plt
import pandas as pd
from pathlib import Path

# --- Helpers for latency & jitter ---
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

def compute_latencies(ins, recvs):
    lats = []
    for t in ins:
        idx = bisect.bisect_right(recvs, t)
        lats.append(recvs[idx] - t if idx < len(recvs) else None)
    return lats

def compute_jitter(latencies):
    vals = [l for l in latencies if l is not None]
    return [abs(vals[i] - vals[i-1]) for i in range(1, len(vals))]

def process_seed(ts_dir, collected_dir, seed, out_root, deploy_type):
    # Include deploy_type in the output directory
    seed_dir = out_root / f"{ts_dir.name}_{deploy_type}" / seed
    seed_dir.mkdir(parents=True, exist_ok=True)
    
    # Identify runs
    runs = sorted(
        [d for d in collected_dir.iterdir() if d.name.startswith(f"client_seed_{seed}_run_")],
        key=lambda d: int(d.name.rsplit('_', 1)[-1])
    )

    avg_data = []
    latency_plot_data = {'audio': {}, 'frame': {}}
    jitter_plot_data = {'audio': {}, 'frame': {}}

    for run_dir in runs:
        run_id = run_dir.name.rsplit('_', 1)[-1]
        csv_file = run_dir / 'latency_data.csv'
        if not csv_file.exists():
            continue
        shutil.copy(csv_file, seed_dir / f'{run_id}_latency_data.csv')

        ins, aud, frm = load_timestamps(csv_file)
        audio_lats = compute_latencies(ins, aud)
        frame_lats = compute_latencies(ins, frm)

        valid_audio = [(t, l) for t, l in zip(ins, audio_lats) if l is not None]
        valid_frame = [(t, l) for t, l in zip(ins, frame_lats) if l is not None]

        times_a, lats_a = zip(*valid_audio) if valid_audio else ([], [])
        times_f, lats_f = zip(*valid_frame) if valid_frame else ([], [])

        latency_plot_data['audio'][run_id] = (times_a, lats_a)
        latency_plot_data['frame'][run_id] = (times_f, lats_f)

        # print(f"frame timestamps {frm}")

        jitter_plot_data['audio'][run_id] = (aud[1:], compute_jitter(aud)) if len(lats_a) > 1 else ([], [])
        jitter_plot_data['frame'][run_id] = (frm[1:], compute_jitter(frm)) if len(lats_f) > 1 else ([], [])

        avg_data.append({
            'Run': run_id,
            'AvgAudioLatency': sum(lats_a)/len(lats_a) if lats_a else 0,
            'AvgFrameLatency': sum(lats_f)/len(lats_f) if lats_f else 0,
        })

    # Plot latency over time (side by side)
    fig, axes = plt.subplots(1, 2, figsize=(14, 6), sharex=False)
    for run_id, (t, l) in latency_plot_data['audio'].items():
        axes[0].plot(t, l, label=f'Run {run_id}', marker='o')
    axes[0].set_title('Audio Latency vs Time')
    axes[0].set_xlabel('Send Time (ms)')
    axes[0].set_ylabel('Latency (ms)')
    axes[0].legend()
    axes[0].grid()

    for run_id, (t, l) in latency_plot_data['frame'].items():
        axes[1].plot(t, l, label=f'Run {run_id}', marker='o')
    axes[1].set_title('Frame Latency vs Time')
    axes[1].set_xlabel('Send Time (ms)')
    axes[1].set_ylabel('Latency (ms)')
    axes[1].legend()
    axes[1].grid()

    fig.tight_layout()
    fig.savefig(seed_dir / 'latency_runs.png')
    plt.close(fig)

    # Plot jitter over time (side by side)
    fig, axes = plt.subplots(1, 2, figsize=(14, 6), sharex=False)
    for run_id, (t, j) in jitter_plot_data['audio'].items():
        axes[0].plot(t, j, label=f'Run {run_id}', marker='o')
    axes[0].set_title('Audio Jitter vs Time')
    axes[0].set_xlabel('Send Time (ms)')
    axes[0].set_ylabel('Jitter (ms)')
    axes[0].legend()
    axes[0].grid()

    for run_id, (t, j) in jitter_plot_data['frame'].items():
        axes[1].plot(t, j, label=f'Run {run_id}', marker='o')
    axes[1].set_title('Frame Jitter vs Time')
    axes[1].set_xlabel('Send Time (ms)')
    axes[1].set_ylabel('Jitter (ms)')
    axes[1].legend()
    axes[1].grid()

    fig.tight_layout()
    fig.savefig(seed_dir / 'jitter_runs.png')
    plt.close(fig)

    # Save average latency table as PNG
    df = pd.DataFrame(avg_data)
    df.to_csv(seed_dir / 'average_latency.csv', index=False)
    fig, ax = plt.subplots(figsize=(6, 1 + 0.4 * len(df)))
    ax.axis('tight')
    ax.axis('off')
    table = ax.table(cellText=df.values, colLabels=df.columns, loc='center', cellLoc='center')
    table.auto_set_font_size(False)
    table.set_fontsize(10)
    table.scale(1, 1.5)
    fig.tight_layout()
    fig.savefig(seed_dir / 'average_latency_table.png')
    plt.close(fig)

def main():
    base = Path(os.getcwd()) / 'data'
    out_root = base / 'graphs' / 'frontend'
    out_root.mkdir(parents=True, exist_ok=True)

    print(base)

    for ts in sorted(base.iterdir()):
        if not ts.name.startswith('2025') or not ts.is_dir():
            continue
        # Check for deploy_type file
        dt_file = ts / 'deploy_type'
        if not dt_file.is_file():
            print(f"  ✗ No deploy_type file in {ts.name}, skipping")
            continue
        # Read the first non-empty line from deploy_type
        with open(dt_file) as f:
            deploy_types = [line.strip() for line in f if line.strip()]
        if not deploy_types:
            print(f"  ✗ Empty deploy_type file in {ts.name}, skipping")
            continue
        deploy_type = deploy_types[0]  # Use the first deploy type
        print(f"  • Deploy type: {deploy_type}")

        for collected in ts.iterdir():
            if not collected.name.startswith('collected'):
                continue
            runs = [d for d in collected.iterdir() if d.name.startswith('client_seed')]
            seeds = set(d.name.split('_')[2] for d in runs)
            for seed in seeds:
                print(f"Processing seed {seed} in {ts.name}/{collected.name}")
                process_seed(ts, collected, seed, out_root, deploy_type)

if __name__=="__main__":
    main()
