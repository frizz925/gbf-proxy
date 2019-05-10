output "public_ip" {
  value = "${vultr_instance.gbf-proxy.ipv4_address}"
}
