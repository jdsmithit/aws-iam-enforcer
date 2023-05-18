locals {
  artifacts_bucket_name = "iam-enforcer-artifacts"
}

module "artifacts_bucket" {
  source = "git::https://github.com/dfds/aws-modules-s3.git?ref=v1.1.0"

  bucket_name                     = local.artifacts_bucket_name
  bucket_versioning_configuration = "Enabled"
  object_ownership                = "BucketOwnerPreferred"
  create_policy                   = true

  logging_bucket_name   = "iam-enforcer-s3-logs"
  source_policy_documents = [data.aws_iam_policy_document.this.json]

}

data "aws_organizations_organization" "this" {}

data "aws_iam_policy_document" "this" {
  statement {
    sid = "AllowIamAccess"
    effect = "Allow"
    principals {
      identifiers = ["*"]
      type        = "AWS"
    }
    resources = [
      "arn:aws:iam:::/*"
    ]
    actions = [
      "iam:DeleteUser",
      "iam:DeleteAccessKey",
      "iam:ListAccessKeys",
      "iam:UpdateAccessKey"
    ]
  }
}