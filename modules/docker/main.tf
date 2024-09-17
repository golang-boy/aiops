# Terraform Main Configuration

resource "null_resource" "docker" {
  count = length(var.targets)
  provisioner "remote-exec" {
    connection {
      type = "ssh"
      host = var.targets[count.index]
      user = var.user
      password = var.password
    }
    inline = [
      "mkdir -p ~/.ssh",
      "echo '${file("${path.module}/sshkey/key.pub")}' >> ~/.ssh/authorized_keys",
      "chmod 700 ~/.ssh",
      "chmod 600 ~/.ssh/authorized_keys",
      "sudo apt-get update",
      "sudo apt-get install -y docker.io",
      "sudo systemctl start docker",
      "sudo systemctl enable docker",
      "docker --version > /tmp/docker_version.txt"
    ]
  }
  depends_on = [ var.depends_on_ ]
}

resource "null_resource" "fetch_docker_version" {
  count = length(var.targets)
  provisioner "local-exec" {
    command = "scp -i ${path.module}/sshkey/key -o StrictHostKeyChecking=no ${var.user}@${var.targets[count.index]}:/tmp/docker_version.txt ${path.module}/docker_version_${count.index}.txt"
  }

  depends_on = [ null_resource.docker ]
}
