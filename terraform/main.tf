terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
    null = {
      source = "hashicorp/null"
      version = "~> 3.0"
    }
  }
  required_version = ">= 1.2.0"
}

# Default provider configuration
provider "aws" {
  region = var.aws_region
  shared_credentials_files = [ "../.credentials" ]
}

# Certificate generation for when certificates are not provided
resource "null_resource" "generate_certificates" {
  count = var.use_own_certificates ? 0 : 1

  provisioner "local-exec" {
    command = "${path.module}/../generate_certs.sh --output-dir ${path.module}/../certs"
  }
}

locals {
  # Determine which services to deploy as individual instances
  services = var.deployment_mode == "monolith" ? [] : ["auth", "worldgen", "engine", "initiator", "music", "score"]
  
  # Create service to location mapping based on deployment pattern
  service_azs = var.deployment_pattern == "multi_az" || var.deployment_pattern == "multi_region" ? {
    auth      = var.availability_zones[0]
    worldgen  = var.availability_zones[1 % length(var.availability_zones)]
    engine    = var.availability_zones[2 % length(var.availability_zones)]
    initiator = var.availability_zones[3 % length(var.availability_zones)]
    music     = var.availability_zones[4 % length(var.availability_zones)]
    score     = var.availability_zones[5 % length(var.availability_zones)]
  } : { for s in local.services : s => var.availability_zones[0] }
  
  # Create service to region mapping for global deployment
  service_regions = { for s in local.services : s => var.aws_region }
  
  # Map regions to provider aliases
  region_to_alias = {
    "us-east-1"      = "us-east-1"
    "us-west-2"      = "us-west-2"
    "eu-west-1"      = "eu-west-1"
    "ap-southeast-1" = "ap-southeast-1"
    "sa-east-1"      = "sa-east-1"
    "ap-northeast-1" = "ap-northeast-1"
  }
  
  # Certificate handling
  certificate_generated = var.use_own_certificates ? false : true
  cert_exists = var.use_own_certificates ? var.certificate_path != "" : length(null_resource.generate_certificates) > 0
  
  cert_path = var.use_own_certificates ? (
    var.certificate_path != "" ? var.certificate_path : ""
  ) : (
    length(null_resource.generate_certificates) > 0 ? "${path.module}/../certs/cert.pem" : ""
  )
  
  key_path = var.use_own_certificates ? (
    var.private_key_path != "" ? var.private_key_path : ""
  ) : (
    length(null_resource.generate_certificates) > 0 ? "${path.module}/../certs/key.pem" : ""
  )
  
  # Make sure certificates exist before trying to read them
  certificate_content = local.cert_path != "" ? filebase64(local.cert_path) : ""
  private_key_content = local.key_path != "" ? filebase64(local.key_path) : ""
  
  # Only calculate hash when using generated certs
  spki_hash = !var.use_own_certificates && length(null_resource.generate_certificates) > 0 ? (
    file("${path.module}/../certs/spki_hash.txt")
  ) : ""
  
  ssh_public_key = var.ssh_private_key_path != "" ? file("${var.ssh_private_key_path}.pub") : ""
}

# Create AWS key pair if needed
resource "aws_key_pair" "deployment_key" {
  count      = var.key_name != "" ? 1 : 0
  key_name   = var.key_name
  public_key = local.ssh_public_key

  lifecycle {
    # Ignore changes to public key after creation
    ignore_changes = [public_key]
  }
}

# VPC for our infrastructure
module "vpc" {
  source = "./modules/vpc"
  count= 1
  # count  = var.deployment_pattern == "global" ? length(var.aws_regions) : 1
  
  region            = var.deployment_pattern == "global" ? var.aws_regions[count.index] : var.aws_region
  availability_zones = var.availability_zones
  cidr_block        = var.vpc_cidr
  project_name      = var.project_name
  environment       = var.environment
}

# Security group for all instances
module "security_groups" {
  source = "./modules/security"
  # count  = var.deployment_pattern == "global" ? length(var.aws_regions) : 1
  count = 1
  
  vpc_id       = module.vpc[count.index].vpc_id
  project_name = var.project_name
  environment  = var.environment
}

# Monolith deployment
module "monolith" {
  source = "./modules/compute"
  count  = var.deployment_mode == "monolith" ? 1 : 0
  
  instance_type   = var.instance_type
  ami_id          = var.ami_id
  subnet_id       = module.vpc[0].public_subnets[0]
  security_groups = [module.security_groups[0].sg_id]
  key_name        = var.key_name != "" ? aws_key_pair.deployment_key[0].key_name : null
  create_key_pair = false  # Key pair is managed at root level
  ssh_public_key  = local.ssh_public_key
  
  user_data = templatefile("${path.module}/templates/monolith_userdata.tpl", {
    aws_region     = var.aws_region
    service_ports  = var.service_ports
    github_token   = var.github_token
    use_own_certificates = var.use_own_certificates
    certificate_content = local.certificate_content
    private_key_content = local.private_key_content
    certificate_validity_days = var.certificate_validity_days
    spki_hash = local.spki_hash
  })
  
  instance_name = "${var.project_name}-monolith"
  environment   = var.environment
  service_name  = "monolith"
}

# Microservices deployment
module "microservices" {
  source = "./modules/compute"
  count  = var.deployment_mode == "microservices" ? length(local.services) : 0
  
  instance_type   = var.instance_type
  ami_id          = var.ami_id
  # subnet_id       = module.vpc[var.deployment_pattern == "global" ? index(var.aws_regions, local.service_regions[local.services[count.index]]) : 0].public_subnets[index(var.availability_zones, local.service_azs[local.services[count.index]])]
  # security_groups = [module.security_groups[var.deployment_pattern == "global" ? index(var.aws_regions, local.service_regions[local.services[count.index]]) : 0].sg_id]
  subnet_id       = module.vpc[0].public_subnets[index(var.availability_zones, local.service_azs[local.services[count.index]])]
  security_groups = [module.security_groups[0].sg_id]
  key_name        = var.key_name != "" ? aws_key_pair.deployment_key[0].key_name : null
  create_key_pair = false  # Key pair is managed at root level
  ssh_public_key  = local.ssh_public_key
  
  user_data = templatefile("${path.module}/templates/microservice_userdata.tpl", {
    aws_region     = local.service_regions[local.services[count.index]]
    service_name   = local.services[count.index]
    service_port   = var.service_ports[local.services[count.index]]
    service_ports  = var.service_ports
    auth_url      = ""  # Will be configured after instance creation
    initiator_url = ""
    score_url     = ""
    engine_url    = ""
    worldgen_url  = ""
    music_url     = ""
    github_token   = var.github_token
    use_own_certificates = var.use_own_certificates
    certificate_content = local.certificate_content
    private_key_content = local.private_key_content
    certificate_validity_days = var.certificate_validity_days
    spki_hash = local.spki_hash
  })
  
  instance_name = "${var.project_name}-${local.services[count.index]}"
  environment   = var.environment
  service_name  = local.services[count.index]
}

# Service configuration updater
resource "null_resource" "service_config_updater" {
  count = var.deployment_mode == "microservices" ? length(local.services) : 0

  triggers = {
    instance_ids = module.microservices[count.index].instance_id
  }

  provisioner "remote-exec" {
    inline = [
      # Wait for service file to exist
      "timeout 300 bash -c 'until [ -f /etc/systemd/system/flappygo-${local.services[count.index]}.service ]; do sleep 5; done'",
      "sudo sed -i 's|^Environment=\"AUTH_URL=.*\"|Environment=\"AUTH_URL=${module.microservices[index(local.services, "auth")].instance_public_dns}:${var.service_ports["auth"]}\"|' /etc/systemd/system/flappygo-${local.services[count.index]}.service",
      "sudo sed -i 's|^Environment=\"INITIATOR_URL=.*\"|Environment=\"INITIATOR_URL=${module.microservices[index(local.services, "initiator")].instance_public_dns}:${var.service_ports["initiator"]}\"|' /etc/systemd/system/flappygo-${local.services[count.index]}.service",
      "sudo sed -i 's|^Environment=\"SCORE_URL=.*\"|Environment=\"SCORE_URL=${module.microservices[index(local.services, "score")].instance_public_dns}:${var.service_ports["score"]}\"|' /etc/systemd/system/flappygo-${local.services[count.index]}.service",
      "sudo sed -i 's|^Environment=\"GAME_ENGINE_URL=.*\"|Environment=\"GAME_ENGINE_URL=${module.microservices[index(local.services, "engine")].instance_public_dns}:${var.service_ports["engine"]}\"|' /etc/systemd/system/flappygo-${local.services[count.index]}.service",
      "sudo sed -i 's|^Environment=\"WORLD_GEN_URL=.*\"|Environment=\"WORLD_GEN_URL=${module.microservices[index(local.services, "worldgen")].instance_public_dns}:${var.service_ports["worldgen"]}\"|' /etc/systemd/system/flappygo-${local.services[count.index]}.service",
      "sudo sed -i 's|^Environment=\"MUSIC_URL=.*\"|Environment=\"MUSIC_URL=${module.microservices[index(local.services, "music")].instance_public_dns}:${var.service_ports["music"]}\"|' /etc/systemd/system/flappygo-${local.services[count.index]}.service",
      "sudo systemctl daemon-reload",
      "sudo systemctl restart flappygo-${local.services[count.index]}"
    ]
  }

  connection {
    type        = "ssh"
    user        = "ec2-user"
    host        = module.microservices[count.index].instance_public_dns
    private_key = file(var.ssh_private_key_path)
  }

  depends_on = [module.microservices]
}

# Client deployment (always needed)
module "client" {
  source = "./modules/compute"
  
  instance_type   = var.instance_type
  ami_id          = var.ami_id
  subnet_id       = module.vpc[0].public_subnets[0]
  security_groups = [module.security_groups[0].sg_id]
  key_name        = var.key_name != "" ? aws_key_pair.deployment_key[0].key_name : null
  create_key_pair = false  # Key pair is managed at root level
  ssh_public_key  = local.ssh_public_key
  
  user_data = templatefile("${path.module}/templates/client_userdata.tpl", {
    aws_region    = var.aws_region
    auth_url      = var.deployment_mode == "microservices" ? module.microservices[index(local.services, "auth")].instance_public_dns : module.monolith[0].instance_public_dns
    initiator_url = var.deployment_mode == "microservices" ? module.microservices[index(local.services, "initiator")].instance_public_dns : module.monolith[0].instance_public_dns
    score_url     = var.deployment_mode == "microservices" ? module.microservices[index(local.services, "score")].instance_public_dns : module.monolith[0].instance_public_dns
    engine_url    = var.deployment_mode == "microservices" ? module.microservices[index(local.services, "engine")].instance_public_dns : module.monolith[0].instance_public_dns
    worldgen_url  = var.deployment_mode == "microservices" ? module.microservices[index(local.services, "worldgen")].instance_public_dns : module.monolith[0].instance_public_dns
    music_url     = var.deployment_mode == "microservices" ? module.microservices[index(local.services, "music")].instance_public_dns : module.monolith[0].instance_public_dns
    auth_port     = var.deployment_mode == "microservices" ? var.service_ports["auth"] : var.service_ports["monolith"]
    initiator_port = var.deployment_mode == "microservices" ? var.service_ports["initiator"] : var.service_ports["monolith"]
    score_port     = var.deployment_mode == "microservices" ? var.service_ports["score"] : var.service_ports["monolith"]
    engine_port    = var.deployment_mode == "microservices" ? var.service_ports["engine"] : var.service_ports["monolith"]
    worldgen_port  = var.deployment_mode == "microservices" ? var.service_ports["worldgen"] : var.service_ports["monolith"]
    music_port     = var.deployment_mode == "microservices" ? var.service_ports["music"] : var.service_ports["monolith"]
    github_token  = var.github_token
    dollar        = "$"
    use_own_certificates = var.use_own_certificates
    certificate_content = local.certificate_content
    private_key_content = local.private_key_content
    certificate_validity_days = var.certificate_validity_days
    spki_hash = local.spki_hash
  })
  
  instance_name = "${var.project_name}-client"
  environment   = var.environment
  service_name  = "client"
}

# Define local output map based on deployment mode
locals {
  service_endpoints = var.deployment_mode == "monolith" ? {
    monolith = module.monolith[0].instance_public_dns
    client   = module.client.instance_public_dns
  } : merge(
    { for s in local.services : s => module.microservices[index(local.services, s)].instance_public_dns },
    { client = module.client.instance_public_dns }
  )
}

# Output the computed service_endpoints
output "service_endpoints" {
  value = local.service_endpoints
}
