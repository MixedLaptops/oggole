# Oggole Deployment Guide

## How It Works

```
You push to GitHub
    ↓
GitHub Actions builds image (docker-publish.yml)
    ↓
Image pushed to ghcr.io
    ↓
GitHub Actions deploys to VM (deploy-to-vm.yml)
    ↓
VM pulls image and runs it
```

**Fully Automated**: Push to main → Auto-builds → Auto-deploys to VM ✅
**Security**: Database stays on VM, never in public image ✅

---

## First Time Setup

### 0. GitHub Secrets Setup

Add these secrets to your GitHub repository (Settings → Secrets and variables → Actions):

```
SSH_PRIVATE_KEY = Your VM SSH private key
SSH_USER = Your VM username (e.g., ubuntu, azureuser)
SSH_HOST = Your VM IP address or hostname
```

**Generate SSH key pair** (if you don't have one):
```bash
# On local machine
ssh-keygen -t ed25519 -C "github-actions" -f ~/.ssh/vm_deploy_key
# Copy public key to VM
ssh-copy-id -i ~/.ssh/vm_deploy_key.pub user@your-vm-ip
# Add private key to GitHub secrets
cat ~/.ssh/vm_deploy_key  # Copy this to SSH_PRIVATE_KEY secret
```

### 1. VM: Run Setup

```bash
# Upload and run setup script
scp vm-setup.sh user@your-vm:~/
ssh user@your-vm './vm-setup.sh'

# Upload configs
scp docker-compose.prod.yml user@your-vm:~/oggole/
scp deploy.sh user@your-vm:~/oggole/
scp nginx/nginx.conf user@your-vm:~/oggole/nginx/
ssh user@your-vm 'chmod +x ~/oggole/deploy.sh'
```

### 2. VM: Set PostgreSQL Password (Optional)

```bash
ssh user@your-vm
echo "export POSTGRES_PASSWORD=your_secure_password" >> ~/.bashrc
source ~/.bashrc
```

If not set, defaults to 'oggole' (fine for development, but use a strong password for production!)

### 3. VM: First Deploy

```bash
ssh user@your-vm
cd ~/oggole
./deploy.sh
```

PostgreSQL will initialize automatically on first run!

### 4. Initialize Database Schema

After containers are running, initialize the database:

```bash
ssh user@your-vm
docker exec oggole-app /app/oggole init-db
docker exec oggole-app /app/oggole seed-data
```

Done! Visit `http://your-vm-ip`

---

## Updating (Fully Automated!)

Just push to main - everything happens automatically:

```bash
git add .
git commit -m "your changes"
git push origin main
```

What happens automatically:
1. ✅ GitHub Actions builds Docker image
2. ✅ Pushes to ghcr.io
3. ✅ SSHs to VM
4. ✅ Pulls latest image
5. ✅ Restarts containers

Check progress: `https://github.com/your-username/oggole/actions`

### Manual Deploy (Optional)
If you want to deploy manually instead:
```bash
ssh user@your-vm
cd ~/oggole
./deploy.sh
```

---

## Common Tasks

### Backup Database
```bash
ssh user@your-vm
cd ~/oggole
# Create a PostgreSQL backup
docker exec oggole-postgres pg_dump -U oggole oggole > backup_$(date +%Y%m%d_%H%M%S).sql

# Or backup entire PostgreSQL data directory
sudo tar -czf postgres_backup_$(date +%Y%m%d_%H%M%S).tar.gz ~/oggole/postgres/
```

### Restore Database
```bash
ssh user@your-vm
cd ~/oggole
# Restore from SQL backup
cat your_backup.sql | docker exec -i oggole-postgres psql -U oggole oggole
```

### View Logs
```bash
ssh user@your-vm
cd ~/oggole
docker compose -f docker-compose.prod.yml logs -f
```

### Restart Services
```bash
ssh user@your-vm
cd ~/oggole
docker compose -f docker-compose.prod.yml restart
```

### Stop Everything
```bash
ssh user@your-vm
cd ~/oggole
docker compose -f docker-compose.prod.yml down
```

---

## Files

**GitHub Workflows** (automated):
- `.github/workflows/docker-publish.yml` - Builds & publishes to ghcr.io
- `.github/workflows/deploy-to-vm.yml` - Auto-deploys to VM after build

**Local Machine**:
- `docker-compose.prod.yml` - Production configuration
- `deploy.sh` - Manual deployment script (optional)
- `vm-setup.sh` - Initial VM setup

**VM**:
- `~/oggole/` - App files
- `~/oggole/postgres/` - PostgreSQL data directory (never in Docker image!)

---

## Troubleshooting

**Container won't start?**
```bash
docker compose -f docker-compose.prod.yml logs
```

**Database not connecting?**
```bash
# Check if PostgreSQL is running
docker compose -f docker-compose.prod.yml ps postgres

# Check PostgreSQL logs
docker compose -f docker-compose.prod.yml logs postgres
```

**Start over?**
```bash
docker compose -f docker-compose.prod.yml down
./deploy.sh
```
