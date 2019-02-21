variable "do_token" {
  type = "string"
}

variable "do_region" {
  type = "string"
  default = "sgp1"
}

variable "ssh_fingerprint" {
  type = "string"
}

variable "ssh_private_key" {
  type = "string"
  default = "~/.ssh/id_rsa"
}

provider "digitalocean" {
  token = "${var.do_token}"
}
