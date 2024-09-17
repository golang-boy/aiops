provider "helm" {
  kubernetes {
    config_path = var.config_path
  }
}

resource "helm_release" "crossplane" {
  # depends_on       = var.depends_on_
  name             = "crossplane"
  repository       = "https://charts.crossplane.io/stable"
  chart            = "crossplane"
  namespace        = "crossplane"
  create_namespace = true
}


# 在远程执行kubectl apply -f 命令

resource "null_resource" "apply_crossplane_provider" {

  depends_on = [ helm_release.crossplane ]

  connection {
    type     = "ssh"
    user     = var.user
    password = var.password
    host     = var.public_ip
  }

  provisioner "file" {
    source      = "${path.module}/provider"
    destination = "/home/${var.user}"
  }

  provisioner "remote-exec" {
    inline = [   
      "kubectl apply -f ~/provider/",
    ]
  }
}
