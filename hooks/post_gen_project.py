import subprocess


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
    try:
        subprocess.run(
            [
                "pulumi",
                "stack",
                "init",
                "{{cookiecutter.pulumi_org}}/{{cookiecutter.pulumi_stack}}",
            ],
            check=True,
        )
    except subprocess.CalledProcessError:
        print(
            "Warning: Pulumi stack '{{cookiecutter.pulumi_org}}/{{cookiecutter.pulumi_stack}}' already exists. Continuing with existing stack."
        )
    subprocess.run(
        [
            "pulumi",
            "stack",
            "select",
            "{{cookiecutter.pulumi_org}}/{{cookiecutter.pulumi_stack}}",
        ],
        check=True,
    )
    print(
        f"Pulumi stack '{{cookiecutter.pulumi_org}}/{{cookiecutter.pulumi_stack}}' selected."
    )


if __name__ == "__main__":
    main()
