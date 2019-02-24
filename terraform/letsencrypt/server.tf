resource "null_resource" "server" {
  connection {
    type = "ssh"
    user = "${var.user}"
    host = "${var.host}"
    private_key = "${file(var.pvt_key)}"
    timeout = "2m"
  }

  provisioner "file" {
    source = "scripts/setup.sh"
    destination = "/tmp/letsencrypt-setup.sh"
  }

  provisioner "remote-exec" {
    inline = [
      "export LETSENCRYPT_EMAIL=${var.email}",
      "export LETSENCRYPT_DOMAINS='${join(" ", var.domains)}'",
      "sudo -E bash /tmp/letsencrypt-setup.sh < /dev/null"
    ]
  }
}
