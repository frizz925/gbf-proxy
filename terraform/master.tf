resource "vultr_server" "master" {
  plan_id           = "${data.vultr_plan.gp_1vcpu_1gb.id}"
  region_id         = "${data.vultr_region.sgp.id}"
  os_id             = "${data.vultr_os.debian_10.id}"
  ssh_key_ids       = ["${vultr_ssh_key.main.id}"]
  firewall_group_id = "${vultr_firewall_group.master.id}"

  ddos_protection         = false
  enable_ipv6             = false
  enable_private_network  = true
  notify_activate         = true

  hostname  = "gbf-proxy"
  tag       = "app:gbf-proxy role:master"
  label     = "Granblue Proxy"
}

resource "vultr_firewall_group" "master" {
  description = "Firewall group for Granblue Proxy"
}

resource "vultr_firewall_rule" "master_ssh_22_bastion" {
  protocol  = "tcp"
  network   = "${var.bastion_cidr}"
  from_port = "22"

  firewall_group_id = "${vultr_firewall_group.master.id}"
}

resource "vultr_firewall_rule" "master_http_80_all" {
  protocol  = "tcp"
  network   = "0.0.0.0/0"
  from_port = "80"

  firewall_group_id = "${vultr_firewall_group.master.id}"
}

resource "vultr_firewall_rule" "master_https_443_all" {
  protocol  = "tcp"
  network   = "0.0.0.0/0"
  from_port = "443"

  firewall_group_id = "${vultr_firewall_group.master.id}"
}

resource "vultr_firewall_rule" "master_http_8088_all" {
  protocol  = "tcp"
  network   = "0.0.0.0/0"
  from_port = "8088"

  firewall_group_id = "${vultr_firewall_group.master.id}"
}
