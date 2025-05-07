# docker-prometheus-8311
Dockerized Prometheus exporter for 8311 community firmware statistics.

## Environment variables

- `SSH_HOST` - IP address of the 8311 PON stick (optional, default: `192.168.11.1`)
- `SSH_PORT` - SSH port of the 8311 PON stick (optional, default: `22`)
- `SSH_USERNAME` - SSH username of the 8311 PON stick (optional, default: `root`)
- `SSH_PASSWORD` - SSH password of the 8311 PON stick (required)

## Usage

```bash
docker run -d -p 9100:9100 -e SSH_PASSWORD=yourpassword e7db/docker-prometheus-8311:latest
```
