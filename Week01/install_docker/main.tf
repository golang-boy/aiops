module "创建腾讯云实例" {
    source = "../../modules/tencentCvm"
    count_num = 2
}
module "每个实例上安装docker" {
    source = "../../modules/docker"
    targets = module.创建腾讯云实例.public_ips
    depends_on_ =  module.创建腾讯云实例 
}