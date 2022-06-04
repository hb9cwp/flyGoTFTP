# from
#  https://registry.terraform.io/providers/fly-apps/fly/latest

# Data sources allow Terraform to use information defined outside of Terraform,
# defined by another separate Terraform configuration, or modified by functions.
#  https://www.terraform.io/language/data-sources

# just do:
#  $ terraform init
#  $ terraform validate
#  $ terraform plan
#  $ terraform apply

terraform {
  required_providers {
    fly = {
      source = "fly-apps/fly"
      version = "0.0.6"
    }
  }
}

provider "fly" {
  #  https://registry.terraform.io/providers/DAlperin/fly-io/latest/docs#example-usage
  # Don't do this:
  #fly_api_token = "abc123" # If not set checks env for FLY_TOKEN

  # Use the FLY_API_TOKEN env variable instead.
  # $ export FLY_API_TOKEN=$(fly auth token)
  # $ echo $FLY_API_TOKEN
}

data "fly_app" "exampleApp" {
  # https://registry.terraform.io/providers/DAlperin/fly-io/latest/docs/data-sources/app
  name = "fly-gotftp-69"
  #depends_on = [fly_app.exampleApp]	<=== error in examples, Data does not depend on any (imported) Resources!
}

output "Data_app" {
  value = data.fly_app.exampleApp
}

output "Data_app_IPaddrs" {
  value = data.fly_app.exampleApp.ipaddresses
}
