terraform {
  backend "s3" {
    encrypt        = true
    bucket         = "iam-enforcer-terraform-state"
    key            = "public-iam-enforcer-lambda/terraform.tfstate"
    region         = "eu-central-1"
    dynamodb_table = "terraform-locks"
  }
}