terraform {
  backend "s3" {
    bucket  = "bitkitchen-tf-state"
    key     = "iconrepo/indexing/dynamodb"
    region  = "eu-west-1"
    encrypt = true
  }
}
