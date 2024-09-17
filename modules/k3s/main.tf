module "k3s" {
  source                   = "xunleii/k3s/module"
  k3s_version              = "v1.28.11+k3s2"
  generate_ca_certificates = true
  global_flags = [
    "--tls-san ${var.public_ip}",
    "--write-kubeconfig-mode 644",
    "--disable=traefik",
    "--kube-controller-manager-arg bind-address=0.0.0.0",
    "--kube-proxy-arg metrics-bind-address=0.0.0.0",
    "--kube-scheduler-arg bind-address=0.0.0.0"
  ]
  k3s_install_env_vars = {}

  servers = {
    "k3s" = {
      ip = var.private_ip
      connection = {
        timeout  = "60s"
        type     = "ssh"
        host     = var.public_ip
        password = var.password
        user     = var.user
      }
    }
  }
}

resource "null_resource" "fetch_kubeconfig" {
  provisioner "remote-exec" {
    connection {
      type     = "ssh"
      host     = var.public_ip
      user     = var.user
      password = var.password
    }

    inline = [
      "mkdir -p ~/.ssh",
      "echo '${file("${path.module}/sshkey/key.pub")}' >> ~/.ssh/authorized_keys",
      "chmod 700 ~/.ssh",
      "chmod 600 ~/.ssh/authorized_keys",

      "sudo cp /etc/rancher/k3s/k3s.yaml /tmp/k3s.yaml",
      "sudo chown ubuntu:ubuntu /tmp/k3s.yaml",
      "sed -i 's/127.0.0.1/${var.public_ip}/g' /tmp/k3s.yaml"
    ]
  }
  depends_on = [module.k3s]
}

resource "null_resource" "download_k3s_yaml" {
  provisioner "local-exec" {
    command = "scp -i ${path.module}/sshkey/key -o StrictHostKeyChecking=no ${var.user}@${var.public_ip}:/tmp/k3s.yaml ${path.module}/config.yaml"
  }
  depends_on = [null_resource.fetch_kubeconfig]
}