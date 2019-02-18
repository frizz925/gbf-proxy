variable "worker_hosts" {
  type = "list"
}

variable "worker_names" {
  type = "list"
}

variable "worker_pvt_key" {
  type = "string"
}

variable "kube_apiserver" {
  type = "string"
}

variable "kube_token" {
  type = "string"
}

variable "kube_hash" {
  type = "string"
}

variable "project_dir" {
  type = "string"
  default = "../.."
}
