variable "name" {
  default = "terraform_test_ecs_slb_config"
}

variable "vpc_cidr_block" {
  default = "172.16.0.0/16"
}

variable "vsw_cidr_block" {
  default = "172.16.0.0/24"
}

# Create a new ECS instance for a VPC
resource "alicloud_security_group" "sg" {
  name        = "tf_test_foo"
  description = "foo"
  vpc_id      = alicloud_vpc.vpc.id
}

resource "alicloud_kms_key" "key" {
  description            = "Hello KMS"
  pending_window_in_days = "7"
  status              = "Enabled"
}

data "alicloud_instance_types" "instance_type" {
  instance_type_family = "ecs.n1"
  cpu_core_count       = "1"
  memory_size          = "1"
}

data "alicloud_zones" "default" {
  available_disk_category     = "cloud_efficiency"
  available_resource_creation = "VSwitch"
  available_instance_type = data.alicloud_instance_types.instance_type.instance_types[0].id
}

# Create a new ECS instance for VPC
resource "alicloud_vpc" "vpc" {
  vpc_name       = var.name
  cidr_block = var.vpc_cidr_block
}

resource "alicloud_vswitch" "vswitch" {
  vpc_id            = alicloud_vpc.vpc.id
  cidr_block        = var.vsw_cidr_block
  zone_id           = data.alicloud_zones.default.zones[0].id
  vswitch_name      = var.name
}

resource "alicloud_instance" "instance" {
  # cn-beijing
  availability_zone = data.alicloud_zones.default.zones[0].id
  # [alicloud_security_group.sg.id]
  security_groups   = alicloud_security_group.sg.*.id

  # series III
  instance_type              = "ecs.n4.large"
  system_disk_category       = "cloud_efficiency"
  system_disk_name           = "test_foo_system_disk_name"
  system_disk_description    = "test_foo_system_disk_description"
  image_id                   = "ubuntu_18_04_64_20G_alibase_20190624.vhd"
  instance_name              = "test_foo"
  vswitch_id                 = alicloud_vswitch.vswitch.id
  internet_max_bandwidth_out = 10
  data_disks {
    name        = "disk2"
    size        = 20
    category    = "cloud_efficiency"
    description = "disk2"
    encrypted   = true
    kms_key_id  = alicloud_kms_key.key.id
  }
}
resource "alicloud_slb" "default" {
  load_balancer_name          = var.name
  load_balancer_spec = "slb.s2.small"
  vswitch_id    = alicloud_vswitch.vswitch.id
}