#!/bin/bash

set -eu

# ============================
# Variables
# ============================

# Set the timezone for the server
TIMEZONE=America/New_Yourk

# Prompt to enter a password for PostgreSQL user
USERNAME=greenlight
read -p "Enter password for greenlight D user: " DB_PASSWORD

export LC_ALL=en_US.UTF-8

# ============================
# Script Logic
# ============================

# Enable the universe repository
add-apt-repository --yes universe

# Update all software packages
# --force-confnew -> update configuration files if newer ones available
apt update
apt --yes -o Dpkg::Options::="--force-confnew" upgrade

# Set the system timezone and install all locales
timedatectl set-timezone ${TIMEZONE}
apt --yes install locales-all

# Add the new user, and give him sudo privileges
useradd --create-hom --shell "/bin/bash" --groups sudo "${USERNAME}"

# Force a password to be set the first time they log in
passwd --delete "${USERNAME}"
chage --lastday 0 "${USERNAME}"

# Copy the SSH keys from the root to the new user
rsync --archive --chown=${USERNAME}:${USERNAME} /root/.ssh /home/${USERNAME}

# Configure the firewall to allow SSH, HTTP, and HTTPS traffic
ufw allow 22
ufw allow 80/tcp
ufw allow 443/tcp
ufw --force enable

# Install fail2ban
apt --yes install fail2ban

# Install the migrate CLI tool
curl -L https://github.com/golang-migrate/migrate/releases/download/v4.14.1/migrate.linux-amd64.tar.gz | tar xvz
mv migrate.linux-amd64 /usr/local/bin/migrate

# Install PostgreSQL
apt --yes install postgresql

# Set up the greenlight DB and create a user with the password entered earlier
sudo -i -u postgres psql -c "CREATE DATABASE greenlight"
sudo -i -u postgres psql -d greenlight -c "CREATE EXTENSION IF NOT EXISTS citext"
sudo -i -u postgres psql -d greenlight -c "CREATE ROLE greenlight WITH LOGIN PASSWORD '${DB_PASSWORD}'"

# Add a DSN for connecting to the greenlight DB to the system-wide environment
echo "GREENLIGHT_DB_DSN='postgres://greenlight:${DB_PASSWORD}@localhost/greenlight'" >> /etc/environment

# Install Caddy (see https://caddyserver.com/docs/install#debian-ubuntu-raspbian)
apt --yes install debian-keyring debian-archive-keyring apt-transport-https
curl -L https://dl.cloudsmith.io/public/caddy/stable/gpg.key | sudo apt-key add -
curl -L https://dl.cloudsmith.io/public/caddy/stable/debian.deb.txt | sudo tee -a /etc/apt/sources.list.d/caddy-stable.list
apt update
apt --yes install caddy

echo 'Script complete!'
echo 'Reboting...'
sleep 3 ; reboot