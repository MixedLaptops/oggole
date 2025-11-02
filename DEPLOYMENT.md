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

### 1. Local: Initialize Database

```bash
cd src
go run init_db.go
go run seed_data.go
```

### 2. VM: Run Setup

```bash
# Upload and run setup script
scp vm-setup.sh user@your-vm:~/
ssh user@your-vm './vm-setup.sh'

# Upload database
scp src/whoknows.db user@your-vm:/var/lib/oggole/data/oggole.db

# Upload configs
scp docker-compose.prod.yml user@your-vm:~/oggole/
scp deploy.sh user@your-vm:~/oggole/
scp nginx/nginx.conf user@your-vm:~/oggole/nginx/
```

### 3. VM: First Deploy

```bash
ssh user@your-vm
cd ~/oggole
./deploy.sh
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
sudo cp /var/lib/oggole/data/oggole.db /var/lib/oggole/data/backup.db
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
- `~/oggole/data/oggole.db` - Database (never in Docker image!)

---

## Troubleshooting

**Container won't start?**
```bash
docker compose -f docker-compose.prod.yml logs
```

**Database missing?**
```bash
ls -la /var/lib/oggole/data/oggole.db
```

**Start over?**
```bash
docker compose -f docker-compose.prod.yml down
./deploy.sh
```
