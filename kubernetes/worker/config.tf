variable "worker_hosts" {
  type = "list"
  default = [
    "172.17.8.101",
    "172.17.8.102",
    "172.17.8.103"
  ]
}

variable "worker_pvt_key" {
  type = "string"
  default = "~/.vagrant.d/insecure_private_key"
}

variable "kube_apiserver" {
  type = "string"
  default = "172.17.8.1:6443"
}

variable "kube_token" {
  type = "string"
  default = "h4qwku.jjralldm8fnuc5qw"
}

variable "kube_hash" {
  type = "string"
  default = "76e3d44c248734240b53baf5a1842b9e04abe9b9f2f8c4bcec4373fa605371a0"
}

variable "project_dir" {
  type = "string"
  default = "../.."
}
