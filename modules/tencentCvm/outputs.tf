# Output Definitions
output "public_ips" {
  value = tencentcloud_instance.tins[*].public_ip
}

output "private_ips" {
  value = tencentcloud_instance.tins[*].private_ip
}
