variable "aws_region" {
  description = "AWS region to deploy resources"
  type        = string
  default     = "us-east-1"
}

variable "aws_regions" {
  description = "List of AWS regions for global deployment"
  type        = list(string)
  default     = ["us-east-1", "us-west-2", "eu-west-1", "ap-southeast-1", "sa-east-1", "ap-northeast-1"]
}

variable "availability_zones" {
  description = "List of availability zones to use within the region"
  type        = list(string)
  default     = ["a", "b", "c"]
}

variable "vpc_cidr" {
  description = "CIDR block for the VPC"
  type        = string
  default     = "10.0.0.0/16"
}

variable "deployment_mode" {
  description = "Deployment mode - monolith or microservices"
  type        = string
  default     = "monolith"
  validation {
    condition     = contains(["monolith", "microservices"], var.deployment_mode)
    error_message = "Deployment mode must be either 'monolith' or 'microservices'."
  }
}

variable "deployment_pattern" {
  description = "Deployment pattern - single_instance, multi_az, multi_region"
  type        = string
  default     = "single_instance"
  validation {
    condition     = contains(["single_instance", "multi_az", "multi_region"], var.deployment_pattern)
    error_message = "Deployment pattern must be one of: 'single_instance', 'multi_az', or 'multi_region'."
  }
}

variable "project_name" {
  description = "Name of the project"
  type        = string
  default     = "flappygo"
}

variable "environment" {
  description = "Deployment environment"
  type        = string
  default     = "dev"
}

variable "instance_type" {
  description = "EC2 instance type"
  type        = string
  default     = "t2.micro"
}

variable "ami_id" {
  description = "AMI ID for EC2 instances"
  type        = string
  default     = "ami-0f88e80871fd81e91"
}

variable "key_name" {
  description = "SSH key name for EC2 instances"
  type        = string
  default     = ""
}

variable "service_ports" {
  description = "Port mapping for each service"
  type        = map(number)
  default     = {
    auth      = 50051
    worldgen  = 50052
    engine    = 50053
    initiator = 50054
    music     = 50055
    score     = 50056
    monolith  = 50051
    client    = 8080
  }
}

variable "use_own_certificates" {
  description = "Whether to use provided certificates or generate self-signed ones"
  type        = bool
  default     = false
}

variable "certificate_path" {
  description = "Path to certificate file (PEM format) when using own certificates"
  type        = string
  default     = ""
}

variable "private_key_path" {
  description = "Path to private key file (PEM format) when using own certificates"
  type        = string
  default     = ""
}

variable "certificate_validity_days" {
  description = "Validity period in days for auto-generated certificates"
  type        = number
  default     = 365
}

variable "github_token" {
  description = "GitHub personal access token for private repository access"
  type        = string
  default     = ""  # Optional, allows public repo access when not set
  sensitive   = true
}

variable "ssh_private_key_path" {
  description = "Path to SSH private key for instance access and configuration (also used for TLS if no certificates provided)"
  type        = string
  default     = ""
}
