locals {
  artifacts_bucket_name = "aws-iam-enforcer-artifacts"
  lambda_name = "aws-iam-enforcer"
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
    sid = "AllowOrganizationAccess"
    effect = "Allow"
    principals {
      identifiers = ["*"]
      type        = "AWS"
    }
    resources = [
      "arn:aws:s3:::${local.artifacts_bucket_name}/*",
      "arn:aws:s3:::${local.artifacts_bucket_name}"
    ]
    actions = [
      "s3:GetObject",
      "s3:GetBucketTagging",
      "s3:GetObjectTagging",
      "s3:GetObjectVersionTagging"
    ]
    condition {
      test     = "StringEquals"
      values   = [data.aws_organizations_organization.this.id]
      variable = "aws:PrincipalOrgID"
    }
  }
}

data "aws_s3_object" "lambda" {
  bucket = var.artifacts_bucket_name
  key    = "${local.lambda_name}-lambda.zip"
}

resource "aws_lambda_function" "lambda" {
  function_name                  = local.lambda_name
  role                           = aws_iam_role.lambda.arn
  timeout                        = 120
  memory_size                    = 512
  reserved_concurrent_executions = 1
  runtime = "go1.x"
  handler       = "main"
  s3_bucket = local.artifacts_bucket_name
  s3_key = "${local.lambda_name}-lambda.zip"
  source_code_hash = data.aws_s3_object.lambda.etag

  environment {
    variables = {
      DRY_RUN_MODE = true
      DISABLE_KEYS_TOGGLE = true
      DISABLE_KEYS_DAYS = 90
      DISABLE_KEYS_TOGGLE = true
      DISABLE_UNUSED_ONLY_TOGGLE = 90
    }
  }

  tracing_config {
    mode = "Active"
  }
}

resource "aws_iam_role" "lambda" {
  name               = local.lambda_name
  assume_role_policy = data.aws_iam_policy_document.lambda_trust.json
}

data "aws_iam_policy_document" "lambda_trust" {
  statement {
    actions = ["sts:AssumeRole"]
    principals {
      identifiers = ["lambda.amazonaws.com"]
      type        = "Service"
    }
  }
}

resource "aws_iam_role_policy_attachment" "lambda" {
  policy_arn = aws_iam_policy.lambda_access.arn
  role       = aws_iam_role.lambda.name
}

resource "aws_iam_policy" "lambda_access" {
  name        = "${local.lambda_name}-lambda-access"
  description = "Access policy for the ${local.lambda_name} lambda"
  policy      = data.aws_iam_policy_document.lambda_access.json
}

data "aws_caller_identity" "current" {}

output "account_id" {
  value = data.aws_caller_identity.current.account_id
}

data "aws_iam_policy_document" "lambda_access" {
  statement {
    sid = "iamAccess"
    actions = [
      "iam:GetAccessKeyLastUsed",
      "iam:ListAccessKeys",
      "iam:ListUsers",
      "iam:UpdateAccessKey"
    ]
    resources = [
      "arn:aws:iam::${data.aws_caller_identity.current.account_id}:user/*"
    ]
  }
}