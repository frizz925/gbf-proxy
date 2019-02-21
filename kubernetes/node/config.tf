variable "node_users" {
  type = "list"
}

variable "node_hosts" {
  type = "list"
}

variable "node_names" {
  type = "list"
}

variable "node_pvt_key" {
  type = "string"
}

variable "project_dir" {
  type = "string"
  default = "../.."
}
