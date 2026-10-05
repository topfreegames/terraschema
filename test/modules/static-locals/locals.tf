locals {
  tiers = {
    "small"  = { instance_class = "db.t4g.small", replicas = 0 }
    "medium" = { instance_class = "db.r6g.large", replicas = 1 }
    "large"  = { instance_class = "db.r6g.xlarge", replicas = 1 }
  }

  regions = ["us-east-1", "sa-east-1"]

  prefixed_regions = [for region in local.regions : "aws-${region}"]

  from_variable = [var.default_value]
}
