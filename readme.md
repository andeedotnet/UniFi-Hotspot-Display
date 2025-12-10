# UniFi Hotspot Display

UniFi Hotspot Display is a lightweight service that hosts an HTML page showing a QR code and a list of active UniFi hotspot vouchers. The vouchers are retrieved via the official UniFi Network API. This project is ideal for environments where guests need quick and easy access to Wi-Fi vouchers. 

Additionally, this tool ensures that the UniFi Network API key and other technical details remain securely hidden from end users.

## Features

- Hosts an HTML page showing a generated QR code for WIFI access and a list of active hotspot vouchers
- You can freely adapt the HTML page to your needs and design.
- Pulls voucher data from the official UniFi Network API
- Shows a list of currently active hotspot vouchers
- Runs inside a Docker container for easy deployment

## Requirements

- Docker
- Docker Compose
- UniFi Network Controller API access

## Installation

Clone the repository:
```
git clone <your-repository-url>
cd unifi-hotspot-display
```

Build and start the container:

```
docker compose up --build --force-recreate
```

(Optional) Configure your reverse proxy (Traefik/NGINX) to enable HTTPS.

## Configuration

Environment variables should be defined in a ```.env```file.

```
TZ=Europe/Berlin

# Unifi Network API settings
UNIFI_HOST=
UNIFI_SITE_ID=
UNIFI_NETWORK_API_KEY=

# WIFI credentials for QR Code generation
QRCODE_SSID=
QRCODE_PASSWORD=
QRCODE_WIFI_AUTH_TYPE=nopass
QRCODE_WIFI_HIDDEN=
```

## How to Use

After starting the Docker container, open your browser and navigate to the server's IP address on port 5005 (for example http://localhost:5005) to access the hosted hotspot display page.

## Recommended Setup

For improved security, it is recommended to run this service behind a reverse proxy such as Traefik or NGINX with proper SSL/TLS configuration.

## Known Limitations
Currently, only vouchers that allow a single login are supported. This is because the UniFi Network API filter ```authorizedGuestCount.eq``` can only be compared to an integer value, not to other fields. Example API call:
```
https://%s/proxy/network/integration/v1/sites/%s/hotspot/vouchers?limit=7&filter=authorizedGuestCount.eq(0)
```

Without filtering by ```authorizedGuestCount.eq(0)```, vouchers that have already been used but are still active will also be displayed in the list. 

Ideally, a filter like ```authorizedGuestCount.lt(authorizedGuestLimit)``` would allow showing all vouchers that have remaining uses, but this is not currently supported by the API.

#### WiFi QR Code Format used in this tool:
```
WIFI:T:<QRCODE_WIFI_AUTH_TYPE>;S:<QRCODE_SSID>;P:<QRCODE_PASSWORD>;H:<QRCODE_WIFI_HIDDEN>;;
```

Use QRCODE_WIFI_AUTH_TYPE 'nopass" for open networks or 'WPA'

## License

This project is released under the MIT License.

