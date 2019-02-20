# resource "null_resource" "provisioner" {
#   provisioner "local-exec" {
#     command = "/bin/bash ../scripts/provision.sh ${pathexpand(var.project_dir)}"
#   }
# }

resource "null_resource" "worker" {
  count = "${length(var.worker_hosts)}"

  connection {
    type = "ssh"
    user = "core"
    host = "${element(var.worker_hosts, count.index)}"
    private_key = "${file(var.worker_pvt_key)}"
    timeout = "2m"
  }

  # Workaround due to on-going issue
  # https://github.com/hashicorp/terraform/issues/13549
  lifecycle {
    create_before_destroy = true
  }

  provisioner "local-exec" {
    command = "/bin/bash ../scripts/node-teardown.sh ${element(var.worker_names, count.index)}"
    when = "destroy"
  }

  provisioner "remote-exec" {
    script = "../scripts/teardown.sh"
    when = "destroy"
  }

  provisioner "remote-exec" {
    script = "../scripts/setup.sh"
  }

  provisioner "remote-exec" {
    inline = [
      "sudo /opt/bin/kubeadm join ${var.kube_apiserver} --token ${var.kube_token} --discovery-token-ca-cert-hash sha256:${var.kube_hash}",
      # "sudo systemctl restart kubelet.service",
      "[ ! -e /etc/kubernetes/manifests ] && sudo mkdir -p /etc/kubernetes/manifests || true"
    ]
  }

  # provisioner "file" {
  #   source = "../files/gbf-proxy.tar.gz"
  #   destination = "/tmp/gbf-proxy.tar.gz"
  # }

  # provisioner "remote-exec" {
  #   script = "../scripts/docker-setup.sh"
  # }
}
