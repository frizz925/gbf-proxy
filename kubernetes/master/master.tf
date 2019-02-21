resource "null_resource" "master" {
  connection {
    type = "ssh"
    user = "${var.master_user}"
    host = "${var.master_host}"
    port = "${var.master_port}"
    private_key = "${file(var.master_pvt_key)}"
  }

  lifecycle {
    create_before_destroy = true
  }

  provisioner "file" {
    source = "../scripts/setup.sh"
    destination = "/tmp/setup.sh"
  }

  provisioner "file" {
    source = "../scripts/master-setup.sh"
    destination = "/tmp/master-setup.sh"
  }

  provisioner "file" {
    source = "../scripts/network-setup.sh"
    destination = "/tmp/network-setup.sh"
  }

  provisioner "remote-exec" {
    inline = [
      "export LOCAL_IFACE=${var.kube_iface}",
      "export KUBEADM_EXTRA_ARGS=--pod-network-cidr=${var.pod_network_cidr}",
      "export K8S_NETWORKING_ADDON=${var.network_addon}",
      "bash /tmp/setup.sh",
      "bash /tmp/master-setup.sh",
      "[ ! -d ~/.kube ] && mkdir ~/.kube",
      "sudo cp /etc/kubernetes/admin.conf ~/.kube/config",
      "sudo chown $(id -u):$(id -g) ~/.kube/config",
      "sudo sysctl net.bridge.bridge-nf-call-iptables=1",
      "bash /tmp/network-setup.sh"
    ]
  }

  provisioner "file" {
    source = "../scripts/node-teardown.sh"
    destination = "/tmp/node-teardown.sh"
    when = "destroy"
  }

  provisioner "remote-exec" {
    scripts = [
      "../scripts/network-teardown.sh",
      "../scripts/master-teardown.sh",
      "../scripts/teardown.sh"
    ]
    when = "destroy"
  }
}
