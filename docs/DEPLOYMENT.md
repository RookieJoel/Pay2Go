# Deployment Guide

## Pay2Go - Production Deployment Guide

This guide covers deploying Pay2Go to production environments.

---

## Prerequisites

- Docker and Docker Compose installed
- PostgreSQL 15+ (or use Docker Compose)
- Go 1.21+ (for building from source)
- Make (optional, for automation)
- SSL/TLS certificates (for HTTPS)

---

## Environment Configuration

### 1. Create Production Environment File

Copy the example environment file and update with production values:

```bash
cp .env.example .env
```

Edit `.env` with production settings:

```bash
# Server Configuration
SERVER_PORT=8080
SERVER_HOST=0.0.0.0

# Database Configuration (use strong password)
DB_HOST=your-db-host
DB_PORT=5432
DB_USER=pay2go_user
DB_PASSWORD=strong-random-password-here
DB_NAME=pay2go_production
DB_SSLMODE=require

# Security (generate strong random secret)
JWT_SECRET=your-super-secret-jwt-key-change-this

# Payment Gateways (production keys)
STRIPE_API_KEY=sk_live_...
PAYPAL_CLIENT_ID=...
PAYPAL_CLIENT_SECRET=...
```

### 2. Generate Secure Secrets

```bash
# Generate JWT secret
openssl rand -base64 32

# Generate database password
openssl rand -base64 24
```

---

## Deployment Methods

### Method 1: Docker Compose (Recommended for Small Scale)

#### 1. Build and Start Services

```bash
# Start database
docker-compose up -d

# Build application Docker image
docker build -t pay2go:latest .

# Run application container
docker run -d \
  --name pay2go-api \
  --env-file .env \
  -p 8080:8080 \
  --link pay2go-postgres:postgres \
  pay2go:latest
```

#### 2. Run Database Migrations

```bash
# Install golang-migrate
brew install golang-migrate  # macOS
# or
curl -L https://github.com/golang-migrate/migrate/releases/download/v4.15.2/migrate.linux-amd64.tar.gz | tar xvz
sudo mv migrate /usr/local/bin/

# Run migrations
migrate -path ./migrations \
  -database "postgres://user:password@localhost:5432/pay2go_production?sslmode=require" \
  up
```

#### 3. Verify Deployment

```bash
# Health check
curl http://localhost:8080/api/v1/health

# Expected response:
# {"status":"ok"}
```

---

### Method 2: Kubernetes (Recommended for Production Scale)

#### 1. Create Kubernetes Manifests

**deployment.yaml**:
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: pay2go-api
  labels:
    app: pay2go
spec:
  replicas: 3
  selector:
    matchLabels:
      app: pay2go
  template:
    metadata:
      labels:
        app: pay2go
    spec:
      containers:
      - name: pay2go
        image: your-registry/pay2go:latest
        ports:
        - containerPort: 8080
        envFrom:
        - secretRef:
            name: pay2go-secrets
        livenessProbe:
          httpGet:
            path: /api/v1/health/live
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /api/v1/health/ready
            port: 8080
          initialDelaySeconds: 10
          periodSeconds: 5
        resources:
          requests:
            memory: "128Mi"
            cpu: "250m"
          limits:
            memory: "512Mi"
            cpu: "1000m"
```

**service.yaml**:
```yaml
apiVersion: v1
kind: Service
metadata:
  name: pay2go-service
spec:
  type: LoadBalancer
  selector:
    app: pay2go
  ports:
  - protocol: TCP
    port: 80
    targetPort: 8080
```

**secrets.yaml**:
```yaml
apiVersion: v1
kind: Secret
metadata:
  name: pay2go-secrets
type: Opaque
stringData:
  DB_HOST: "postgres-service"
  DB_PORT: "5432"
  DB_USER: "pay2go_user"
  DB_PASSWORD: "your-secure-password"
  DB_NAME: "pay2go_production"
  JWT_SECRET: "your-jwt-secret"
```

#### 2. Deploy to Kubernetes

```bash
# Create namespace
kubectl create namespace pay2go

# Apply configurations
kubectl apply -f k8s/secrets.yaml -n pay2go
kubectl apply -f k8s/deployment.yaml -n pay2go
kubectl apply -f k8s/service.yaml -n pay2go

# Check deployment status
kubectl get pods -n pay2go
kubectl get services -n pay2go
```

---

### Method 3: Cloud Platforms

#### AWS Elastic Beanstalk

```bash
# Install EB CLI
pip install awsebcli

# Initialize EB
eb init -p docker pay2go-api

# Create environment
eb create production-env

# Deploy
eb deploy
```

#### Google Cloud Run

```bash
# Build and push image
gcloud builds submit --tag gcr.io/PROJECT-ID/pay2go

# Deploy
gcloud run deploy pay2go-api \
  --image gcr.io/PROJECT-ID/pay2go \
  --platform managed \
  --region us-central1 \
  --allow-unauthenticated
```

#### Heroku

```bash
# Login to Heroku
heroku login

# Create app
heroku create pay2go-api

# Add PostgreSQL addon
heroku addons:create heroku-postgresql:hobby-dev

# Deploy
git push heroku main

# Run migrations
heroku run migrate -path ./migrations -database $DATABASE_URL up
```

---

## Database Setup

### Production Database Configuration

#### PostgreSQL Configuration (`postgresql.conf`)

```conf
# Connection Settings
max_connections = 200
shared_buffers = 256MB

# Write Ahead Log
wal_level = replica
max_wal_size = 1GB
min_wal_size = 80MB

# Query Tuning
effective_cache_size = 1GB
random_page_cost = 1.1

# Logging
log_statement = 'mod'
log_min_duration_statement = 1000
```

#### Create Database and User

```sql
-- Connect as postgres superuser
CREATE DATABASE pay2go_production;

CREATE USER pay2go_user WITH ENCRYPTED PASSWORD 'strong-password';

GRANT ALL PRIVILEGES ON DATABASE pay2go_production TO pay2go_user;

-- Grant schema permissions
\c pay2go_production
GRANT ALL ON SCHEMA public TO pay2go_user;
GRANT ALL ON ALL TABLES IN SCHEMA public TO pay2go_user;
GRANT ALL ON ALL SEQUENCES IN SCHEMA public TO pay2go_user;
```

#### Run Migrations

```bash
make migrate-up
# or
migrate -path ./migrations \
  -database "postgres://pay2go_user:password@host:5432/pay2go_production?sslmode=require" \
  up
```

---

## SSL/TLS Configuration

### Using Nginx as Reverse Proxy

**nginx.conf**:
```nginx
upstream pay2go_backend {
    server localhost:8080;
}

server {
    listen 80;
    server_name api.pay2go.com;
    return 301 https://$server_name$request_uri;
}

server {
    listen 443 ssl http2;
    server_name api.pay2go.com;

    ssl_certificate /etc/ssl/certs/pay2go.crt;
    ssl_certificate_key /etc/ssl/private/pay2go.key;
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers HIGH:!aNULL:!MD5;

    location / {
        proxy_pass http://pay2go_backend;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

---

## Monitoring and Logging

### Application Logging

Logs are written to stdout. Configure log aggregation:

**Using ELK Stack**:
```bash
# Install Filebeat
curl -L -O https://artifacts.elastic.co/downloads/beats/filebeat/filebeat-8.0.0-linux-x86_64.tar.gz
tar xzvf filebeat-8.0.0-linux-x86_64.tar.gz

# Configure Filebeat to ship logs to Elasticsearch
```

**Using CloudWatch (AWS)**:
```bash
# Install CloudWatch agent
wget https://s3.amazonaws.com/amazoncloudwatch-agent/ubuntu/amd64/latest/amazon-cloudwatch-agent.deb
sudo dpkg -i -E ./amazon-cloudwatch-agent.deb
```

### Health Monitoring

Set up monitoring for these endpoints:

- `GET /api/v1/health` - Overall health
- `GET /api/v1/health/ready` - Readiness check
- `GET /api/v1/health/live` - Liveness check

**Example Prometheus Scrape Config**:
```yaml
scrape_configs:
  - job_name: 'pay2go'
    metrics_path: '/api/v1/metrics'
    static_configs:
      - targets: ['localhost:8080']
```

---

## Security Checklist

- [ ] Use HTTPS/TLS for all API traffic
- [ ] Set strong `JWT_SECRET` (min 32 characters)
- [ ] Use PostgreSQL with SSL mode enabled (`sslmode=require`)
- [ ] Store secrets in environment variables or secret management service
- [ ] Enable database connection encryption
- [ ] Configure firewall to restrict database access
- [ ] Implement rate limiting (already configured in middleware)
- [ ] Set up API key rotation policy
- [ ] Enable audit logging for all transactions
- [ ] Configure CORS appropriately
- [ ] Use prepared statements (already implemented)
- [ ] Regular security updates and patches
- [ ] Enable database backup and recovery

---

## Backup and Recovery

### Database Backups

**Automated Daily Backup**:
```bash
#!/bin/bash
# backup.sh
DATE=$(date +%Y%m%d_%H%M%S)
BACKUP_DIR="/backups"
DB_NAME="pay2go_production"

pg_dump -h localhost -U pay2go_user $DB_NAME | gzip > $BACKUP_DIR/backup_$DATE.sql.gz

# Keep only last 30 days
find $BACKUP_DIR -name "backup_*.sql.gz" -mtime +30 -delete
```

**Schedule with Cron**:
```bash
# Run daily at 2 AM
0 2 * * * /opt/scripts/backup.sh
```

### Restore from Backup

```bash
# Restore database
gunzip < backup_20240115_020000.sql.gz | psql -h localhost -U pay2go_user pay2go_production
```

---

## Scaling Considerations

### Horizontal Scaling

The application is stateless and can be horizontally scaled:

```bash
# Kubernetes example
kubectl scale deployment pay2go-api --replicas=5 -n pay2go
```

### Database Connection Pooling

Configure connection pool in production:

```go
// In config.go, add:
db.SetMaxOpenConns(25)
db.SetMaxIdleConns(5)
db.SetConnMaxLifetime(5 * time.Minute)
```

### Caching Layer (Future Enhancement)

Consider adding Redis for:
- Rate limiting state
- API response caching
- Session management

---

## Performance Optimization

### Database Indexes

Indexes are already defined in migrations:

```sql
CREATE INDEX idx_transactions_partner_id ON transactions(partner_id);
CREATE INDEX idx_transactions_status ON transactions(status);
CREATE INDEX idx_transactions_idempotency_key ON transactions(idempotency_key);
```

### Application Tuning

**Optimize Fiber App**:
```go
app := fiber.New(fiber.Config{
    Prefork:       true,  // Use SO_REUSEPORT socket option
    CaseSensitive: true,
    StrictRouting: false,
    ServerHeader:  "Pay2Go",
    AppName:       "Pay2Go API v1.0",
})
```

---

## Troubleshooting

### Common Issues

**Database Connection Failed**:
```bash
# Check database is running
docker ps | grep postgres

# Test connection
psql -h localhost -p 5432 -U pay2go_user -d pay2go_production

# Check logs
docker logs pay2go-postgres
```

**API Returns 500 Errors**:
```bash
# Check application logs
docker logs pay2go-api

# Check database connectivity
curl http://localhost:8080/api/v1/health/ready
```

**Migration Failures**:
```bash
# Check current migration version
migrate -path ./migrations \
  -database "postgres://..." \
  version

# Force version if dirty
migrate -path ./migrations \
  -database "postgres://..." \
  force VERSION
```

---

## Rollback Procedure

### Application Rollback

```bash
# Kubernetes
kubectl rollout undo deployment/pay2go-api -n pay2go

# Docker
docker stop pay2go-api
docker run -d --name pay2go-api pay2go:previous-version
```

### Database Rollback

```bash
# Rollback last migration
migrate -path ./migrations \
  -database "postgres://..." \
  down 1
```

---

## Maintenance

### Zero-Downtime Deployment

```bash
# Kubernetes rolling update
kubectl set image deployment/pay2go-api \
  pay2go=pay2go:v1.1 \
  --record \
  -n pay2go

# Monitor rollout
kubectl rollout status deployment/pay2go-api -n pay2go
```

### Database Maintenance

```bash
# Vacuum database
psql -h localhost -U pay2go_user -d pay2go_production -c "VACUUM ANALYZE;"

# Reindex
psql -h localhost -U pay2go_user -d pay2go_production -c "REINDEX DATABASE pay2go_production;"
```

---

## Support and Documentation

- **API Documentation**: See `/docs/API.md`
- **Architecture**: See `/docs/ARCHITECTURE.md`
- **Database Schema**: See `/docs/DATABASE_DESIGN.md`

For production support, contact DevOps team at: devops@pay2go.com
