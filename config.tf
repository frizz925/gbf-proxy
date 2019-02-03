provider "aws" {
  region = "ap-southeast-1"
  shared_credentials_file = "${pathexpand("~/.aws/credentials")}"
  profile = "gbf-proxy"
}

data "aws_ami" "coreos" {
  most_recent = true

  filter {
    name = "name"
    values = ["CoreOS-stable-1967.4.0-hvm-*"]
  }

  filter {
    name = "virtualization-type"
    values = ["hvm"]
  }

  filter {
    name = "product-code.type"
    values = ["marketplace"]
  }

  filter {
    name = "product-code"
    values = ["ryg425ue2hwnsok9ccfastg4"]
  }

  owners = ["679593333241"] # Canonical
}

resource "aws_instance" "proxy" {
  ami = "${data.aws_ami.coreos.id}"
  instance_type = "t2.micro"

  tags = {
      Name = "Granblue Proxy"
  }
}

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

  ports {
    internal = 80
    external = 8080
  }

  ports {
    internal = 443
    external = 4443
  }

  volumes {
    host_path = "${path.cwd}/nginx/conf.d"
    container_path = "/etc/nginx/conf.d"
    read_only = true
  }
}
