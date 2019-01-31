variable "host" {
    type = "string"
    default = "17"
}

variable "ssh_key_private" {
    type = "string"
    default = "ssh/id_rsa"
}

variable "nginx_path" {
    type = "string"
    default = "/home/core/nginx"
}

provider "docker" {
    host = "tcp://${var.host}:2375/"
}

resource "null_resource" "provision" {
    connection {
        type = "ssh"
        host = "${var.host}"
        user = "core"
        private_key = "${file("~/.vagrant.d/insecure_private_key")}"
    }

    provisioner "remote-exec" {
        inline = [
            "[ -d ${var.nginx_path} ] && rm -rf ${var.nginx_path}",
            "mkdir -p ${var.nginx_path}/{conf.d,certs}"
        ]
    }

    provisioner "file" {
        source = "conf.d"
        destination = "${var.nginx_path}"
    }

    provisioner "file" {
        source = "certs/gbf-proxy.bundle.crt"
        destination = "${var.nginx_path}/certs/gbf-proxy.bundle.crt"
    }

    provisioner "file" {
        source = "certs/gbf-proxy.key"
        destination = "${var.nginx_path}/certs/gbf-proxy.key"
    }

    provisioner "file" {
        source = "certs/dhparam.pem"
        destination = "${var.nginx_path}/certs/dhparam.pem"
    }

    provisioner "remote-exec" {
        inline = [
            "find ${var.nginx_path} -type f -exec chmod 644 {} \\+",
            "find ${var.nginx_path} -type d -exec chmod 755 {} \\+",
            "find ${var.nginx_path}/certs -iname '*.key' -exec chmod 600 {} \\+"
        ]
    }
}

resource "docker_image" "nginx" {
    name = "nginx:1.15"
}

resource "docker_container" "nginx-server" {
    name = "nginx-server"
    image = "${docker_image.nginx.latest}"
    ports = {
        internal = 80
        external = 80
    }

    ports = {
        internal = 443
        external = 443
    }

    volumes = {
        container_path = "/etc/nginx/conf.d"
        host_path = "${var.nginx_path}/conf.d"
        read_only = true
    }

    volumes = {
        container_path = "/etc/nginx/certs"
        host_path = "${var.nginx_path}/certs"
        read_only = true
    }
}