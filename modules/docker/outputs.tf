
output "versions" {
  value = [ for i in range(length(var.targets)):
   file("${path.module}/docker_version_${i}.txt")
   ]
}