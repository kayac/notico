terraform {
  required_version = "~> 1.10.2"

  required_providers {
    aws = {
      version = "~> 5.81.0"
    }
  }

  backend "s3" {
  }
}

provider "aws" {
  region = var.region

  default_tags {
    tags = {
      ManagedBy = "Terraform"
      Service   = "notico"
      Repo      = "kayac/notico"
    }
  }
}
