variable "docker_sock" {
  type = "string"
  description = "The endpoint for the Docker socket"
  default = "unix:///var/run/docker.sock"
}

provider "docker" {
  host = "${var.docker_sock}"
}

resource "docker_network" "gbf-proxy_network" {
  name = "gbf-proxy_network"
}

resource "docker_container" "controller" {
  name = "gbf-proxy_controller"
  image = "gbf-proxy:latest"
  hostname = "controller"
  command = ["controller", "8000"]

  networks_advanced {
    name = "${docker_network.gbf-proxy_network.id}"
    aliases = ["controller"]
  }

  ports {
    internal = 8000
    external = 8000
  }
}

resource "docker_container" "proxy" {
  name = "gbf-proxy_proxy"
  image = "gbf-proxy:latest"
  hostname = "proxy"
  command = ["proxy", "8088", "controller:8000"]

  networks_advanced {
    name = "${docker_network.gbf-proxy_network.id}"
    aliases = ["proxy"]
  }

  ports {
    internal = 8088
    external = 8088
  }

  depends_on = ["docker_container.controller"]
}

resource "docker_container" "nginx" {
  name = "gbf-proxy_nginx"
  image = "gbf-proxy-nginx:latest"
  hostname = "nginx"

  networks_advanced {
    name = "${docker_network.gbf-proxy_network.id}"
    aliases = ["nginx"]
  }

  ports {
    internal = 80
    external = 8080
  }

  ports {
    internal = 443
    external = 4443
  }

  depends_on = ["docker_container.proxy"]
}
