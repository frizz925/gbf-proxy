variable "master_image" {
  type = "string"
  default = "coreos-stable"
}

variable "master_size" {
  type = "string"
  default = "s-2vcpu-2gb"
}

variable "master_name" {
  type = "string"
  default = "gbf-proxy-master"
}

resource "digitalocean_droplet" "master" {
  name = "${var.master_name}"
  image = "${var.master_image}"
  region = "${var.do_region}"
  size = "${var.master_size}"
  ssh_keys = [
    "${var.ssh_fingerprint}"
  ]
  private_networking = true

  connection {
    type = "ssh"
    user = "core"
    private_key = "${file(var.ssh_private_key)}"
    timeout = "2m"
  }

  provisioner "remote-exec" {
    scripts = [
      "../../kubernetes/scripts/setup.sh",
      "../../kubernetes/scripts/master-setup.sh",
      "../../kubernetes/scripts/network-setup.sh"
    ]
  }

  provisioner "remote-exec" {
    script = "../../kubernetes/scripts/teardown.sh"
    when = "destroy"
  }
}

resource "digitalocean_firewall" "master" {
  name = "gbf-proxy"
  droplet_ids = ["${digitalocean_droplet.master.id}"]

  inbound_rule = [
    {
      protocol = "tcp"
      port_range = "22"
    },
    {
      protocol = "tcp"
      port_range = "80"
    },
    {
      protocol = "tcp"
      port_range = "443"
    },
    {
      protocol = "tcp"
      port_range = "8088"
    },
    {
      protocol = "icmp"
    }
  ]
}

resource "digitalocean_floating_ip" "master" {
  droplet_id = "${digitalocean_droplet.master.id}"
  region = "${digitalocean_droplet.master.region}"
}
