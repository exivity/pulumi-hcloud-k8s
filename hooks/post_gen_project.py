import subprocess
import random
import string


def pulumi_config_set(path: str, value: str, secret: bool = False):
    """Set a Pulumi configuration value."""
    cmd = ["pulumi", "config", "set", "--path", path]
    if secret:
        cmd.append("--secret")
    cmd.append(value)
    subprocess.run(cmd, check=True)


def pulumi_config_remove(path: str):
    """Remove a Pulumi configuration value."""
    subprocess.run(["pulumi", "config", "rm", "--path", path], check=True)


def pulumi_stack_init(stack_name: str):
    """Initialize a Pulumi stack."""
    try:
        subprocess.run(["pulumi", "stack", "init", stack_name], check=True)
    except subprocess.CalledProcessError:
        print(
            f"Warning: Pulumi stack '{stack_name}' already exists. Continuing with existing stack."
        )


def pulumi_stack_select(stack_name: str):
    """Select a Pulumi stack."""
    subprocess.run(["pulumi", "stack", "select", stack_name], check=True)


def main():
    # delete the default go.mod file if it exists
    try:
        subprocess.run(["rm", "go.mod"], check=True)
    except subprocess.CalledProcessError:
        pass

    # Create a Go module with the specified module path
    module_path = "{{cookiecutter.go_module_path}}"
    subprocess.run(["go", "mod", "init", module_path], check=True)
    subprocess.run(["go", "mod", "tidy"], check=True)

    # Initialize golangci-lint with a separate mod file
    subprocess.run(
        [
            "go",
            "mod",
            "init",
            "-modfile=golangci-lint.mod",
            "golangci-lint",
        ],
        check=True,
    )
    subprocess.run(
        [
            "go",
            "get",
            "-tool",
            "-modfile=golangci-lint.mod",
            "github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest",
        ],
        check=True,
    )
    print("Golangci-lint initialized with separate mod file.")

    # initialize the Pulumi project and select stack
    stack_name = "{{cookiecutter.pulumi_org}}/{{cookiecutter.pulumi_stack}}"
    pulumi_stack_init(stack_name)
    pulumi_stack_select(stack_name)
    print(f"Pulumi stack '{stack_name}' selected.")

    # Set the Pulumi configuration for the Hetzner token
    if "{{cookiecutter.hetzner_token}}":
        pulumi_config_set(
            "hcloud-k8s:hetzner.token", "{{cookiecutter.hetzner_token}}", secret=True
        )
    if "{{cookiecutter.hetzner_cluster_token}}":
        pulumi_config_set(
            "hcloud-k8s:kubernetes.hcloud_token",
            "{{cookiecutter.hetzner_cluster_token}}",
            secret=True,
        )

    # Set the Talos API allowed CIDRs if provided
    if "{{cookiecutter.talos_api_allowed_cidrs}}":
        print("Setting Talos API allowed CIDRs...")
        cidrs = "{{cookiecutter.talos_api_allowed_cidrs}}".split(",")
        for id in range(len(cidrs)):
            pulumi_config_set(f"hcloud-k8s:firewall.vpn_cidrs[{id}]", cidrs[id].strip())
    else:
        print(
            "No Talos API allowed CIDRs provided. Allowing Talos API access from all IPs."
        )
        pulumi_config_set("hcloud-k8s:firewall.open_talos_api", "true")

    # Set the secretbox encryption secret if not already set
    print("Setting secretbox encryption secret...")
    secretbox_encryption_secret = "".join(
        random.choices(string.ascii_letters + string.digits, k=32)
    )
    pulumi_config_set(
        "hcloud-k8s:talos.secretbox_encryption_secret",
        secretbox_encryption_secret,
        secret=True,
    )

    if "{{cookiecutter.controlplane_enable_ha}}" == "False":
        pulumi_config_remove("hcloud-k8s:control_plane.node_pools[1]")
        pulumi_config_remove("hcloud-k8s:control_plane.node_pools[1]")

    # setup longhorn configuration
    if "{{cookiecutter.enable_longhorn}}" == "True":
        pulumi_config_set("hcloud-k8s:talos.enable_longhorn", "true")
    else:
        pulumi_config_remove("hcloud-k8s:kubernetes.csi")

    # setup hetzner CSI configuration
    if "{{cookiecutter.enable_hetzner_csi}}" == "True":
        print("Hetzner CSI is enabled. Encrypting storage...")
        secretbox_encryption_secret = "".join(
            random.choices(string.ascii_letters + string.digits, k=32)
        )
        pulumi_config_set(
            "hcloud-k8s:kubernetes.csi.encrypted_secret",
            secretbox_encryption_secret,
            secret=True,
        )
    else:
        pulumi_config_remove("hcloud-k8s:kubernetes.csi")

    # remove the cluster autoscaler configuration if not enabled
    if "{{cookiecutter.enable_cluster_autoscaler}}" == "False":
        pulumi_config_remove("hcloud-k8s:kubernetes.cluster_auto_scaler")

    # remove the hetzner kubelet cert approver configuration if not enabled
    if "{{cookiecutter.enable_kubelet_cert_approver}}" == "False":
        pulumi_config_remove("hcloud-k8s:kubernetes.kubelet_serving_cert_approver")

    # remove the kubernetes metrics server configuration if not enabled
    if "{{cookiecutter.enable_metrics_server}}" == "False":
        pulumi_config_remove("hcloud-k8s:kubernetes.kubernetes_metrics_server")


if __name__ == "__main__":
    main()
