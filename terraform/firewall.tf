resource "vultr_firewall_group" "gbf-proxy" {
  description = "Firewall group for Granblue Proxy"
}

resource "vultr_firewall_rule" "ssh" {
  firewall_group_id = "${vultr_firewall_group.gbf-proxy.id}"
  cidr_block        = "${var.bastion_cidr}"
  protocol          = "tcp"
  from_port         = 22
  to_port           = 22
}

resource "vultr_firewall_rule" "http" {
  firewall_group_id = "${vultr_firewall_group.gbf-proxy.id}"
  cidr_block        = "0.0.0.0/0"
  protocol          = "tcp"
  from_port         = 80
  to_port           = 80
}

resource "vultr_firewall_rule" "https" {
  firewall_group_id = "${vultr_firewall_group.gbf-proxy.id}"
  cidr_block        = "0.0.0.0/0"
  protocol          = "tcp"
  from_port         = 443
  to_port           = 443
}

resource "vultr_firewall_rule" "proxy" {
  firewall_group_id = "${vultr_firewall_group.gbf-proxy.id}"
  cidr_block        = "0.0.0.0/0"
  protocol          = "tcp"
  from_port         = 8088
  to_port           = 8088
}
