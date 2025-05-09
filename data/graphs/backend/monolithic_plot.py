#!/usr/bin/env python3
import os
import shutil
import pandas as pd
import matplotlib.pyplot as plt
from pathlib import Path

def process_csv(csv_path: Path, out_root: Path):
    name = csv_path.stem
    out_dir = out_root / name
    out_dir.mkdir(exist_ok=True)

    # 1) copy the CSV itself
    shutil.copy(csv_path, out_dir / csv_path.name)

    # 2) load
    df = pd.read_csv(csv_path)

    # 3) average latency by DestSvcName
    avg = df.groupby('DestSvcName')['ReqTime'].mean().reset_index()
    avg.columns = ['DestSvcName', 'AverageLatency']
    avg.to_csv(out_dir / 'average_latency_by_destsvcname.csv', index=False)

    # 4) latency-over-time mapping
    mapping = df.groupby('DestSvcName')['ReqTime'].apply(list).to_dict()

    # 5) plot average latency bar
    plt.figure(figsize=(10,6))
    plt.bar(avg['DestSvcName'], avg['AverageLatency'])
    plt.xlabel('DestSvcName')
    plt.ylabel('Average Latency (ReqTime)')
    plt.title(f'Average Latency by DestSvcName\n({name})')
    plt.grid(axis='y')
    plt.tight_layout()
    plt.savefig(out_dir / 'average_latency.png')
    plt.close()

    # 6) plot latency trends
    plt.figure(figsize=(12,8))
    for svc, lats in mapping.items():
        plt.plot(range(len(lats)), lats, marker='o', label=svc)
    plt.xlabel('Instance Index')
    plt.ylabel('Latency (ReqTime)')
    plt.title(f'Latency Trends Over Time\n({name})')
    plt.legend()
    plt.grid()
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

