# Variable Definitions

variable "instance_name" {
  default = "my_instance"
}

variable "count_num" {
  default = 1
}

variable "secret_id" {
  default = "your_secret_id"
}

variable "secret_key" {
  default = "your_secret_key"
}


variable "password" {
  default = "password123"
}

variable "cpu_core_count" {
  default = 2
}

variable "memory_size" {
  description = "内存单位为GB"
  default = 4
}

variable "disk" {
  type = object({
    dtype = string 
    size = number 
  })
  default = {
    dtype = "CLOUD_BSSD"
    size = 50
  }
}

variable "allocate_public_ip" {
  default = true
}

variable "internet_max_bandwidth_out" {
  default = 100
}


variable "instance_charge_type" {
  default = "SPOTPAID"
}

variable "region" {
  type = string
  default = "ap-hongkong"
}

variable "instance_family" {
  type = list(string)
  default = ["SA5"]
}

variable "image" {
  type = object({
    image_type = list(string)
    os_name = string
  })
  default = {
    image_type = [ "PUBLIC_IMAGE" ]
    os_name = "ubuntu"
  }
}

variable "ingress" {
  type =list(string)
  default = [ 
    "ACCEPT#0.0.0.0/0#22#TCP",
    "ACCEPT#0.0.0.0/0#6443#TCP",
 ]
}

variable "egress" {
  type =list(string)
  default = [ 
    "ACCEPT#0.0.0.0/0#ALL#ALL"
 ]
}
