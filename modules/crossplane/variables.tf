# Variable Definitions


variable "config_path" {
  description = "k3s config file path"
  type        = string
}

variable "public_ip" {
  type = string
}

variable "user" {
  default = "ubuntu"
}

variable "password" {
  default = "password123"
}

