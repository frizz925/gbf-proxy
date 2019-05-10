variable "bastion_cidr" {
  type = "string"
}

variable "region_id" {
  type = "string"
  default = "40" # Singapore
}

variable "plan_id" {
  type = "string"
  default = "201" # 1vCPU 1GB RAM $5.00
}

variable "os_id" {
  type = "string"
  default = "244" # Debian 9 x64 (Stretch)
}
