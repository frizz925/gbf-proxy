output "public_ip" {
  value = "${cidrhost(vultr_reserved_ip.gbf-proxy.cidr, 0)}"
}
