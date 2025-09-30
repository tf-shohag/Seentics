# AWS RDS PostgreSQL Deployment Guide

This guide explains how to deploy Seentics Cloud with AWS RDS PostgreSQL instead of the local TimescaleDB container.

## Overview

We've migrated from TimescaleDB to PostgreSQL to enable hosting on AWS RDS. This provides:
- **Managed Database**: AWS handles backups, updates, and maintenance
- **High Availability**: Multi-AZ deployments for production
- **Scalability**: Easy vertical and horizontal scaling
- **Security**: VPC isolation and encryption at rest/in transit

## Prerequisites

- AWS Account with appropriate permissions
- AWS CLI configured
- Docker and Docker Compose installed
- Seentics Cloud codebase

## Step 1: Create AWS RDS PostgreSQL Instance

### Using AWS Console

1. **Navigate to RDS Console**
   - Go to AWS Console → RDS → Create database

2. **Database Configuration**
   ```
   Engine: PostgreSQL
   Version: 15.x (latest stable)
   Template: Production (for prod) or Dev/Test (for staging)
   ```

3. **Instance Configuration**
   ```
   DB Instance Class: 
   - Development: db.t3.micro (free tier)
   - Production: db.t3.medium or larger
   
   Storage:
   - Type: gp3 (General Purpose SSD)
   - Size: 100 GB minimum (auto-scaling enabled)
   ```

4. **Database Settings**
   ```
   DB Instance Identifier: seentics-analytics-db
   Master Username: seentics
   Master Password: <strong-password>
   ```

5. **Connectivity**
   ```
   VPC: Default or custom VPC
   Subnet Group: Default
   Public Access: Yes (for development) / No (for production)
   Security Group: Create new or use existing
   ```

6. **Additional Configuration**
   ```
   Initial Database Name: seentics_analytics
   Backup Retention: 7 days (minimum)
   Monitoring: Enable Enhanced Monitoring
   ```

### Using AWS CLI

```bash
# Create RDS instance
aws rds create-db-instance \
    --db-instance-identifier seentics-analytics-db \
    --db-instance-class db.t3.medium \
    --engine postgres \
    --engine-version 15.4 \
    --master-username seentics \
    --master-user-password YOUR_STRONG_PASSWORD \
    --allocated-storage 100 \
    --storage-type gp3 \
    --storage-encrypted \
    --db-name seentics_analytics \
    --vpc-security-group-ids sg-xxxxxxxxx \
    --backup-retention-period 7 \
    --multi-az \
    --publicly-accessible
```

## Step 2: Configure Security Group

Create or update security group to allow PostgreSQL connections:

```bash
# Create security group
aws ec2 create-security-group \
    --group-name seentics-rds-sg \
    --description "Security group for Seentics RDS PostgreSQL"

# Add inbound rule for PostgreSQL (port 5432)
aws ec2 authorize-security-group-ingress \
    --group-id sg-xxxxxxxxx \
    --protocol tcp \
    --port 5432 \
    --source-group sg-xxxxxxxxx  # Application security group
```

## Step 3: Update Environment Configuration

### Development Environment (.env)

```bash
# PostgreSQL (Analytics Service) - AWS RDS
DATABASE_URL=postgres://seentics:YOUR_PASSWORD@seentics-analytics-db.xxxxxxxxx.us-east-1.rds.amazonaws.com:5432/seentics_analytics?sslmode=require

# Other configurations remain the same
MONGODB_URI=mongodb://seentics:seentics_mongo_pass@localhost:27017/seentics?authSource=admin
REDIS_URL=redis://:seentics_redis_pass@localhost:6379
```

### Production Environment (.env.production)

```bash
# PostgreSQL (Analytics Service) - AWS RDS
POSTGRES_USER=seentics
POSTGRES_PASSWORD=YOUR_STRONG_PASSWORD
POSTGRES_DB=seentics_analytics
POSTGRES_HOST=seentics-analytics-db.xxxxxxxxx.us-east-1.rds.amazonaws.com
ANALYTICS_DATABASE_URL=postgres://${POSTGRES_USER}:${POSTGRES_PASSWORD}@${POSTGRES_HOST}:5432/${POSTGRES_DB}?sslmode=require
```

## Step 4: Update Docker Compose

### For Development (docker-compose.yml)

Remove the PostgreSQL service since we're using RDS:

```yaml
version: "3.9"

services:
  api-gateway:
    # ... existing configuration
    
  user-service:
    # ... existing configuration
    
  analytics-service:
    build: ./services/analytics
    container_name: analytics-service
    ports:
      - "3002:3002"
    env_file: .env
    environment:
      - ENVIRONMENT
      - PORT=3002
      - DATABASE_URL=${DATABASE_URL}  # Points to RDS
      - LOG_LEVEL=info
      - JWT_SECRET=${JWT_SECRET}
      - REDIS_URL=redis://:seentics_redis_pass@redis:6379
      - REDIS_PASSWORD=seentics_redis_pass
      - GLOBAL_API_KEY=${GLOBAL_API_KEY}
      - USER_SERVICE_URL=http://user-service:3001
      - MAXMIND_LICENSE_KEY=${MAXMIND_LICENSE_KEY}
      - MAXMIND_DB_PATH=/data/GeoLite2-City.mmdb
    volumes:
      - maxmind-data:/data
    depends_on:
      redis:
        condition: service_healthy
    restart: unless-stopped
    networks:
      - seentics-network

  # Remove postgres service - using RDS instead
  
  mongodb:
    # ... existing configuration
    
  redis:
    # ... existing configuration

volumes:
  mongo-data:
  redis-data:
  maxmind-data:
  # Remove postgres-data volume

networks:
  seentics-network:
    driver: bridge
```

## Step 5: Run Database Migrations

### Option 1: Using Migration Script

Update the migration script to connect to RDS:

```bash
#!/bin/bash

# Set RDS connection details
export PGHOST="seentics-analytics-db.xxxxxxxxx.us-east-1.rds.amazonaws.com"
export PGPORT="5432"
export PGUSER="seentics"
export PGPASSWORD="YOUR_PASSWORD"
export PGDATABASE="seentics_analytics"

echo "🚀 Running PostgreSQL migrations on AWS RDS"

# Run migration
psql -f migrations/000001_initial_schema_postgresql.up.sql

echo "✅ Migrations completed successfully!"
```

### Option 2: Using psql directly

```bash
psql -h seentics-analytics-db.xxxxxxxxx.us-east-1.rds.amazonaws.com \
     -U seentics \
     -d seentics_analytics \
     -f migrations/000001_initial_schema_postgresql.up.sql
```

## Step 6: Test Connection

Test the connection from your application:

```bash
# Test connection
psql -h seentics-analytics-db.xxxxxxxxx.us-east-1.rds.amazonaws.com \
     -U seentics \
     -d seentics_analytics \
     -c "SELECT version();"
```

## Step 7: Production Deployment Considerations

### SSL/TLS Configuration

Always use SSL in production:

```bash
DATABASE_URL=postgres://seentics:PASSWORD@HOST:5432/seentics_analytics?sslmode=require
```

### Connection Pooling

For production, consider using connection pooling:

```bash
# In your application configuration
MAX_CONNECTIONS=100
MIN_CONNECTIONS=25
CONNECTION_LIFETIME=2h
IDLE_TIMEOUT=15m
```

### Monitoring and Alerts

Set up CloudWatch monitoring:

1. **Database Metrics**
   - CPU Utilization
   - Database Connections
   - Read/Write IOPS
   - Free Storage Space

2. **Custom Metrics**
   - Query performance
   - Connection pool usage
   - Error rates

### Backup Strategy

Configure automated backups:

```bash
# Enable automated backups
aws rds modify-db-instance \
    --db-instance-identifier seentics-analytics-db \
    --backup-retention-period 30 \
    --preferred-backup-window "03:00-04:00" \
    --preferred-maintenance-window "sun:04:00-sun:05:00"
```

## Step 8: Cost Optimization

### Development Environment

- Use `db.t3.micro` (free tier eligible)
- Single AZ deployment
- Minimal storage with auto-scaling

### Production Environment

- Use appropriate instance size based on load
- Enable Multi-AZ for high availability
- Use Reserved Instances for cost savings
- Set up storage auto-scaling

## Troubleshooting

### Common Issues

1. **Connection Timeout**
   ```bash
   # Check security group rules
   aws ec2 describe-security-groups --group-ids sg-xxxxxxxxx
   ```

2. **SSL Connection Errors**
   ```bash
   # Download RDS CA certificate
   wget https://s3.amazonaws.com/rds-downloads/rds-ca-2019-root.pem
   ```

3. **Performance Issues**
   ```bash
   # Check connection pool settings
   # Monitor CloudWatch metrics
   # Consider read replicas for read-heavy workloads
   ```

### Health Check Updates

The health check endpoint will now show PostgreSQL information instead of TimescaleDB:

```json
{
  "status": "healthy",
  "database": "connected",
  "postgres": "available",
  "partitions": true,
  "timestamp": "2024-01-01T00:00:00Z",
  "version": "1.0.0"
}
```

## Migration Checklist

- [ ] Create AWS RDS PostgreSQL instance
- [ ] Configure security groups
- [ ] Update environment variables
- [ ] Remove PostgreSQL service from Docker Compose
- [ ] Run database migrations
- [ ] Test application connectivity
- [ ] Update monitoring and alerts
- [ ] Configure backups
- [ ] Document connection details securely

## Cost Estimation

### Development
- **db.t3.micro**: ~$13/month (free tier eligible)
- **Storage (20GB)**: ~$2.3/month
- **Total**: ~$15/month

### Production
- **db.t3.medium**: ~$62/month
- **Storage (100GB)**: ~$11.5/month
- **Multi-AZ**: Additional 100% cost
- **Total**: ~$147/month

## Next Steps

1. **Set up monitoring**: Configure CloudWatch dashboards
2. **Implement read replicas**: For read-heavy workloads
3. **Configure automated scaling**: Based on metrics
4. **Set up disaster recovery**: Cross-region backups
5. **Security hardening**: VPC endpoints, encryption

For questions or issues, refer to the [AWS RDS Documentation](https://docs.aws.amazon.com/rds/) or contact the development team.
