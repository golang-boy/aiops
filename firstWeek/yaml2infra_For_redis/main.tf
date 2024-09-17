# 1.申请cvm
module "创建腾讯云实例" {
    source = "../../modules/tencentCvm"
    count_num = 1
}

# 2.安装k3s
module "安装k3s" {
    source = "../../modules/k3s"

    public_ip = module.创建腾讯云实例.public_ips[0]
    private_ip = module.创建腾讯云实例.private_ips[0]

    # depends_on = [ module.创建腾讯云实例 ]
}

# 3.安装crossplane
module "安装crossplane" {
    source = "../../modules/crossplane"

    config_path = module.安装k3s.kube_config
    public_ip = module.安装k3s.public_ip
    # depends_on = [ module.安装k3s ]
}

# 4.安装argo-cd
module "安装argo" {
  source = "../../modules/argocd"
  config_path = module.安装k3s.kube_config
  public_ip = module.安装k3s.public_ip

#   depends_on = [ module.安装k3s ]
}

# 5.yaml2infra方式申请云redis