variable "vpc_id" {
  description = "VPC ID for security groups"
  type        = string
}

variable "project_name" {
  description = "Name of the project"
  type        = string
}

variable "environment" {
  description = "Deployment environment"
  type        = string
}

resource "aws_security_group" "flappygo_sg" {
  name        = "${var.project_name}-sg"
  description = "Security group for FlappyGo application"
  vpc_id      = var.vpc_id
  
  # Allow HTTP
  ingress {
    from_port   = 80
    to_port     = 80
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
    description = "HTTP access"
  }
  
  # Allow HTTPS
  ingress {
    from_port   = 443
    to_port     = 443
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
    description = "HTTPS access"
  }
  
  # Allow SSH
  ingress {
    from_port   = 22
    to_port     = 22
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
    description = "SSH access"
  }
  
  # Allow gRPC ports (50051-50060)
  ingress {
    from_port   = 50051
    to_port     = 50060
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
    description = "gRPC service ports"
  }
  
  # Allow WebTransport port (TCP)
  ingress {
    from_port   = 4433
    to_port     = 4434
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
    description = "WebTransport TCP ports"
  }
  
  # Allow WebTransport UDP (QUIC)
  ingress {
    from_port   = 4433
    to_port     = 4434
    protocol    = "udp"
    cidr_blocks = ["0.0.0.0/0"]
    description = "WebTransport UDP ports"
  }
  
  # Allow web client port
  ingress {
    from_port   = 8080
    to_port     = 8080
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
    description = "Web client port"
  }
  
  # Allow all outbound traffic
  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
    description = "Allow all outbound traffic"
  }
  
  tags = {
    Name        = "${var.project_name}-sg"
    Environment = var.environment
  }
}

output "sg_id" {
  value = aws_security_group.flappygo_sg.id
}
