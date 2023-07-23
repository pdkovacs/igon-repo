provider "aws" {
  access_key                  = "mockAccessKey"
  region                      = "us-east-1"
  secret_key                  = "mockSecretKey"
  skip_credentials_validation = true
  skip_metadata_api_check     = true
  skip_requesting_account_id  = true

  endpoints {
    dynamodb = "http://localhost:8000"
  }
}

resource "aws_dynamodb_table" "icons" {
  name           = "icons"
  billing_mode   = "PROVISIONED"
  read_capacity  = 20
  write_capacity = 20
  hash_key       = "IconName"
  range_key      = "IconfileDescriptor"

  attribute {
    name = "IconName" # <icon-id>#<icon-name>
    type = "S"
  }

  attribute {
    name = "IconfileDescriptor" # <format>#<size>
    type = "S"
  }

  ttl {
    attribute_name = "TimeToExist"
    enabled        = false
  }

  tags = {
    Name        = "dynamodb-table-icons"
    Environment = "dev"
  }
}

resource "aws_dynamodb_table" "icon_files" {
  name           = "icon_files"
  billing_mode   = "PROVISIONED"
  read_capacity  = 20
  write_capacity = 20
  hash_key       = "Tag"
  range_key      = "IconId"

  attribute {
    name = "Tag"
    type = "S"
  }

  attribute {
    name = "IconId"
    type = "S"
  }

  ttl {
    attribute_name = "TimeToExist"
    enabled        = false
  }

  tags = {
    Name        = "dynamodb-table-icon_files"
    Environment = "dev"
  }
}
