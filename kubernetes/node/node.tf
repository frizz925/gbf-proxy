resource "null_resource" "provisioner" {
  provisioner "local-exec" {
    command = "/bin/bash ../scripts/provision.sh ${pathexpand(var.project_dir)}"
  }
}

resource "null_resource" "node" {
  count = "${length(var.node_hosts)}"

  connection {
    type = "ssh"
    user = "${element(var.node_users, count.index)}"
    host = "${element(var.node_hosts, count.index)}"
    private_key = "${file(var.node_pvt_key)}"
    timeout = "2m"
  }

  provisioner "file" {
    source = "../files/gbf-proxy.tar.gz"
    destination = "/tmp/gbf-proxy.tar.gz"
  }

  provisioner "file" {
    source = "../files/gbf-proxy-web.tar.gz"
    destination = "/tmp/gbf-proxy-web.tar.gz"
  }

  provisioner "file" {
    source = "../files/gbf-proxy-version"
    destination = "/tmp/gbf-proxy-version"
  }

  provisioner "remote-exec" {
    script = "../scripts/docker-setup.sh"
  }

  provisioner "remote-exec" {
    script = "../scripts/docker-teardown.sh"
    when = "destroy"
  }

  depends_on = ["null_resource.provisioner"]
}
