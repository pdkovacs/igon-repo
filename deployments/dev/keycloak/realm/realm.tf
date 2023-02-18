terraform {
  required_providers {
    keycloak = {
      source = "mrparkers/keycloak"
      version = "4.0.0"
    }
  }
}

provider "keycloak" {
    client_id     = "terraform"
    client_secret = "884e0f95-0f42-4a63-9b1f-94274655669e"
    url           = "${var.keycloak_url}"
}

resource "keycloak_realm" "realm" {
  realm = "my-realm"

  attributes = {
    userProfileEnabled = true
  }
}

resource "keycloak_openid_client" "iconrepo" {
  realm_id            = keycloak_realm.realm.id
  client_id           = "iconrepo"
  client_secret       = "Xb5BtE9RvMWCjJGfeMJDYOWIZGKSMm3z"

  name                = "Icon Repository"
  enabled             = true

  access_type         = "CONFIDENTIAL"
  valid_redirect_uris = [
    "http://${var.app_hostname}:8080/openid-callback" # app
    # "http://${var.app_hostname}:4180/oauth2/callback", # oauth-proxy
    # "http://${var.app_hostname}:9999/oauth2/callback"  # load-balancer (nginx or simple_router)
  ]
  standard_flow_enabled = true

  login_theme = "keycloak"
}

resource "keycloak_openid_group_membership_protocol_mapper" "group_membership_mapper" {
  realm_id  = keycloak_realm.realm.id
  client_id = keycloak_openid_client.iconrepo.id
  name      = "group-membership-mapper"

  claim_name = "groups"
  full_path = false
}

variable "app_hostname" {
  type = string
  default = "iconrepo"
}

variable "keycloak_url" {
  type = string
  default = "http://keycloak:8080"
}
