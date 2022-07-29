# https://static.kancloud.cn/yunduanio/terraform-alicloud

# terraform 加载变量按照以下的顺序
# terraform.tfvars
# terraform.tfvars.json
# *.auto.tfvars or *.auto.tfvars.json
# any -var and -var-file options on the command line, in the order they are provided.

# 三种变量的区别
# Input variables are like function arguments.
# Output values are like function return values.
# Local values are like a function's temporary local variables.

terraform {
  backend "oss" {
    bucket = "bucket-website-my-00001"
    endpoint = "oss-cn-qingdao.aliyuncs.com"
    region = "cn-qingdao"

    tablestore_endpoint = "https://myinstance.cn-qingdao.ots.aliyuncs.com"
    tablestore_table = "mytable"
    profile = "course"
  }
}

data "alicloud_zones" "default" {
  available_resource_creation = "VSwitch"
}

resource "alicloud_vpc" "default" {
  cidr_block = "172.16.0.0/12"
  vpc_name   = "VpcConfig"
}

resource "alicloud_vswitch" "default" {
  vpc_id            = alicloud_vpc.default.id
  vswitch_name      = "vswitch"
  cidr_block        = cidrsubnet(alicloud_vpc.default.cidr_block, 4, 4)
  zone_id           = data.alicloud_zones.default.ids.0
}

resource "alicloud_network_acl" "default" {
  vpc_id           = alicloud_vpc.default.id
  network_acl_name = "network_acl"
  description      = "network_acl"
  ingress_acl_entries {
    description            = "tf-testacc"
    network_acl_entry_name = "tcp23"
    source_cidr_ip         = "196.168.2.0/21"
    policy                 = "accept"
    port                   = "22/80"
    protocol               = "tcp"
  }
  egress_acl_entries {
    description            = "tf-testacc"
    network_acl_entry_name = "tcp23"
    destination_cidr_ip    = "0.0.0.0/0"
    policy                 = "accept"
    port                   = "-1/-1"
    protocol               = "all"
  }
  resources {
    resource_id   = alicloud_vswitch.default.id
    resource_type = "VSwitch"
  }
}

resource "alicloud_eip" "eip" {
  bandwidth            = "10"

  # Internet charge type of the EIP, Valid values are PayByBandwidth, PayByTraffic. Default to PayByBandwidth. From version 1.7.1, default to PayByTraffic. It is only PayByBandwidth when instance_charge_type is PrePaid

  internet_charge_type = "PayByTraffic"

  # Elastic IP instance charge type. Valid values are "PrePaid" and "PostPaid". Default to "PostPaid".
  payment_type = "PostPaid"
}

resource "alicloud_eip_association" "eip_asso" {
  allocation_id = alicloud_eip.eip.id
  instance_id   = alicloud_instance.instance.id
}

# 使用file provisioner 拷贝文件到服务器
resource "null_resource" "copy" {
  # 等待server eip eip绑定完毕后再执行拷贝文件的操作
  depends_on = [alicloud_instance.instance,alicloud_eip.eip]
  triggers = {
    key = "${uuid()}"
  }

  provisioner "file" {
    source      = "./html/"
    destination = "/usr/share/nginx/html/"

    connection {
      type     = "ssh"
      user     = "root"
      password = var.ecs_password
      host     = "${alicloud_eip.eip.ip_address}"
    }
  }
}
variable "ecs_password" {
  default = "12345678"
}
resource "alicloud_dns_record" "dns" {
  name        = "xxxxxxx.com"
  host_record = "www"
  type        = "A"
  value       = "${alicloud_instance.instance.public_ip}"
}

resource "alicloud_slb_server_certificate" "foo" {
  name               = "slbservercertificate"
  server_certificate = file("${path.module}/server_certificate.pem")
  private_key        = file("${path.module}/private_key.pem")
}

resource "alicloud_slb_listener" "default" {
  # ...
  server_certificate_id=alicloud_slb_server_certificate.foo.id
}

resource "alicloud_disk" "disk" {
  zone_id = alicloud_instance.instance.availability_zone
  category          = "cloud_ssd"
  size              = 200
  count             = 1
}

resource "alicloud_ecs_auto_snapshot_policy" "policy" {
  name            = "tf-testAcc"
  repeat_weekdays = ["5"]
  retention_days  = -1
  time_points     = ["1","12"]
}

resource "alicloud_ecs_auto_snapshot_policy_attachment" "attachment" {
  auto_snapshot_policy_id = alicloud_ecs_auto_snapshot_policy.policy.id
  disk_id                 = alicloud_disk.disk.id
}