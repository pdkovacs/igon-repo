resource "aws_dynamodb_table" "icons" {
  name           = "icons"
  billing_mode   = "PROVISIONED"
  read_capacity  = 5
  write_capacity = 5
  hash_key       = "IconName"

  attribute {
    name = "IconName" # <icon-id>#<icon-name>
    type = "S"
  }
}

resource "aws_dynamodb_table" "icon_tags" {
  name           = "icon_tags"
  billing_mode   = "PROVISIONED"
  read_capacity  = 5
  write_capacity = 5
  hash_key       = "Tag"


  attribute {
    name = "Tag"
    type = "S"
  }
}

resource "aws_dynamodb_table" "icons_locks" {
  name           = "icons_locks"
  billing_mode   = "PROVISIONED"
  read_capacity  = 5
  write_capacity = 5
  hash_key       = "key"

  attribute {
    name = "key"
    type = "S"
  }
}

resource "aws_dynamodb_table" "icon_tags_locks" {
  name           = "icon_tags_locks"
  billing_mode   = "PROVISIONED"
  read_capacity  = 5
  write_capacity = 5
  hash_key       = "key"


  attribute {
    name = "key"
    type = "S"
  }
}
