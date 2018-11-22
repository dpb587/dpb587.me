variable "region" { default = "us-east-1" }
variable "owner" { default = "dpb587" }
variable "repository" { default = "dpb587me-bosh-release" } # dpb587.me

provider "aws" {
  region = "${var.region}"
}

#
# output
#

output "ci_access_key" {
  value = "${aws_iam_access_key.ci.id}"
}

output "ci_secret_key" {
  value = "${aws_iam_access_key.ci.secret}"
  sensitive = true
}

output "ci_deploy_ssh_key" {
  value = "${tls_private_key.ci_deploy_ssh_key.private_key_pem}"
  sensitive = true
}

output "ci_deploy_ssh_key_pub" {
  value = "${trimspace(tls_private_key.ci_deploy_ssh_key.public_key_openssh)} ci@terraform"
}

#
# github
#

resource "tls_private_key" "ci_deploy_ssh_key" {
  algorithm = "RSA"
  rsa_bits  = 4096
}

#
# iam
#

resource "aws_iam_user" "ci" {
  name = "${var.repository}-ci"
}

resource "aws_iam_access_key" "ci" {
  user = "${aws_iam_user.ci.name}"
}

#
# s3
#

resource "aws_s3_bucket" "bucket" {
  bucket = "${var.owner}-${var.repository}-${var.region}"
  versioning {
    enabled = true
  }
}

data "aws_iam_policy_document" "ci_s3" {
  statement {
    actions = [
      "s3:GetObject",
      "s3:PutObject",
    ]
    effect = "Allow"
    resources = [
      "${aws_s3_bucket.bucket.arn}/*",
    ]
  }
}

resource "aws_iam_user_policy" "ci_s3" {
  name = "s3"
  user = "${aws_iam_user.ci.name}"
  policy = "${data.aws_iam_policy_document.ci_s3.json}"
}
