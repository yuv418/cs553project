#!/usr/bin/env python3
import os
import shutil
from pathlib import Path

import pandas as pd
import matplotlib.pyplot as plt
import math


def process_csv(csv_path: Path, out_root: Path):
    name = csv_path.stem
    out_dir = out_root / name
    out_dir.mkdir(exist_ok=True)

    # Copy the CSV itself
    shutil.copy(csv_path, out_dir / csv_path.name)

    # Load data
    df = pd.read_csv(csv_path)

    # 1) Average latency
    avg = df.groupby('DestSvcName')['ReqTime'].mean().reset_index()
    avg.columns = ['DestSvcName', 'AverageLatency']
    avg.to_csv(out_dir / 'average_latency_by_destsvcname.csv', index=False)

    # 2) Time-series mapping
    mapping = df.groupby('DestSvcName')['ReqTime'].apply(list).to_dict()
    services = list(mapping.keys())

    # 3) Plot average latency bar
    plt.figure(figsize=(10,6))
    plt.bar(avg['DestSvcName'], avg['AverageLatency'])
    plt.xlabel('DestSvcName')
    plt.ylabel('Average Latency (ReqTime)')
    plt.title(f'Average Latency by DestSvcName\n({name})')
    plt.grid(axis='y')
    plt.tight_layout()
    plt.savefig(out_dir / 'average_latency.png')
    plt.close()

    # 4) Plot small-multiples with distinct colors
    n = len(services)
    cols = 2
    rows = math.ceil(n / cols)
    fig, axes = plt.subplots(rows, cols, figsize=(cols*5, rows*3), sharey=True)

    # Choose a categorical colormap
    cmap = plt.get_cmap('tab10')

    for idx, svc in enumerate(services):
        ax = axes.flat[idx]
        lats = mapping[svc]
        color = cmap(idx % cmap.N)
        ax.plot(range(len(lats)), lats, marker='o', linestyle='-', color=color)
        ax.set_title(svc)
        ax.set_xlabel('Instance')
        ax.set_ylabel('Latency')
        ax.grid(True)

    # Hide unused subplots
    for ax in axes.flat[n:]:
        ax.set_visible(False)

    fig.suptitle(f'Latency Trends Over Time\n({name})', y=1.02)
    plt.tight_layout()
    plt.savefig(out_dir / 'latency_trends.png')
    plt.close()


def main():
    backend_dir = Path.home() / 'data' / 'graphs' / 'backend'
    if not backend_dir.is_dir():
        print(f"❌ Backend folder not found: {backend_dir}")
        return

    for csv_path in backend_dir.iterdir():
        if csv_path.is_file() and csv_path.suffix.lower() == '.csv':
            print(f"→ Processing {csv_path.name}")
            process_csv(csv_path, backend_dir)


if __name__ == '__main__':
    main()

