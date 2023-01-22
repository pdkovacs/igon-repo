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

resource "keycloak_user" "user" {
  realm_id = keycloak_realm.realm.id
  username = "bob"
  enabled  = true

  email      = "bob@domain.com"
  first_name = "Bob"
  last_name  = "Bobson"
}

resource "keycloak_group" "icon_editor" {
  realm_id = keycloak_realm.realm.id
  name     = "ICON_EDITOR"
}

resource "keycloak_user_groups" "user_groups" {
  realm_id = keycloak_realm.realm.id
  user_id = keycloak_user.alice.id

  group_ids  = [
    keycloak_group.icon_editor.id
  ]
}

resource "keycloak_user" "alice" {
  realm_id   = keycloak_realm.realm.id
  username   = "alice"
  enabled    = true

  email      = "alice@domain.com"
  first_name = "Alice"
  last_name  = "Aliceberg"

  initial_password {
    value     = "al1ce"
    temporary = false
  }
}

resource "keycloak_user" "joe" {
  realm_id   = keycloak_realm.realm.id
  username   = "joe"
  enabled    = true

  email      = "joe@domain.com"
  first_name = "Joe"
  last_name  = "Joeberg"

  initial_password {
    value     = "j0e"
    temporary = false
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
    "http://${var.app_hostname}:8091/openid-callback", # app
    "http://${var.app_hostname}:4180/oauth2/callback", # oauth-proxy
    "http://${var.app_hostname}:9999/oauth2/callback"  # load-balancer (nginx or simple_router)
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
