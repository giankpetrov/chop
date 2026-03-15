package filters

import (
	"strings"
	"testing"
)

func TestFilterTerraformPlan(t *testing.T) {
	raw := `
Terraform used the selected providers to generate the following execution plan. Resource actions are indicated with the following symbols:
  + create
  ~ update in-place
  - destroy

Terraform will perform the following actions:

  # aws_instance.web will be created
  + resource "aws_instance" "web" {
      + ami                          = "ami-0c55b159cbfafe1f0"
      + arn                          = (known after apply)
      + associate_public_ip_address  = true
      + availability_zone            = (known after apply)
      + cpu_core_count               = (known after apply)
      + cpu_threads_per_core         = (known after apply)
      + disable_api_stop             = (known after apply)
      + disable_api_termination      = (known after apply)
      + ebs_optimized                = (known after apply)
      + get_password_data            = false
      + host_id                      = (known after apply)
      + host_resource_group_arn      = (known after apply)
      + iam_instance_profile         = (known after apply)
      + id                           = (known after apply)
      + instance_initiated_shutdown_behavior = (known after apply)
      + instance_state               = (known after apply)
      + instance_type                = "t3.micro"
      + ipv6_address_count           = (known after apply)
      + ipv6_addresses               = (known after apply)
      + key_name                     = "my-key"
      + monitoring                   = false
      + outpost_arn                  = (known after apply)
      + password_data                = (known after apply)
      + placement_group              = (known after apply)
      + placement_partition_number   = (known after apply)
      + primary_network_interface_id = (known after apply)
      + private_dns                  = (known after apply)
      + private_ip                   = (known after apply)
      + public_dns                   = (known after apply)
      + public_ip                    = (known after apply)
      + secondary_private_ips        = (known after apply)
      + security_groups              = (known after apply)
      + source_dest_check            = true
      + subnet_id                    = (known after apply)
      + tags                         = {
          + "Name" = "web-server"
        }
      + tags_all                     = {
          + "Name" = "web-server"
        }
      + tenancy                      = (known after apply)
      + user_data                    = (known after apply)
      + user_data_base64             = (known after apply)
      + user_data_replace_on_change  = false
      + vpc_security_group_ids       = (known after apply)
    }

  # aws_s3_bucket.assets will be created
  + resource "aws_s3_bucket" "assets" {
      + acceleration_status         = (known after apply)
      + acl                         = (known after apply)
      + arn                         = (known after apply)
      + bucket                      = "my-assets-bucket"
      + bucket_domain_name          = (known after apply)
      + bucket_prefix               = (known after apply)
      + bucket_regional_domain_name = (known after apply)
      + force_destroy               = false
      + hosted_zone_id              = (known after apply)
      + id                          = (known after apply)
      + object_lock_enabled         = (known after apply)
      + policy                      = (known after apply)
      + region                      = (known after apply)
      + request_payer               = (known after apply)
      + tags                        = {
          + "Environment" = "production"
        }
      + tags_all                    = {
          + "Environment" = "production"
        }
      + website_domain              = (known after apply)
      + website_endpoint            = (known after apply)
    }

  # aws_security_group.web will be updated in-place
  ~ resource "aws_security_group" "web" {
        id                     = "sg-12345678"
        name                   = "web-sg"
      ~ description            = "Allow HTTP" -> "Allow HTTP and HTTPS"
      ~ ingress                = [
          + {
              + cidr_blocks      = ["0.0.0.0/0"]
              + from_port        = 443
              + protocol         = "tcp"
              + to_port          = 443
            },
          # (1 unchanged element hidden)
        ]
        tags                   = {
            "Name" = "web-sg"
        }
    }

  # aws_db_instance.main will be updated in-place
  ~ resource "aws_db_instance" "main" {
        id                     = "mydb"
        name                   = "production"
      ~ instance_class         = "db.t3.micro" -> "db.t3.small"
      ~ allocated_storage      = 20 -> 50
        engine                 = "postgres"
        engine_version         = "15.4"
        multi_az               = false
        storage_type           = "gp3"
        tags                   = {}
    }

  # aws_route53_record.www will be destroyed
  - resource "aws_route53_record" "www" {
      - id      = "Z1234567890_www.example.com_A"
      - name    = "www.example.com"
      - records = ["1.2.3.4"]
      - ttl     = 300
      - type    = "A"
      - zone_id = "Z1234567890"
    }

Plan: 2 to add, 2 to change, 1 to destroy.

─────────────────────────────────────────────────────────────────────────────

Note: You didn't use the -out option to save this plan, so Terraform can't
guarantee to take exactly these actions if you run "terraform apply" now.
`

	got, err := filterTerraformPlan(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should have summary
	if !strings.Contains(got, "2 to add, 2 to change, 1 to destroy") {
		t.Errorf("expected plan summary, got:\n%s", got)
	}

	// Should have each resource on one line
	if !strings.Contains(got, "~ aws_instance.web (create)") {
		t.Errorf("expected aws_instance.web create, got:\n%s", got)
	}
	if !strings.Contains(got, "~ aws_s3_bucket.assets (create)") {
		t.Errorf("expected aws_s3_bucket.assets create, got:\n%s", got)
	}
	if !strings.Contains(got, "~ aws_security_group.web (update)") {
		t.Errorf("expected aws_security_group.web update, got:\n%s", got)
	}
	if !strings.Contains(got, "~ aws_route53_record.www (destroy)") {
		t.Errorf("expected aws_route53_record.www destroy, got:\n%s", got)
	}

	// Should show changed attributes for updates
	if !strings.Contains(got, "description") {
		t.Errorf("expected changed attribute 'description' for security group, got:\n%s", got)
	}
	if !strings.Contains(got, "instance_class") {
		t.Errorf("expected changed attribute 'instance_class' for db, got:\n%s", got)
	}

	// Should NOT contain full attribute blocks
	if strings.Contains(got, "known after apply") {
		t.Errorf("should not contain 'known after apply' noise, got:\n%s", got)
	}
	if strings.Contains(got, "-out option") {
		t.Errorf("should not contain -out note, got:\n%s", got)
	}

	// Token savings >= 75%
	rawTokens := len(strings.Fields(raw))
	filteredTokens := len(strings.Fields(got))
	savings := 100.0 - float64(filteredTokens)/float64(rawTokens)*100.0
	if savings < 75.0 {
		t.Errorf("expected >=75%% token savings, got %.1f%% (raw=%d, filtered=%d)", savings, rawTokens, filteredTokens)
	}
}

func TestFilterTerraformPlanNoChanges(t *testing.T) {
	raw := `
aws_instance.web: Refreshing state... [id=i-0123456789abcdef0]
aws_s3_bucket.assets: Refreshing state... [id=my-assets-bucket]

No changes. Your infrastructure matches the configuration.

Terraform has compared your real infrastructure against your configuration
and found no differences, so no changes are needed.
`

	got, err := filterTerraformPlan(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got != "no changes" {
		t.Errorf("expected 'no changes', got:\n%s", got)
	}
}

func TestFilterTerraformPlanEmpty(t *testing.T) {
	got, err := filterTerraformPlan("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "" {
		t.Errorf("expected empty, got %q", got)
	}
}

func TestFilterTerraformPlanWithErrors(t *testing.T) {
	raw := `
Error: Reference to undeclared resource

  on main.tf line 15, in resource "aws_instance" "web":
  15:   subnet_id = aws_subnet.main.id

A managed resource "aws_subnet" "main" has not been declared in the root module.

Error: Invalid reference

  on main.tf line 20:
  20:   vpc_id = var.undefined_var
`

	got, err := filterTerraformPlan(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(got, "Error: Reference to undeclared resource") {
		t.Errorf("expected error preserved, got:\n%s", got)
	}
	if !strings.Contains(got, "Error: Invalid reference") {
		t.Errorf("expected second error preserved, got:\n%s", got)
	}
}

func TestOpenTofuRoutesToTerraformFilter(t *testing.T) {
	for _, sub := range []string{"plan", "apply", "init"} {
		f := get("tofu", []string{sub})
		if f == nil {
			t.Errorf("expected filter for tofu %s, got nil", sub)
		}
	}
}
