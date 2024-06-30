locals {
  keycloak_url = "http://keycloak:8080"
  client_id    = "iconrepo"
  client_name  = "Icon Repository"
  app_hostname = "iconrepo.local.com"
}

terraform {
  required_providers {
    keycloak = {
      source = "mrparkers/keycloak"
      version = "4.4.0"
    }
  }
}

provider "keycloak" {
    client_id     = "terraform"
    client_secret = "${var.tf_client_secret}"
    url           = "${local.keycloak_url}"
}

resource "keycloak_realm" "realm" {
  realm = "my-realm"
}

resource "keycloak_openid_client" "iconrepo_client" {
  realm_id            = keycloak_realm.realm.id
  client_id           = "${local.client_id}"
  client_secret       = "${var.client_secret}"

  name                = "${local.client_name}"
  enabled             = true

  access_type         = "CONFIDENTIAL"
  valid_redirect_uris = [
    "http://${local.app_hostname}/*"
  ]
  standard_flow_enabled = true

  login_theme = "keycloak"
}

resource "keycloak_openid_group_membership_protocol_mapper" "iconrepo_group_membership_mapper" {
  realm_id  = keycloak_realm.realm.id
  client_id = keycloak_openid_client.iconrepo_client.id
  name      = "group-membership-mapper"

  claim_name = "groups"
  full_path = false
}

variable "tf_client_secret" {
  type = string
  default = "iconrepo.local.com"
}

variable "client_secret" {
  type = string
}
