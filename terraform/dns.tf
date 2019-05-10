resource "vultr_dns_domain" "gbf-proxy" {
  domain  = "gbf-proxy.moe"
  ip      = "${cidrhost(vultr_reserved_ip.gbf-proxy.cidr, 0)}"
}

resource "vultr_dns_record" "gbf-proxy" {
  domain  = "${vultr_dns_domain.gbf-proxy.id}"
  name    = "main"
  type    = "A"
  data    = "${vultr_dns_domain.gbf-proxy.ip}"
  ttl     = 300
}
