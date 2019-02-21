resource "null_resource" "master" {
  count = "${length(var.worker_hosts)}"

  connection {
    type = "ssh"
    user = "${var.master_user}"
    host = "${var.master_host}"
    private_key = "${file(var.master_pvt_key)}"
    timeout = "2m"
  }

  # Workaround due to on-going issue
  # https://github.com/hashicorp/terraform/issues/13549
  lifecycle {
    create_before_destroy = true
  }

  provisioner "file" {
    source = "../scripts/node-teardown.sh"
    destination = "/tmp/node-teardown.sh"
    when = "destroy"
  }

  provisioner "remote-exec" {
    inline = [
      "bash /tmp/node-teardown.sh ${element(var.worker_names, count.index)}"
    ]
    when = "destroy"
  }
}

resource "null_resource" "worker" {
  count = "${length(var.worker_hosts)}"

  connection {
    type = "ssh"
    user = "${element(var.worker_users, count.index)}"
    host = "${element(var.worker_hosts, count.index)}"
    private_key = "${file(var.worker_pvt_key)}"
    timeout = "2m"
  }

  lifecycle {
    create_before_destroy = true
  }

  provisioner "remote-exec" {
    script = "../scripts/setup.sh"
  }

  provisioner "remote-exec" {
    inline = [
      "sudo /opt/bin/kubeadm join ${var.kube_apiserver} --token ${var.kube_token} --discovery-token-ca-cert-hash sha256:${var.kube_hash}",
      "[ ! -e /etc/kubernetes/manifests ] && sudo mkdir -p /etc/kubernetes/manifests || true"
    ]
  }

  # Tear-down stuff go here
  provisioner "remote-exec" {
    script = "../scripts/teardown.sh"
    when = "destroy"
  }
}
