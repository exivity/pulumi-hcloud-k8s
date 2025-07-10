packer {
  required_plugins {
    hcloud = {
      source  = "github.com/hetznercloud/hcloud"
      version = "~> 1"
    }
  }
}

variable "talos_image_id" {
  type    = string
}

variable "talos_version" {
  type    = string
}

variable "arch" {
  type    = string
}

variable "server_type" {
  type    = string
}

variable "server_location" {
  type    = string
}

variable "snapshot_name" {
  type    = string
}

locals {
  image = "https://factory.talos.dev/image/${var.talos_image_id}/${var.talos_version}/hcloud-${var.arch}.raw.xz"
}

source "hcloud" "talos" {
  rescue       = "linux64"
  image        = "debian-11"
  location     = "${var.server_location}"
  server_type  = "${var.server_type}"
  ssh_username = "root"

  snapshot_name = "${var.snapshot_name}"
  snapshot_labels = {
    type    = "infra",
    os      = "talos",
    version = "${var.talos_version}",
    arch    = "${var.arch}",
  }
}

build {
  sources = ["source.hcloud.talos"]

  provisioner "shell" {
    inline = [
      "apt-get install -y wget",
      "wget -O /tmp/talos.raw.xz ${local.image}",
      "xz -d -c /tmp/talos.raw.xz | dd of=/dev/sda && sync",
    ]
  }
}
