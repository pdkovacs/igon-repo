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

data "keycloak_realm" "realm" {
    realm = "my-realm"
}

locals {
  users = jsondecode(file(pathexpand(var.user_list_file)))
  groups = tolist(toset(flatten([for user in local.users : user.groups])))
}

resource "keycloak_user" "all_users" {
  count      = length(local.users)
  realm_id   = data.keycloak_realm.realm.id
  username   = local.users[count.index].username
  enabled    = true

  email      = local.users[count.index].email
  first_name = local.users[count.index].first_name
  last_name  = local.users[count.index].last_name

  initial_password {
    value     = local.users[count.index].password
    temporary = false
  }
}

resource "keycloak_group" "iconrepo" {
  count    = length(local.groups)
  realm_id = data.keycloak_realm.realm.id
  name     = local.groups[count.index]
}

resource "keycloak_user_groups" "user_groups" {
  count    = length(local.users)
  realm_id = data.keycloak_realm.realm.id
  user_id = [for x in keycloak_user.all_users :  x.id if x.username == local.users[count.index].username][0]

  group_ids  = [for x in matchkeys(keycloak_group.iconrepo, local.groups, local.users[count.index].groups) : x.id]
}

variable "app_hostname" {
  type = string
  default = "iconrepo"
}

variable "keycloak_url" {
  type = string
  default = "http://keycloak:8080"
}

variable "user_list_file" {
  default = "~/.icon-repo.users"
}
