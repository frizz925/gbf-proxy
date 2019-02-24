variable "user" {
  type = "string"
  default = "core"
}

variable "host" {
  type = "string"
}

variable "pvt_key" {
  type = "string"
  default = "~/.ssh/id_rsa"
}

variable "domains" {
  type = "list"
}

variable "email" {
  type = "string"
}

variable "production" {
  default = true
}
