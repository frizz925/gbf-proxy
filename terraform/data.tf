data "vultr_region" "sgp" {
  filter {
    name = "regioncode"
    values = ["SGP"]
  }
}

data "vultr_plan" "gp_1vcpu_1gb" {
  filter {
    name = "vcpu_count"
    values = ["1"]
  }
  filter {
    name = "ram"
    values = ["1024"]
  }
  filter {
    name = "disk"
    values = ["25"]
  }
}

data "vultr_os" "debian_10" {
  filter {
    name = "name"
    values = ["Debian 10 x64 (buster)"]
  }
}
