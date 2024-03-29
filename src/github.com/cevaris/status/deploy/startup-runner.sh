set -ex

# Talk to the metadata server to get the project id and location of application binary.
PROJECTID=$(curl -s "http://metadata.google.internal/computeMetadata/v1/project/project-id" -H "Metadata-Flavor: Google")
DEPLOY_LOCATION=$(curl -s "http://metadata.google.internal/computeMetadata/v1/instance/attributes/app-location" -H "Metadata-Flavor: Google")

# Install logging monitor. The monitor will automatically pickup logs send to
# syslog.
curl -s "https://storage.googleapis.com/signals-agents/logging/google-fluentd-install.sh" | bash
service google-fluentd restart &

# Install dependencies from apt
apt-get update
apt-get install -yq ca-certificates supervisor less

# Get the application tar from the GCS bucket.
gsutil cp $DEPLOY_LOCATION /app.tar
mkdir -p /app
tar -x -f /app.tar -C /app
chmod +x /app/runner

# Create a goapp user. The application will run as this user.
getent passwd goapp || /usr/sbin/useradd -m -d /home/goapp goapp
chown -R goapp:goapp /app

# Configure supervisor to run the Go app.
cat >/etc/supervisor/conf.d/goapp.conf << EOF
[program:goapp]
directory=/app
command=/app/runner
autostart=true
autorestart=true
user=goapp
environment=HOME="/home/goapp",USER="goapp"
stdout_logfile=syslog
stderr_logfile=syslog
EOF

supervisorctl reread
supervisorctl update

# Application should now be running under supervisor