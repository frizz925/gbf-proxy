resource "vultr_ssh_key" "main" {
  name    = "Granblue Proxy SSH Key"
  ssh_key = "${trimspace(file("certs/id_rsa.pub"))}"
}
