# Variable Definitions
variable "targets" {
  description = "示例列表"
  type        = list(string)
}

variable "user" {
  default = "ubuntu"
}

variable "password" {
  default = "password123"
}

variable "depends_on_" {
  default = null
}