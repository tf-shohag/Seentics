# AWS EC2 Deployment Guide for Seentics Cloud

This guide walks you through deploying the Seentics Cloud platform on an AWS t3.medium instance with domain configuration.

## Prerequisites

- AWS account with EC2 access
- Domain names: `seentics.com` and `api.seentics.com` pointing to your EC2 instance
- SSH key pair for EC2 access

## Step 1: Launch EC2 Instance

1. **Launch t3.medium instance:**
   - AMI: Ubuntu 22.04 LTS
   - Instance Type: t3.medium (2 vCPU, 4 GB RAM)
   - Storage: 20-30 GB gp3 SSD
   - Security Group: Allow ports 22 (SSH), 80 (HTTP), 443 (HTTPS)

2. **Connect to your instance:**
```bash
ssh -i your-key.pem ubuntu@your-ec2-public-ip
```

## Step 2: Install Dependencies

```bash
# Update system
sudo apt update && sudo apt upgrade -y

# Install Docker
curl -fsSL https://get.docker.com -o get-docker.sh
sudo sh get-docker.sh
sudo usermod -aG docker ubuntu

# Install Docker Compose
sudo curl -L "https://github.com/docker/compose/releases/latest/download/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
sudo chmod +x /usr/local/bin/docker-compose

# Install Git
sudo apt install git -y

# Install Certbot for SSL
sudo apt install certbot -y

# Logout and login again for Docker group changes
exit
```

## Step 3: Clone and Setup Project

```bash
# Clone the repository
git clone https://github.com/your-username/seentics-cloud.git
cd seentics-cloud

# Create production environment file
cp .env.production.example .env.production
```

## Step 4: Configure Environment Variables

Edit `.env.production` with your actual values:

```bash
nano .env.production
```

**Required configurations:**
- `GLOBAL_API_KEY`: Generate a strong random key
- `USER_JWT_SECRET`: Generate a strong random key  
- `ANALYTICS_JWT_SECRET`: Generate a strong random key
- `WORKFLOW_JWT_SECRET`: Generate a strong random key
- `POSTGRES_PASSWORD`: Strong database password
- `MONGO_ROOT_PASSWORD`: Strong database password
- `REDIS_PASSWORD`: Strong database password
- `NEXT_PUBLIC_ADMIN_CODE`: Admin access code for dashboard

**Optional but recommended:**
- OAuth credentials (Google, GitHub)
- Lemon Squeezy API keys for payments
- Resend API key for emails

## Step 5: DNS Configuration

Configure your domain DNS records:

```
A    seentics.com        -> YOUR_EC2_PUBLIC_IP
A    www.seentics.com    -> YOUR_EC2_PUBLIC_IP  
A    api.seentics.com    -> YOUR_EC2_PUBLIC_IP
```

## Step 6: SSL Certificate Setup

```bash
# Stop any running web servers
sudo systemctl stop apache2 nginx 2>/dev/null || true

# Obtain SSL certificates
sudo certbot certonly --standalone -d seentics.com -d www.seentics.com
sudo certbot certonly --standalone -d api.seentics.com

# Create SSL directory structure
mkdir -p prod/nginx/ssl/certs/seentics.com
mkdir -p prod/nginx/ssl/certs/api.seentics.com  
mkdir -p prod/nginx/ssl/private/seentics.com
mkdir -p prod/nginx/ssl/private/api.seentics.com

# Copy certificates
sudo cp /etc/letsencrypt/live/seentics.com/fullchain.pem prod/nginx/ssl/certs/seentics.com/
sudo cp /etc/letsencrypt/live/seentics.com/privkey.pem prod/nginx/ssl/private/seentics.com/
sudo cp /etc/letsencrypt/live/api.seentics.com/fullchain.pem prod/nginx/ssl/certs/api.seentics.com/
sudo cp /etc/letsencrypt/live/api.seentics.com/privkey.pem prod/nginx/ssl/private/api.seentics.com/

# Set proper permissions
sudo chown -R $USER:$USER prod/nginx/ssl/
chmod -R 600 prod/nginx/ssl/private/
chmod -R 644 prod/nginx/ssl/certs/
```

## Step 7: Deploy Application

```bash
# Build and start all services
docker-compose -f docker-compose.prod.yml up -d --build

# Check service status
docker-compose -f docker-compose.prod.yml ps

# View logs if needed
docker-compose -f docker-compose.prod.yml logs -f
```

## Step 8: Verify Deployment

1. **Frontend:** Visit https://seentics.com
2. **API Health:** Visit https://api.seentics.com/health
3. **Admin Dashboard:** Visit https://seentics.com/admin

## Step 9: Setup Auto-renewal for SSL

```bash
# Add cron job for certificate renewal
sudo crontab -e
```

Add this line:
```
0 12 * * * /usr/bin/certbot renew --quiet && cd /home/ubuntu/seentics-cloud && docker-compose -f docker-compose.prod.yml restart nginx
```

## Step 10: Setup Monitoring (Optional)

```bash
# View real-time logs
docker-compose -f docker-compose.prod.yml logs -f

# Monitor resource usage
docker stats

# Check disk usage
df -h
```

## Architecture Overview

```
Internet
    ↓
AWS Load Balancer (optional)
    ↓
EC2 t3.medium
    ↓
Nginx Reverse Proxy
    ├── seentics.com → Frontend (Next.js)
    └── api.seentics.com → API Gateway
                              ├── User Service
                              ├── Analytics Service  
                              ├── Workflows Service
                              └── Databases (MongoDB, TimescaleDB, Redis)
```

## Resource Requirements

**t3.medium specifications:**
- 2 vCPU
- 4 GB RAM
- Up to 5 Gbps network performance

**Service allocation:**
- Frontend: 256MB RAM, 0.3 CPU
- API Gateway: 256MB RAM, 0.3 CPU
- User Service: 384MB RAM, 0.4 CPU
- Analytics Service: 512MB RAM, 0.5 CPU
- Workflows Service: 384MB RAM, 0.4 CPU
- MongoDB: 512MB RAM, 0.4 CPU
- TimescaleDB: 768MB RAM, 0.6 CPU
- Redis: 256MB RAM, 0.2 CPU
- Nginx: 128MB RAM, 0.2 CPU

## Troubleshooting

**Common issues:**

1. **Services not starting:** Check logs with `docker-compose logs service-name`
2. **SSL errors:** Verify certificate paths and permissions
3. **Memory issues:** Monitor with `docker stats` and adjust limits if needed
4. **DNS issues:** Verify domain records with `nslookup seentics.com`

**Useful commands:**
```bash
# Restart specific service
docker-compose -f docker-compose.prod.yml restart service-name

# Rebuild and restart
docker-compose -f docker-compose.prod.yml up -d --build service-name

# View service logs
docker-compose -f docker-compose.prod.yml logs -f service-name

# Access service shell
docker-compose -f docker-compose.prod.yml exec service-name sh
```

## Security Considerations

1. **Firewall:** Only allow necessary ports (22, 80, 443)
2. **SSH:** Use key-based authentication, disable password auth
3. **Updates:** Regularly update system packages and Docker images
4. **Secrets:** Never commit `.env.production` to version control
5. **Monitoring:** Set up CloudWatch or similar monitoring
6. **Backups:** Regular database backups to S3

## Scaling Considerations

For higher traffic, consider:
- Upgrading to larger EC2 instance (t3.large, t3.xlarge)
- Using AWS RDS for databases
- Adding AWS ElastiCache for Redis
- Using AWS Application Load Balancer
- Implementing horizontal scaling with multiple instances
