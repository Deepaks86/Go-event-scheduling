provider "google" {
  credentials = file("/Users/deepak/Deepak/GoCodes/stackgenAddons/service-account.json")  # Path to your GCP service account key JSON
  project     = "steady-course-447013-r6"
  region      = "asia-south1"
  zone        = "asia-south1-c"
}

resource "google_compute_instance" "default" {
  name         = "event-scheduler-vm"
  machine_type = "e2-medium"
  zone         = "asia-south1-c"
  tags         = ["event-scheduler-vm"]

  boot_disk {
    initialize_params {
      image = "ubuntu-2404-noble-amd64-v20250313"
    }
  }

metadata_startup_script = <<-EOF
#!/bin/bash
# Update the system
sudo apt-get update -y
sudo apt-get upgrade -y

# Install Docker runtime
curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo apt-key add -
sudo add-apt-repository "deb [arch=amd64] https://download.docker.com/linux/ubuntu $(lsb_release -cs) stable"
sudo apt-get update -y
sudo apt-get install -y docker-ce docker-ce-cli containerd.io -y
sudo systemctl start docker
sudo systemctl enable docker
sudo usermod -aG docker $USER
# Install Docker-compose
sudo curl -L "https://github.com/docker/compose/releases/download/1.29.2/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose 
sudo chmod +x /usr/local/bin/docker-compose

# Clone the GitHub repository (replace with your actual repo URL)
cd /home/ubuntu
git clone https://github.com/Deepaks86/Go-event-scheduling.git

# Navigate to the directory where the startup script is
cd /home/ubuntu/Go-event-scheduling/docker

# Run startup.sh script
chmod +x startup.sh
./startup.sh
    ./startup.sh
  EOF

  network_interface {
    network = "default"
    access_config {
    }
  }

}

resource "google_compute_firewall" "allow-8080" {
  name    = "allow-8080"
  network = "default"

  allow {
    protocol = "tcp"
    ports    = ["8080"]
  }
  target_tags = ["event-scheduler-vm"]
  source_ranges = ["0.0.0.0/0"]  # Open to all IPs (public internet)
}

