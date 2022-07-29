#provider "alicloud" {
#  region                  = var.region != "" ? var.region : null
#}
#
#locals {
#  backend_appname         = "${var.bu}-tf"
#  oss_bucket_name         = "${local.backend_appname}-${var.env}-10001-oss"
#  ots_instance_name       = var.ots_create_instance == true ? "${local.backend_appname}-${var.env}-01" : var.existing_ots_instance_name
#  ots_create_table        = var.ots_create_instance == true ? true : var.ots_create_table
#  ots_table_name          = var.ots_create_table == true ? "${local.backend_appname}-${var.env}-10001-tb" : var.ots_table_name
#  cmk_alias_name          = "${local.backend_appname}-${var.env}-10001-cmk"
#  kms_existing_key_id     = var.kms_create_cmk == true ? module.remote-backend-kms.this_kms_key_id[0] : var.kms_existing_key_id
#  tablestore_endpoint     = "https://${local.ots_instance_name}.${var.region}.ots.aliyuncs.com"
#}
#
#module "remote-backend-oss" {
#  source          = "terraform-alicloud-modules/terraform-alicloud-oss-bucket"
#  bucket_name     = local.oss_bucket_name
#  acl             = "private"
#  versioning      = "Enabled"
#  region          = var.region
#  redundancy_type = "ZRS"
#  lifecycle_rule  = [
#    {
#      id      = "oss-backend-10001-rule"
#      prefix  = ""
#      enabled = true
#      expiration = [
#        {
#          days = 365
#        },
#      ]
#    },
#  ]
#  server_side_encryption_rule = [
#    {
#      sse_algorithm      = "KMS"
#      kms_master_key_id  = local.kms_existing_key_id
#    },
#  ]
#  tags = {
#    creater = "terraform"
#    appname = local.backend_appname
#    env     = var.env
#    bu      = var.bu
#  }
#}
#
#module "remote-backend-kms" {
#  source                     = "terraform-alicloud-modules/terraform-alicloud-kms"
#  create_cmk                 = var.kms_create_cmk
#  existing_key_id            = local.kms_existing_key_id
#  alias_name                 = local.cmk_alias_name
#  region                     = var.region
#  protection_level           = "HSM"
#  key_spec                   = "Aliyun_AES_256"
#  pending_window_in_days     = "30"
#  status                     = "Enabled"
#  encrypt                    = false
#  decrypt                    = false
#}
#
#module "remote-backend-ots" {
#  source                     = "terraform-alicloud-modules/terraform-alicloud-table-store"
#  create_instance            = var.use_ots ? var.ots_create_instance : false
#  use_existing_instance      = var.use_ots && var.ots_create_instance == false ? false : true
#  create_table               = local.ots_create_table
#  region                     = var.region
#  instance_name              = local.ots_instance_name
#  existing_ots_instance_name = local.ots_instance_name
#  table_name                 = replace(local.ots_table_name, "-", "_")
#  accessed_by                = "Any"
#  instance_type              = "HighPerformance"
#  primary_key = [
#    {
#      name = "LockID"
#      type = "String"
#    },
#  ]
#  time_to_live     = -1
#  max_version      = 1
#  tags = {
#    creater = "terraform"
#    appname = local.backend_appname
#    env     = var.env
#    bu      = var.bu
#  }
#}
#
#resource "local_file" "this" {
#  content                  = <<EOF
#    terraform {
#      backend "oss" {
#        bucket              = "${local.oss_bucket_name}"
#        prefix              = "${var.bu}/${var.state_path}"
#        key                 = "${var.state_name}"
#        acl                 = "${var.state_acl}"
#        region              = "${var.region}"
#        encrypt             = "${var.encrypt_state}"
#        %{ if var.use_ots == true }tablestore_endpoint = "${local.tablestore_endpoint}"
#        tablestore_table    = "${replace(local.ots_table_name, "-", "_")}"%{ endif }
#      }
#    }
#    EOF
#  file_permission         = 0644
#  filename                = "${path.root}/backend.tf"
#}