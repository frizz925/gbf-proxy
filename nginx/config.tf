provider "docker" {
    host = "unix:///var/run/docker.sock"
}

resource "docker_image" "nginx" {
    name = "nginx:1.15"
    keep_locally = true
}

resource "docker_container" "nginx-server" {
    name = "nginx-server"
    image = "${docker_image.nginx.latest}"

    ports = {
        internal = 80
        external = 8080
    }

    ports = {
        internal = 43
        external = 4443
    }

    volumes = {
        container_path = "/etc/nginx/conf.d"
        host_path = "${path.cwd}/conf.d"
        read_only = true
    }

    volumes = {
        container_path = "/etc/nginx/certs"
        host_path = "${path.cwd}/certs"
        read_only = true
    }
}