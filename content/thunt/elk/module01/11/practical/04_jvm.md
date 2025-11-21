---
showTableOfContents: true
title: "Starting and Verifying Elasticsearch"
type: "page"
---




## Step 4.1: Enable and Start Elasticsearch

```bash
# Enable Elasticsearch to start on boot
sudo systemctl enable elasticsearch

# Start Elasticsearch service
sudo systemctl start elasticsearch

# Check service status
sudo systemctl status elasticsearch
```

You should see output like:

```
● elasticsearch.service - Elasticsearch
     Loaded: loaded (/lib/systemd/system/elasticsearch.service; enabled; vendor preset: enabled)
     Active: active (running) since Mon 2025-11-10 10:30:45 UTC; 5s ago
```

## Step 4.2: Verify Elasticsearch is Running

Wait about 30 seconds for Elasticsearch to fully start, then test:

```bash
# Test with curl (if security is enabled, use credentials)
curl -k -u elastic:YOUR_PASSWORD_HERE https://localhost:9200

# If you disabled security:
# curl http://localhost:9200
```

**Expected Response**:

```json
{
  "name" : "elk-node-01",
  "cluster_name" : "threat-hunting-elk",
  "cluster_uuid" : "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
  "version" : {
    "number" : "8.11.0",
    "build_flavor" : "default",
    "build_type" : "deb",
    "build_hash" : "abc123",
    "build_date" : "2024-01-15T10:05:34.472820Z",
    "build_snapshot" : false,
    "lucene_version" : "9.8.0",
    "minimum_wire_compatibility_version" : "7.17.0",
    "minimum_index_compatibility_version" : "7.0.0"
  },
  "tagline" : "You Know, for Search"
}
```

**Troubleshooting**: If Elasticsearch doesn't start:

```bash
# Check logs for errors
sudo journalctl -u elasticsearch -f

# Or check the log file directly
sudo tail -f /var/log/elasticsearch/threat-hunting-elk.log
```

Common issues:

- **Memory errors**: Reduce heap size
- **Port already in use**: Another service is using port 9200
- **Permission errors**: Check file ownership in `/var/lib/elasticsearch`



---
[|TOC|]({{< ref "../../../moc.md" >}})
[|PREV|]({{< ref "./03_config.md" >}})
[|NEXT|]({{< ref "./05_start.md" >}})

