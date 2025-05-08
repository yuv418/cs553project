variable "instance_type" {
  description = "EC2 instance type"
  type        = string
}

variable "ami_id" {
  description = "AMI ID for EC2 instances"
  type        = string
}

variable "subnet_id" {
  description = "Subnet ID for instance"
  type        = string
}

variable "security_groups" {
  description = "Security group IDs"
  type        = list(string)
}

variable "key_name" {
  description = "SSH key name"
  type        = string
  default     = ""
}

variable "user_data" {
  description = "User data script for instance initialization"
  type        = string
}

variable "instance_name" {
  description = "Name tag for the instance"
  type        = string
}

variable "environment" {
  description = "Deployment environment"
  type        = string
}

variable "service_name" {
  description = "Name of the service running on this instance"
  type        = string
}

variable "create_key_pair" {
  description = "Whether to create a new key pair in AWS"
  type        = bool
  default     = false
}

variable "ssh_public_key" {
  description = "SSH public key content when creating a new key pair"
  type        = string
  default     = ""
}

variable "certificate_content" {
  description = "Base64 encoded certificate content"
  type        = string
  default     = ""
}

variable "private_key_content" {
  description = "Base64 encoded private key content"
  type        = string
  default     = ""
}

variable "spki_hash" {
  description = "SPKI hash for WebTransport"
  type        = string
  default     = ""
}

# Only create key pair when explicitly requested
resource "aws_key_pair" "instance_key" {
  count      = var.create_key_pair ? 1 : 0
  key_name   = var.key_name
  public_key = var.ssh_public_key

  lifecycle {
    create_before_destroy = true
  }
}

resource "aws_instance" "service_instance" {
  ami                    = var.ami_id
  instance_type          = var.instance_type
  subnet_id              = var.subnet_id
  vpc_security_group_ids = var.security_groups
  key_name               = var.key_name
  user_data              = var.user_data

  lifecycle {
    ignore_changes = [user_data]
  }
  
  root_block_device {
    volume_size = 20
    volume_type = "gp2"
  }
  
  tags = {
    Name        = var.instance_name
    Environment = var.environment
    Service     = var.service_name
  }
}

output "instance_id" {
  value = aws_instance.service_instance.id
}

output "instance_private_ip" {
  value = aws_instance.service_instance.private_ip
}

output "instance_public_ip" {
  value = aws_instance.service_instance.public_ip
}

output "instance_public_dns" {
  value = aws_instance.service_instance.public_dns
}
