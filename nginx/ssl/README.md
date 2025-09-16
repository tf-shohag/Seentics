# SSL Certificate Setup for Production

This directory should contain SSL certificates for your domains.

## Required Certificate Structure

```
/prod/nginx/ssl/
├── certs/
│   ├── seentics.com/
│   │   └── fullchain.pem
│   └── api.seentics.com/
│       └── fullchain.pem
└── private/
    ├── seentics.com/
    │   └── privkey.pem
    └── api.seentics.com/
        └── privkey.pem
```

## Obtaining SSL Certificates

### Option 1: Let's Encrypt (Recommended)

1. Install certbot on your AWS EC2 instance:
```bash
sudo apt update
sudo apt install certbot python3-certbot-nginx
```

2. Obtain certificates for both domains:
```bash
# For main domain
sudo certbot certonly --standalone -d seentics.com -d www.seentics.com

# For API subdomain
sudo certbot certonly --standalone -d api.seentics.com
```

3. Copy certificates to the correct locations:
```bash
# Create directories
sudo mkdir -p /path/to/seentics-cloud/prod/nginx/ssl/certs/seentics.com
sudo mkdir -p /path/to/seentics-cloud/prod/nginx/ssl/certs/api.seentics.com
sudo mkdir -p /path/to/seentics-cloud/prod/nginx/ssl/private/seentics.com
sudo mkdir -p /path/to/seentics-cloud/prod/nginx/ssl/private/api.seentics.com

# Copy certificates
sudo cp /etc/letsencrypt/live/seentics.com/fullchain.pem /path/to/seentics-cloud/prod/nginx/ssl/certs/seentics.com/
sudo cp /etc/letsencrypt/live/seentics.com/privkey.pem /path/to/seentics-cloud/prod/nginx/ssl/private/seentics.com/
sudo cp /etc/letsencrypt/live/api.seentics.com/fullchain.pem /path/to/seentics-cloud/prod/nginx/ssl/certs/api.seentics.com/
sudo cp /etc/letsencrypt/live/api.seentics.com/privkey.pem /path/to/seentics-cloud/prod/nginx/ssl/private/api.seentics.com/
```

### Option 2: CloudFlare SSL (Alternative)

If using CloudFlare, you can generate Origin Server certificates from the CloudFlare dashboard and place them in the appropriate directories.

## Auto-renewal Setup

Set up automatic certificate renewal:
```bash
sudo crontab -e
```

Add this line:
```
0 12 * * * /usr/bin/certbot renew --quiet && docker-compose -f /path/to/seentics-cloud/docker-compose.prod.yml restart nginx
```

## DNS Configuration

Make sure your DNS records point to your AWS EC2 instance:

```
A    seentics.com        -> YOUR_EC2_PUBLIC_IP
A    www.seentics.com    -> YOUR_EC2_PUBLIC_IP
A    api.seentics.com    -> YOUR_EC2_PUBLIC_IP
```

## Security Notes

- Keep private keys secure and never commit them to version control
- Use strong SSL configurations (TLS 1.2+ only)
- Enable HSTS headers (already configured in nginx.conf)
- Consider using OCSP stapling for better performance
