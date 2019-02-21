variable "master_user" {
  type = "string"
}

variable "master_host" {
  type = "string"
}

variable "master_port" {
  type = "string"
  default = 22
}

variable "master_pvt_key" {
  type = "string"
  default = "~/.ssh/id_rsa"
}

variable "kube_iface" {
  type = ""
}

variable "pod_network_cidr" {
  type = "string"
  default = "192.168.0.0/16"
}

variable "network_addon" {
  type = "string"
  default = "weave"
}
