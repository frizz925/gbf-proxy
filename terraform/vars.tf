variable "vultr_api_key" {
  type = "string"
}

variable "bastion_cidr" {
  type = "string"
  default = "0.0.0.0/0"
}
