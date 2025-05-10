#!/usr/bin/env python3
import os
import csv
import shutil
from pathlib import Path
import pandas as pd
import matplotlib.pyplot as plt
import math

# Common CSV header
HEADER = ["SrcSvcName", "SrcSvcVerb", "DestSvcName", "DestSvcVerb", "GameId", "ReqTime"]
# For microservices
SERVICES = ["auth", "engine", "initiator", "music", "score", "worldgen"]

def combine_micro(ts_dir: Path, collected: str, remote_dir: Path, deploy: str, out_dir: Path) -> Path:
    out_name = f"{deploy}_{collected}_combined.csv"
    tmp_path = ts_dir / out_name
    print(f"  • [Micro] Creating {tmp_path}")

    with open(tmp_path, "w", newline="") as fout:
        writer = csv.writer(fout)
        writer.writerow(HEADER)
        for svc in SERVICES:
            svc_csv = remote_dir / svc / "stats.csv"
            if svc == "auth" and not svc_csv.exists():
                svc_csv.parent.mkdir(parents=True, exist_ok=True)
                with open(svc_csv, "w", newline="") as f:
                    f.write(",".join(HEADER) + "\n")
            if not svc_csv.exists():
                print(f"    – skipping missing {svc_csv}")
                continue
            with open(svc_csv, newline="") as f:
                reader = csv.reader(f)
                next(reader, None)
                for row in reader:
                    writer.writerow(row)

    dest = out_dir / out_name
    shutil.copy(tmp_path, dest)
    print(f"    → Copied to {dest}")
    return tmp_path

def combine_monolith(ts_dir: Path, collected: str, remote_dir: Path, out_dir: Path) -> Path:
    out_name = f"monolith_single_instance_{collected}_combined.csv"
    tmp_path = ts_dir / out_name
    mono_dir = remote_dir / "monolith"
    print(f"  • [Monolith] Creating {tmp_path}")

    with open(tmp_path, "w", newline="") as fout:
        writer = csv.writer(fout)
        writer.writerow(HEADER)
        new_csv = mono_dir / "stats.csv"
        if new_csv.exists():
            with open(new_csv, newline="") as f:
                reader = csv.reader(f)
                next(reader, None)
                for row in reader:
                    writer.writerow(row)
        else:
            print(f"    – missing {new_csv}")
        old_csv = mono_dir / "stats_old.csv"
        if old_csv.exists():
            print(f"    – appending {old_csv}")
            with open(old_csv, newline="") as f:
                reader = csv.reader(f)
                next(reader, None)
                for row in reader:
                    writer.writerow(row)
        else:
            print(f"    – no {old_csv}")

    dest = out_dir / out_name
    shutil.copy(tmp_path, dest)
    print(f"    → Copied to {dest}")
    return tmp_path

def process_csv(csv_path: Path, out_root: Path):
    name = csv_path.stem
    out_dir = out_root / name
    out_dir.mkdir(exist_ok=True)

    # Copy the CSV
    shutil.copy(csv_path, out_dir / csv_path.name)

    # Load data
    try:
        df = pd.read_csv(csv_path)
    except Exception as e:
        print(f"    – Error reading {csv_path}: {e}")
        return

    # 1) Average latency
    avg = df.groupby('DestSvcName')['ReqTime'].mean().reset_index()
    avg.columns = ['DestSvcName', 'AverageLatency']
    avg.to_csv(out_dir / 'average_latency_by_destsvcname.csv', index=False)

    # 2) Time-series mapping
    mapping = df.groupby('DestSvcName')['ReqTime'].apply(list).to_dict()
    services = list(mapping.keys())

    # 3) Plot average latency bar
    plt.figure(figsize=(10, 6))
    plt.bar(avg['DestSvcName'], avg['AverageLatency'])
    plt.xlabel('Destination Service', fontsize=12)
    plt.ylabel('Latency (ns)', fontsize=12)
    plt.title(f'Average Latency by Service ({name})', fontsize=14, pad=15)
    plt.grid(axis='y')
    plt.tight_layout(pad=2.0)
    avg_plot_path = out_dir / 'average_latency.png'
    plt.savefig(avg_plot_path, bbox_inches='tight')
    plt.close()
    print(f"    – Saved average latency plot: {avg_plot_path}")

    # 4) Plot small-multiples with distinct colors
    n = len(services)
    cols = 2
    rows = math.ceil(n / cols)
    fig, axes = plt.subplots(rows, cols, figsize=(cols * 5, rows * 3), sharey=True)
    cmap = plt.get_cmap('tab10')

    for idx, svc in enumerate(services):
        ax = axes.flat[idx]
        lats = mapping[svc]
        color = cmap(idx % cmap.N)
        ax.plot(range(len(lats)), lats, marker='o', linestyle='-', color=color)
        ax.set_title(svc, fontsize=12)
        ax.set_xlabel('Instance', fontsize=10)
        ax.set_ylabel('Latency (ns)', fontsize=10)
        ax.grid(True)

    for ax in axes.flat[n:]:
        ax.set_visible(False)

    fig.suptitle(f'Latency Trends ({name})', fontsize=14, y=1.05)
    plt.tight_layout(pad=2.0)
    trends_plot_path = out_dir / 'latency_trends.png'
    plt.savefig(trends_plot_path, bbox_inches='tight')
    plt.close()
    print(f"    – Saved latency trends plot: {trends_plot_path}")

def main():
    base = Path(os.getcwd()) / "data"
    graphs_dir = base / "graphs" / "backend"
    graphs_dir.mkdir(parents=True, exist_ok=True)
    print(f"Graphs output directory: {graphs_dir}")

    for entry in sorted(base.iterdir()):
        if not entry.is_dir() or entry.name == "graphs":
            continue
        ts_dir = entry
        print(f"\n=== Processing {ts_dir} ===")
        dt_file = ts_dir / "deploy_type"
        if not dt_file.is_file():
            print("  ✗ no deploy_type file, skipping")
            continue

        with open(dt_file) as f:
            deploys = [line.strip() for line in f if line.strip()]

        collected_dirs = [d for d in ts_dir.iterdir() if d.is_dir() and d.name.startswith("collected")]
        if not collected_dirs:
            print("  ✗ no collected_* dirs, skipping")
            continue

        for collected_dir in sorted(collected_dirs):
            collected = collected_dir.name
            remote_dir = collected_dir / "remote"
            if not remote_dir.is_dir():
                print(f"  ✗ no remote/ under {collected}, skipping")
                continue

            print(f" • Collected set: {collected}")
            for deploy in deploys:
                if deploy.startswith(("microservices_", "microservices_multi_region")):
                    csv_path = combine_micro(ts_dir, collected, remote_dir, deploy, graphs_dir)
                    process_csv(csv_path, graphs_dir)
                elif deploy.startswith("monolith"):
                    csv_path = combine_monolith(ts_dir, collected, remote_dir, graphs_dir)
                    process_csv(csv_path, graphs_dir)
                else:
                    print(f"    – unknown deploy '{deploy}', skipping")

if __name__ == "__main__":
    main()
