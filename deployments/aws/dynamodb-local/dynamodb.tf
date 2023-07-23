provider "aws" {
  access_key                  = "mockAccessKey"
  region                      = "eu-west-1"
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

  attribute {
    name = "IconName" # <icon-id>#<icon-name>
    type = "S"
  }

}

resource "aws_dynamodb_table" "icon_tags" {
  name           = "icon_tags"
  billing_mode   = "PROVISIONED"
  read_capacity  = 20
  write_capacity = 20
  hash_key       = "Tag"


  attribute {
    name = "Tag"
    type = "S"
  }

}
