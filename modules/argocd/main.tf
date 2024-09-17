provider "helm" {
  kubernetes {
    config_path = var.config_path
  }
}

resource "helm_release" "argo_cd" {
  name             = "argocd"
  repository       = "https://argoproj.github.io/argo-helm"
  chart            = "argo-cd"
  namespace        = "argocd"
  create_namespace = true
}


# 获取argocd的admin密码
resource "null_resource" "get_argocd_admin_password" {
  # provisioner "local-exec" {
  #   command = "kubectl port-forward svc/argocd-server -n argocd 8080:80"
  # }

  # provisioner "local-exec" {
  #   command = "kubectl -n argocd get secret argocd-initial-admin-secret -o jsonpath='{.data.password}' | base64 --decode"
  # }

  provisioner "remote-exec" {

    connection {
      type     = "ssh"
      user     = var.user
      password = var.password
      host     = var.public_ip
    }

    inline = [ 
      "kubectl get pods -n argocd",
      # "kubectl port-forward svc/argocd-server -n argocd 8080:80",
      # "kubectl -n argocd get secret argocd-initial-admin-secret -o jsonpath='{.data.password}' | base64 --decode",
     ]
  }
}