resource "vultr_ssh_key" "gbf-proxy" {
  name = "Granblue Proxy Public Key"
  public_key  = "${file("~/.ssh/gbf-proxy_id_rsa.pub")}"
}

resource "vultr_instance" "gbf-proxy" {
  name              = "Granblue Proxy"
  region_id         = "${var.region_id}"
  plan_id           = "${var.plan_id}"
  os_id             = "${var.os_id}"
  ssh_key_ids       = ["${vultr_ssh_key.gbf-proxy.id}"]
  hostname          = "gbf-proxy"
  tag               = "gbf-proxy"
  firewall_group_id = "${vultr_firewall_group.gbf-proxy.id}"
}
