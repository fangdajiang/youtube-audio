provider "alicloud" {}

variable "vpc_name" {
  default = "terraform_test_ecs_config"
}

resource "alicloud_vpc" "vpc" {
  vpc_name       = var.vpc_name
  cidr_block = "172.16.0.0/12"
}

resource "alicloud_vswitch" "vsw" {
  vpc_id            = alicloud_vpc.vpc.id
  cidr_block        = "172.16.0.0/21"
  zone_id = data.alicloud_zones.default.zones[0].id
}

resource "alicloud_instance" "youtube-audio" {
  key_name = "hk-test"

  # 可用区
  #availability_zone = "cn-shanghai"
  # 绑定安全组
  security_groups = alicloud_security_group.sg.*.id

  # 实例规格
  instance_type        = "ecs.n1.small"
  # 系统盘类型
  system_disk_category = "cloud_efficiency"
  # 系统镜像 ubuntu_22_04_x64_20G_alibase_20220628.vhd
  image_id             = "m-j6c2yv3ppzgbhft9y4f7"
  # 实例名称
  instance_name        = "centos-youtube-audio"
  # 所在交换机
  vswitch_id = alicloud_vswitch.vsw.id
  # 公网带宽，设置internet_max_bandwidth_out > 0 可以分配一个public IP
  internet_max_bandwidth_out = 10

  instance_charge_type = "PostPaid"
  internet_charge_type = "PayByTraffic"
}

data "alicloud_zones" "default" {
  available_disk_category     = "cloud_efficiency"
  available_resource_creation = "VSwitch"
  available_instance_type = data.alicloud_instance_types.instance_type.instance_types[0].id
}

data "alicloud_instance_types" "instance_type" {
  instance_type_family = "ecs.n1"
  cpu_core_count       = "1"
  memory_size          = "2"
}

resource "alicloud_security_group" "sg" {
  name = "audio"
  security_group_type = "normal"
  vpc_id = alicloud_vpc.vpc.id
  description = "Terraform created."
}

resource "alicloud_security_group_rule" "icmp" {
  description       = "ping allowed"
  type              = "ingress"
  ip_protocol       = "icmp"
  nic_type          = "intranet"
  policy            = "accept"
  port_range        = "-1/-1"
  priority          = 100
  security_group_id = alicloud_security_group.sg.id
  cidr_ip           = "0.0.0.0/0"
}
resource "alicloud_security_group_rule" "allow_22" {
  description       = "ssh only"
  type              = "ingress"
  # tcp/udp/icmp,gre,all
  ip_protocol       = "tcp"
  # the default value is internet
  nic_type          = "intranet"
  # accept/drop
  policy            = "accept"
  # Default to "-1/-1". When the protocol is tcp or udp, each side port number range from 1 to 65535 and '-1/-1' will be invalid. For example, 1/200 means that the range of the port numbers is 1-200. Other protocols' 'port_range' can only be "-1/-1", and other values will be invalid
  port_range        = "22/22"
  priority          = 100
  security_group_id = alicloud_security_group.sg.id
  cidr_ip           = "0.0.0.0/0"
}