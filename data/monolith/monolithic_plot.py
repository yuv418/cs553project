import pandas as pd
import matplotlib.pyplot as plt

# Read the CSV file
file_path = '~/data/monolith/Sample.csv'
df = pd.read_csv(file_path)

# Task 1: Calculate Average Latency for Each DestSvcName
avg_latency = df.groupby('DestSvcName')['ReqTime'].mean().reset_index()
avg_latency.columns = ['DestSvcName', 'AverageLatency']

# Print average latency results
print("Average Latency by DestSvcName:")
print(avg_latency)
print("\n")

# Task 2: Instance to Latency Mapping (Latency Over Time)
# Group by DestSvcName and collect ReqTime as a list to preserve order
latency_mapping = df.groupby('DestSvcName')['ReqTime'].apply(list).to_dict()

# Print instance to latency mapping
print("Instance to Latency Mapping (ReqTime per instance):")
for svc, latencies in latency_mapping.items():
    print(f"{svc}: {latencies}")
print("\n")

# Plot 1: Average Latency Bar Plot (Save, don't show)
plt.figure(figsize=(10, 6))
plt.bar(avg_latency['DestSvcName'], avg_latency['AverageLatency'], color='skyblue')
plt.xlabel('DestSvcName')
plt.ylabel('Average Latency (ReqTime in microseconds)')
plt.title('Average Latency by DestSvcName')
plt.grid(True, axis='y')
plt.tight_layout()

# Save the average latency plot
plt.savefig('average_latency.png')
plt.close()  # Close to prevent display

# Plot 2: Latency Over Time Line Plot (Save, don't show)
plt.figure(figsize=(12, 8))
for svc, latencies in latency_mapping.items():
    plt.plot(range(len(latencies)), latencies, label=svc, marker='o')

plt.xlabel('Instance Index (Time Order)')
plt.ylabel('Latency (ReqTime in microseconds)')
plt.title('Latency Trends Over Time by DestSvcName')
plt.legend()
plt.grid(True)
plt.tight_layout()

# Save the latency over time plot
plt.savefig('latency_trends.png')
plt.close()  # Close to prevent display

# Save average latency results to CSV
avg_latency.to_csv('average_latency_by_destsvcname.csv', index=False)
