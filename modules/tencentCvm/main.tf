# Terraform Main Configuration
provider "tencentcloud" {
  region = var.region
  secret_id = var.secret_id
  secret_key = var.secret_key
}

data "tencentcloud_availability_zones_by_product" "default" {
  product = "cvm"
}

data "tencentcloud_images" "default" {
   image_type = var.image.image_type
   os_name = var.image.os_name
}

data "tencentcloud_instance_types" "default" {
  filter {
    name = "instance-family"
    values = var.instance_family
  }

  cpu_core_count = var.cpu_core_count
  memory_size = var.memory_size
  exclude_sold_out = true
}

resource "tencentcloud_security_group" "default" {
  name        = "tf-security-group"
  description = "make it accessible for both production and stage ports"
}


resource "tencentcloud_security_group_lite_rule" "default" {
  security_group_id = tencentcloud_security_group.default.id
  ingress = var.ingress
  egress = var.egress
}


resource "tencentcloud_instance" "tins" {
  depends_on = [ tencentcloud_security_group_lite_rule.default ]
  count = var.count_num
  instance_name = var.instance_name
  availability_zone = data.tencentcloud_availability_zones_by_product.default.zones[0].name
  image_id = data.tencentcloud_images.default.images[0].image_id
  instance_type = data.tencentcloud_instance_types.default.instance_types[0].instance_type
  password = var.password

  system_disk_type = var.disk.dtype
  system_disk_size = var.disk.size

  allocate_public_ip = var.allocate_public_ip
  internet_max_bandwidth_out = var.internet_max_bandwidth_out
  instance_charge_type = var.instance_charge_type

  orderly_security_groups = [ tencentcloud_security_group.default.id ]

}