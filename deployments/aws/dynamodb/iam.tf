locals {
  service_user = "iconrepo_dynamodb"
}

resource "aws_iam_user" "user" {
  name = local.service_user
}

# https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/iam_access_key
resource "aws_iam_access_key" "user_key" {
  user = aws_iam_user.user.name
  pgp_key = "keybase:${var.pgp_key_owner}"
}

# /!\ ERROR must NOT BE an INLINE policy  
# https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/iam_user_policy
# resource "aws_iam_user_policy" "user_policy" {
#   name   = var.project_name
#   user   = aws_iam_user.user.name
#   policy = data.aws_iam_policy_document.user_policy.json
# }

# https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/iam_policy
resource "aws_iam_policy" "policy" {
  name   = local.service_user
  policy = data.aws_iam_policy_document.user_policy.json
}

# https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/iam_user_policy_attachment
resource "aws_iam_user_policy_attachment" "user_policy_attachment" {
  user       = aws_iam_user.user.name
  policy_arn = aws_iam_policy.policy.arn
}

# https://registry.terraform.io/providers/hashicorp/aws/latest/docs/data-sources/iam_policy_document
data "aws_iam_policy_document" "user_policy" {
  statement {
    actions = [
      "dynamodb:GetItem",
      "dynamodb:Scan",
      "dynamodb:PutItem",
      "dynamodb:UpdateItem",
      "dynamodb:DeleteItem",
    ]

    resources = [
      aws_dynamodb_table.icons.arn,
      aws_dynamodb_table.icon_tags.arn,
      aws_dynamodb_table.icons_locks.arn,
      aws_dynamodb_table.icon_tags_locks.arn,
    ]
  }
}

output "access_key_id" {
  value = aws_iam_access_key.user_key.id
}

output "encrypted_access_key_secret" {
  value = aws_iam_access_key.user_key.encrypted_secret
}
